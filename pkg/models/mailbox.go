// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package models

import (
	"strings"
	"time"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/builder"
	"xorm.io/xorm"
)

const (
	MailboxFolderInbox = "inbox"
	MailboxFolderSent  = "sent"
)

// MailboxUser is the safe user shape exposed by the private internal mailbox.
type MailboxUser struct {
	ID       int64  `json:"id" readOnly:"true" doc:"The user's numeric id."`
	Name     string `json:"name" readOnly:"true" doc:"The user's display name."`
	Username string `json:"username" readOnly:"true" doc:"The user's username."`
}

// MailboxMessage is a private internal message shared only by its sender and recipient.
type MailboxMessage struct {
	ID          int64 `xorm:"bigint autoincr not null unique pk" json:"id" readOnly:"true" doc:"The message id."`
	SenderID    int64 `xorm:"bigint not null index" json:"sender_id" readOnly:"true" doc:"The sender user id."`
	RecipientID int64 `xorm:"bigint not null index" json:"recipient_id" doc:"The recipient user id."`
	ReplyToID   int64 `xorm:"bigint null index" json:"reply_to_id" doc:"Optional message id this message replies to."`

	Subject string `xorm:"varchar(500) not null" json:"subject" maxLength:"500" doc:"Message subject."`
	Body    string `xorm:"text not null" json:"body" maxLength:"50000" doc:"Message body."`

	ReadAt           time.Time `xorm:"datetime null index" json:"read_at" readOnly:"true" doc:"When the recipient marked the message as read; zero means unread."`
	SenderDeleted    bool      `xorm:"bool not null default false index" json:"-"`
	RecipientDeleted bool      `xorm:"bool not null default false index" json:"-"`

	Sender      *MailboxUser         `xorm:"-" json:"sender" readOnly:"true" doc:"Safe sender information."`
	Recipient   *MailboxUser         `xorm:"-" json:"recipient" readOnly:"true" doc:"Safe recipient information."`
	Attachments []*MailboxAttachment `xorm:"-" json:"attachments" readOnly:"true" doc:"Files attached to this message."`

	Created time.Time `xorm:"created not null index" json:"created" readOnly:"true" doc:"When the message was sent."`
	Updated time.Time `xorm:"updated not null" json:"updated" readOnly:"true" doc:"When the message was last changed."`
}

func (*MailboxMessage) TableName() string { return "user_mailbox_messages" }

func mailboxAuthUser(a web.Auth) (*user.User, bool) {
	u, ok := a.(*user.User)
	return u, ok && u != nil && u.ID > 0
}

func (m *MailboxMessage) loadForPermission(s *xorm.Session) error {
	if m.ID <= 0 {
		return nil
	}
	stored := &MailboxMessage{}
	has, err := s.ID(m.ID).Get(stored)
	if err != nil {
		return err
	}
	if !has {
		return nil
	}
	*m = *stored
	return nil
}

func (m *MailboxMessage) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	u, ok := mailboxAuthUser(a)
	if !ok {
		return false, 0, nil
	}
	if err := m.loadForPermission(s); err != nil {
		return false, 0, err
	}
	if m.SenderID == u.ID {
		return !m.SenderDeleted, 0, nil
	}
	if m.RecipientID == u.ID {
		return !m.RecipientDeleted, 0, nil
	}
	return false, 0, nil
}

func (m *MailboxMessage) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	u, ok := mailboxAuthUser(a)
	if !ok {
		return false, nil
	}
	return u.ID != m.RecipientID, nil
}

func (m *MailboxMessage) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	u, ok := mailboxAuthUser(a)
	if !ok {
		return false, nil
	}
	if err := m.loadForPermission(s); err != nil {
		return false, err
	}
	return m.RecipientID == u.ID && !m.RecipientDeleted, nil
}

func (m *MailboxMessage) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	u, ok := mailboxAuthUser(a)
	if !ok {
		return false, nil
	}
	if err := m.loadForPermission(s); err != nil {
		return false, err
	}
	if m.SenderID == u.ID {
		return !m.SenderDeleted, nil
	}
	if m.RecipientID == u.ID {
		return !m.RecipientDeleted, nil
	}
	return false, nil
}

func safeMailboxUser(u *user.User) *MailboxUser {
	if u == nil {
		return nil
	}
	return &MailboxUser{ID: u.ID, Name: u.Name, Username: u.Username}
}

func hydrateMailboxUsers(s *xorm.Session, messages []*MailboxMessage) error {
	ids := make([]int64, 0, len(messages)*2)
	seen := map[int64]struct{}{}
	for _, message := range messages {
		if _, ok := seen[message.SenderID]; !ok {
			ids = append(ids, message.SenderID)
			seen[message.SenderID] = struct{}{}
		}
		if _, ok := seen[message.RecipientID]; !ok {
			ids = append(ids, message.RecipientID)
			seen[message.RecipientID] = struct{}{}
		}
	}
	users, err := user.GetUsersByIDs(s, ids)
	if err != nil {
		return err
	}
	for _, message := range messages {
		message.Sender = safeMailboxUser(users[message.SenderID])
		message.Recipient = safeMailboxUser(users[message.RecipientID])
	}
	return nil
}

func validateMailboxMessage(message *MailboxMessage) error {
	message.Subject = strings.TrimSpace(message.Subject)
	message.Body = strings.TrimSpace(message.Body)
	if message.RecipientID <= 0 {
		return ErrInvalidData{Message: "mailbox recipient is required"}
	}
	if message.Subject == "" {
		message.Subject = "(No subject)"
	}
	if len([]rune(message.Subject)) > 500 {
		return ErrInvalidData{Message: "mailbox subject is too long"}
	}
	if message.Body == "" {
		return ErrInvalidData{Message: "mailbox message body is required"}
	}
	if len([]rune(message.Body)) > 50000 {
		return ErrInvalidData{Message: "mailbox message body is too long"}
	}
	return nil
}

func CreateMailboxMessage(s *xorm.Session, a web.Auth, input *MailboxMessage) (*MailboxMessage, error) {
	sender, ok := mailboxAuthUser(a)
	if !ok {
		return nil, ErrGenericForbidden{}
	}
	message := &MailboxMessage{
		SenderID:    sender.ID,
		RecipientID: input.RecipientID,
		ReplyToID:   input.ReplyToID,
		Subject:     input.Subject,
		Body:        input.Body,
	}
	if err := validateMailboxMessage(message); err != nil {
		return nil, err
	}
	can, err := message.CanCreate(s, a)
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, ErrGenericForbidden{}
	}

	recipient, err := user.GetUserByID(s, message.RecipientID)
	if err != nil {
		return nil, err
	}
	if recipient.IsBot() {
		return nil, ErrInvalidData{Message: "messages cannot be sent to bot users"}
	}
	if message.ReplyToID > 0 {
		parent := &MailboxMessage{ID: message.ReplyToID}
		canRead, _, err := parent.CanRead(s, a)
		if err != nil {
			return nil, err
		}
		if !canRead {
			return nil, ErrGenericForbidden{}
		}
		if parent.SenderID != message.RecipientID && parent.RecipientID != message.RecipientID {
			return nil, ErrInvalidData{Message: "reply recipient is not part of the original message"}
		}
	}
	if _, err := s.Insert(message); err != nil {
		return nil, err
	}
	message.Sender = safeMailboxUser(sender)
	message.Recipient = safeMailboxUser(recipient)

	if err := notifyMailboxMessageReceived(
		s,
		sender,
		recipient,
		message,
	); err != nil {
		return nil, err
	}

	return message, nil
}

func GetMailboxMessage(s *xorm.Session, a web.Auth, messageID int64) (*MailboxMessage, bool, error) {
	message := &MailboxMessage{ID: messageID}
	can, _, err := message.CanRead(s, a)
	if err != nil {
		return nil, false, err
	}
	if !can {
		return nil, false, nil
	}
	if err := hydrateMailboxUsers(s, []*MailboxMessage{message}); err != nil {
		return nil, false, err
	}
	if err := hydrateMailboxAttachments(s, []*MailboxMessage{message}); err != nil {
		return nil, false, err
	}
	return message, true, nil
}

func mailboxListSession(s *xorm.Session, userID int64, folder, search string, unreadOnly bool) (*xorm.Session, error) {
	query := s.Table((&MailboxMessage{}).TableName())
	switch folder {
	case MailboxFolderInbox:
		query = query.Where("recipient_id = ?", userID).And("recipient_deleted = ?", false)
		if unreadOnly {
			query = query.And("read_at IS NULL")
		}
	case MailboxFolderSent:
		query = query.Where("sender_id = ?", userID).And("sender_deleted = ?", false)
	default:
		return nil, ErrInvalidData{Message: "invalid mailbox folder"}
	}
	search = strings.TrimSpace(search)
	if search != "" {
		pattern := "%" + search + "%"
		query = query.And(builder.Or(builder.Like{"subject", pattern}, builder.Like{"body", pattern}))
	}
	return query, nil
}

func ListMailboxMessages(s *xorm.Session, a web.Auth, folder, search string, unreadOnly bool, page, perPage int) ([]*MailboxMessage, int64, error) {
	u, ok := mailboxAuthUser(a)
	if !ok {
		return nil, 0, ErrGenericForbidden{}
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 30
	}
	countQuery, err := mailboxListSession(s, u.ID, folder, search, unreadOnly)
	if err != nil {
		return nil, 0, err
	}
	total, err := countQuery.Count(&MailboxMessage{})
	if err != nil {
		return nil, 0, err
	}
	listQuery, err := mailboxListSession(s, u.ID, folder, search, unreadOnly)
	if err != nil {
		return nil, 0, err
	}
	messages := []*MailboxMessage{}
	if err := listQuery.OrderBy("created desc, id desc").Limit(perPage, (page-1)*perPage).Find(&messages); err != nil {
		return nil, 0, err
	}
	if err := hydrateMailboxUsers(s, messages); err != nil {
		return nil, 0, err
	}
	if err := hydrateMailboxAttachments(s, messages); err != nil {
		return nil, 0, err
	}
	return messages, total, nil
}

func SetMailboxMessageRead(s *xorm.Session, a web.Auth, messageID int64, read bool) (*MailboxMessage, bool, error) {
	message := &MailboxMessage{ID: messageID}
	can, err := message.CanUpdate(s, a)
	if err != nil {
		return nil, false, err
	}
	if !can {
		return nil, false, nil
	}
	if read {
		message.ReadAt = time.Now()
	} else {
		message.ReadAt = time.Time{}
	}
	message.Updated = time.Now()
	if _, err := s.ID(message.ID).Cols("read_at", "updated").Update(message); err != nil {
		return nil, true, err
	}

	if err := syncMailboxMessageNotificationRead(
		s,
		message.RecipientID,
		message.ID,
		read,
	); err != nil {
		return nil, true, err
	}

	if err := hydrateMailboxUsers(s, []*MailboxMessage{message}); err != nil {
		return nil, true, err
	}
	if err := hydrateMailboxAttachments(s, []*MailboxMessage{message}); err != nil {
		return nil, true, err
	}
	return message, true, nil
}

func DeleteMailboxMessage(s *xorm.Session, a web.Auth, messageID int64) (bool, error) {
	u, ok := mailboxAuthUser(a)
	if !ok {
		return false, ErrGenericForbidden{}
	}
	message := &MailboxMessage{ID: messageID}
	can, err := message.CanDelete(s, a)
	if err != nil {
		return false, err
	}
	if !can {
		return false, nil
	}
	message.Updated = time.Now()
	if message.SenderID == u.ID {
		message.SenderDeleted = true
		if _, err := s.ID(message.ID).Cols("sender_deleted", "updated").Update(message); err != nil {
			return true, err
		}
	} else {
		message.RecipientDeleted = true
		if _, err := s.ID(message.ID).Cols("recipient_deleted", "updated").Update(message); err != nil {
			return true, err
		}
	}
	if message.SenderDeleted && message.RecipientDeleted {
		if err := DeleteMailboxAttachmentsForMessage(s, message.ID); err != nil {
			return true, err
		}
		if _, err := s.ID(message.ID).Delete(&MailboxMessage{}); err != nil {
			return true, err
		}
	}
	return true, nil
}

func CountUnreadMailboxMessages(s *xorm.Session, a web.Auth) (int64, error) {
	u, ok := mailboxAuthUser(a)
	if !ok {
		return 0, ErrGenericForbidden{}
	}
	return s.Where("recipient_id = ?", u.ID).
		And("recipient_deleted = ?", false).
		And("read_at IS NULL").
		Count(&MailboxMessage{})
}

func SearchMailboxRecipients(s *xorm.Session, a web.Auth, search string) ([]*MailboxUser, error) {
	current, ok := mailboxAuthUser(a)
	if !ok {
		return nil, ErrGenericForbidden{}
	}
	query := s.Where("status = ?", user.StatusActive).
		And(builder.IsNull{"deletion_scheduled_at"}).
		And(builder.Or(builder.IsNull{"bot_owner_id"}, builder.Eq{"bot_owner_id": 0})).
		And("id != ?", current.ID)
	search = strings.TrimSpace(search)
	if search != "" {
		pattern := "%" + search + "%"
		query = query.And(builder.Or(builder.Like{"username", pattern}, builder.Like{"name", pattern}))
	}
	users := []*user.User{}
	if err := query.OrderBy("name asc, username asc").Limit(30).Find(&users); err != nil {
		return nil, err
	}
	result := make([]*MailboxUser, 0, len(users))
	for _, u := range users {
		result = append(result, safeMailboxUser(u))
	}
	return result, nil
}
