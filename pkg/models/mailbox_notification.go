// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package models

import (
	"strings"

	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"

	"xorm.io/xorm"
)

const mailboxMessageReceivedNotificationName = "mailbox.message.received"

// MailboxMessageReceivedNotification is an in-app notification created
// for the recipient of every new internal mailbox message.
type MailboxMessageReceivedNotification struct {
	Doer      *user.User `json:"doer"`
	MessageID int64      `json:"message_id"`
	Subject   string     `json:"subject"`
	Preview   string     `json:"preview"`
}

func init() {
	notifications.Register(func() notifications.PersistedNotification {
		return &MailboxMessageReceivedNotification{}
	})
}

func (n *MailboxMessageReceivedNotification) Name() string {
	return mailboxMessageReceivedNotificationName
}

func (n *MailboxMessageReceivedNotification) SubjectID() int64 {
	return n.MessageID
}

func (n *MailboxMessageReceivedNotification) ProjectID() int64 {
	return 0
}

// Internal mailbox messages create only an in-app notification.
// They must not create a second external email.
func (n *MailboxMessageReceivedNotification) ToMail(_ string) *notifications.Mail {
	return nil
}

func (n *MailboxMessageReceivedNotification) ToDB() interface{} {
	return n
}

func mailboxNotificationPreview(body string) string {
	clean := strings.Join(strings.Fields(body), " ")
	runes := []rune(clean)

	const maxRunes = 140
	if len(runes) <= maxRunes {
		return clean
	}

	return string(runes[:maxRunes]) + "…"
}

func notifyMailboxMessageReceived(
	s *xorm.Session,
	sender *user.User,
	recipient *user.User,
	message *MailboxMessage,
) error {
	return notifications.Notify(
		recipient,
		&MailboxMessageReceivedNotification{
			Doer:      sender,
			MessageID: message.ID,
			Subject:   message.Subject,
			Preview:   mailboxNotificationPreview(message.Body),
		},
		s,
	)
}

// Keep the notification read state synchronized with the mailbox message.
func syncMailboxMessageNotificationRead(
	s *xorm.Session,
	userID int64,
	messageID int64,
	read bool,
) error {
	items, err := notifications.GetNotificationsForNameAndUser(
		s,
		userID,
		mailboxMessageReceivedNotificationName,
		messageID,
	)
	if err != nil {
		return err
	}

	for _, item := range items {
		if err := notifications.MarkNotificationAsRead(
			s,
			item,
			read,
		); err != nil {
			return err
		}
	}

	return nil
}
