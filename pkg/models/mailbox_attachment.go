// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package models

import (
	"io"
	"time"

	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// MailboxAttachment links a mailbox message to a file in Vikunja's normal file storage.
// Forwarded messages may share a FileID; the physical file is removed only after the
// final mailbox attachment referencing it is deleted.
type MailboxAttachment struct {
	ID        int64 `xorm:"bigint autoincr not null unique pk" json:"id" readOnly:"true"`
	MessageID int64 `xorm:"bigint not null index" json:"message_id" readOnly:"true"`
	FileID    int64 `xorm:"bigint not null index" json:"-"`

	File *files.File `xorm:"-" json:"file" readOnly:"true"`

	Created time.Time `xorm:"created not null index" json:"created" readOnly:"true"`
}

func (*MailboxAttachment) TableName() string {
	return "user_mailbox_attachments"
}

type MailboxAttachmentToUpload struct {
	Reader   io.ReadSeeker
	Filename string
	Size     uint64
}

func hydrateMailboxAttachments(
	s *xorm.Session,
	messages []*MailboxMessage,
) error {
	if len(messages) == 0 {
		return nil
	}

	messageIDs := make([]int64, 0, len(messages))

	for _, message := range messages {
		message.Attachments = []*MailboxAttachment{}
		messageIDs = append(messageIDs, message.ID)
	}

	attachments := []*MailboxAttachment{}

	if err := s.
		In("message_id", messageIDs).
		Asc("id").
		Find(&attachments); err != nil {
		return err
	}

	if len(attachments) == 0 {
		return nil
	}

	fileIDs := make([]int64, 0, len(attachments))

	for _, attachment := range attachments {
		fileIDs = append(fileIDs, attachment.FileID)
	}

	fileMap := make(map[int64]*files.File)

	if err := s.
		In("id", fileIDs).
		Find(&fileMap); err != nil {
		return err
	}

	byMessage := make(map[int64][]*MailboxAttachment)

	for _, attachment := range attachments {
		attachment.File = fileMap[attachment.FileID]
		byMessage[attachment.MessageID] = append(
			byMessage[attachment.MessageID],
			attachment,
		)
	}

	for _, message := range messages {
		message.Attachments = byMessage[message.ID]

		if message.Attachments == nil {
			message.Attachments = []*MailboxAttachment{}
		}
	}

	return nil
}

func UploadMailboxAttachments(
	s *xorm.Session,
	a web.Auth,
	messageID int64,
	uploads []*MailboxAttachmentToUpload,
) (
	success []*MailboxAttachment,
	failures []error,
	err error,
) {
	current, ok := mailboxAuthUser(a)

	if !ok {
		return nil, nil, ErrGenericForbidden{}
	}

	message := &MailboxMessage{ID: messageID}

	if err := message.loadForPermission(s); err != nil {
		return nil, nil, err
	}

	// Attachments may only be added by the sender.
	if message.SenderID != current.ID || message.SenderDeleted {
		return nil, nil, ErrGenericForbidden{}
	}

	for _, upload := range uploads {
		file, fileErr := files.CreateWithSession(
			s,
			upload.Reader,
			upload.Filename,
			upload.Size,
			a,
		)

		if fileErr != nil {
			failures = append(failures, fileErr)
			continue
		}

		attachment := &MailboxAttachment{
			MessageID: messageID,
			FileID:    file.ID,
			File:      file,
		}

		if _, insertErr := s.Insert(attachment); insertErr != nil {
			_ = file.Delete(s)
			failures = append(failures, insertErr)
			continue
		}

		success = append(success, attachment)
	}

	return success, failures, nil
}

// CopyMailboxAttachments adds references to already existing mailbox files.
// This is used by forwarding: no browser download/re-upload is needed.
func CopyMailboxAttachments(
	s *xorm.Session,
	a web.Auth,
	targetMessageID int64,
	attachmentIDs []int64,
) error {
	if len(attachmentIDs) == 0 {
		return nil
	}

	current, ok := mailboxAuthUser(a)

	if !ok {
		return ErrGenericForbidden{}
	}

	target := &MailboxMessage{ID: targetMessageID}

	if err := target.loadForPermission(s); err != nil {
		return err
	}

	if target.SenderID != current.ID || target.SenderDeleted {
		return ErrGenericForbidden{}
	}

	seen := make(map[int64]struct{}, len(attachmentIDs))

	for _, attachmentID := range attachmentIDs {
		if attachmentID <= 0 {
			continue
		}

		if _, duplicate := seen[attachmentID]; duplicate {
			continue
		}

		seen[attachmentID] = struct{}{}

		source := &MailboxAttachment{}

		has, err := s.ID(attachmentID).Get(source)

		if err != nil {
			return err
		}

		if !has {
			return ErrInvalidData{
				Message: "mailbox attachment does not exist",
			}
		}

		sourceMessage := &MailboxMessage{
			ID: source.MessageID,
		}

		canRead, _, err := sourceMessage.CanRead(s, a)

		if err != nil {
			return err
		}

		if !canRead {
			return ErrGenericForbidden{}
		}

		copyAttachment := &MailboxAttachment{
			MessageID: targetMessageID,
			FileID:    source.FileID,
		}

		if _, err := s.Insert(copyAttachment); err != nil {
			return err
		}
	}

	return nil
}

func LoadMailboxAttachmentForDownload(
	s *xorm.Session,
	a web.Auth,
	messageID int64,
	attachmentID int64,
) (*MailboxAttachment, bool, error) {
	message := &MailboxMessage{
		ID: messageID,
	}

	canRead, _, err := message.CanRead(s, a)

	if err != nil {
		return nil, false, err
	}

	if !canRead {
		return nil, false, nil
	}

	attachment := &MailboxAttachment{}

	has, err := s.
		Where(
			"id = ? AND message_id = ?",
			attachmentID,
			messageID,
		).
		Get(attachment)

	if err != nil {
		return nil, false, err
	}

	if !has {
		return nil, false, nil
	}

	file := &files.File{}

	has, err = s.ID(attachment.FileID).Get(file)

	if err != nil {
		return nil, false, err
	}

	if !has {
		return nil, false, nil
	}

	if err := file.LoadFileByID(); err != nil {
		return nil, false, err
	}

	attachment.File = file

	return attachment, true, nil
}

// DeleteMailboxAttachmentsForMessage removes attachment links for a message.
// A physical file is deleted only when no forwarded/copy attachment still references it.
func DeleteMailboxAttachmentsForMessage(
	s *xorm.Session,
	messageID int64,
) error {
	attachments := []*MailboxAttachment{}

	if err := s.
		Where("message_id = ?", messageID).
		Find(&attachments); err != nil {
		return err
	}

	if len(attachments) == 0 {
		return nil
	}

	if _, err := s.
		Where("message_id = ?", messageID).
		Delete(&MailboxAttachment{}); err != nil {
		return err
	}

	seenFiles := map[int64]struct{}{}

	for _, attachment := range attachments {
		if _, done := seenFiles[attachment.FileID]; done {
			continue
		}

		seenFiles[attachment.FileID] = struct{}{}

		remaining, err := s.
			Where("file_id = ?", attachment.FileID).
			Count(&MailboxAttachment{})

		if err != nil {
			return err
		}

		if remaining > 0 {
			continue
		}

		file := &files.File{
			ID: attachment.FileID,
		}

		if err := file.Delete(s); err != nil &&
			!files.IsErrFileDoesNotExist(err) {
			return err
		}
	}

	return nil
}
