// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package apiv2

import (
	"context"
	"net/http"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/danielgtaylor/huma/v2"
)

type mailboxMessageBody struct {
	Body models.MailboxMessage
}

type mailboxMessagesBody struct {
	Body Paginated[*models.MailboxMessage]
}

type mailboxRecipientsBody struct {
	Body []*models.MailboxUser
}

type mailboxUnreadBody struct {
	Body struct {
		Count int64 `json:"count" doc:"Number of unread inbox messages."`
	}
}

type mailboxCreateBody struct {
	RecipientID          int64   `json:"recipient_id" minimum:"1" doc:"Recipient user id."`
	ReplyToID            int64   `json:"reply_to_id,omitempty" minimum:"0" doc:"Optional message id this message replies to."`
	Subject              string  `json:"subject" maxLength:"500" doc:"Message subject; an empty subject becomes '(No subject)'."`
	Body                 string  `json:"body" minLength:"1" maxLength:"50000" doc:"Message body."`
	ForwardAttachmentIDs []int64 `json:"forward_attachment_ids,omitempty" doc:"Existing mailbox attachment ids to include when forwarding."`
}

type mailboxReadBody struct {
	Read bool `json:"read" doc:"Whether the inbox message should be marked as read."`
}

func RegisterMailboxRoutes(api huma.API) {
	tags := []string{"mailbox"}

	Register(api, huma.Operation{
		OperationID: "mailbox-list",
		Summary:     "List mailbox messages",
		Description: "Lists the authenticated user's inbox or sent messages. Mailboxes are private and never available to public link shares.",
		Method:      http.MethodGet,
		Path:        "/mailbox/messages",
		Tags:        tags,
	}, mailboxList)

	Register(api, huma.Operation{
		OperationID: "mailbox-read",
		Summary:     "Read a mailbox message",
		Description: "Returns one private mailbox message if the authenticated user is its sender or recipient.",
		Method:      http.MethodGet,
		Path:        "/mailbox/messages/{message}",
		Tags:        tags,
	}, mailboxRead)

	Register(api, huma.Operation{
		OperationID:   "mailbox-send",
		Summary:       "Send a mailbox message",
		Description:   "Sends a private internal message to another active Vikunja user.",
		Method:        http.MethodPost,
		Path:          "/mailbox/messages",
		Tags:          tags,
		DefaultStatus: http.StatusCreated,
	}, mailboxSend)

	Register(api, huma.Operation{
		OperationID: "mailbox-mark-read",
		Summary:     "Mark a mailbox message read or unread",
		Description: "Changes the read state of a message in the authenticated recipient's inbox.",
		Method:      http.MethodPut,
		Path:        "/mailbox/messages/{message}/read",
		Tags:        tags,
	}, mailboxMarkRead)

	Register(api, huma.Operation{
		OperationID:   "mailbox-delete",
		Summary:       "Delete a mailbox message from the current user's folder",
		Description:   "Hides a message only from the authenticated user's inbox or sent folder. The other participant keeps their copy.",
		Method:        http.MethodDelete,
		Path:          "/mailbox/messages/{message}",
		Tags:          tags,
		DefaultStatus: http.StatusNoContent,
	}, mailboxDelete)

	Register(api, huma.Operation{
		OperationID: "mailbox-unread-count",
		Summary:     "Get unread mailbox count",
		Description: "Returns the authenticated user's unread inbox message count.",
		Method:      http.MethodGet,
		Path:        "/mailbox/unread-count",
		Tags:        tags,
	}, mailboxUnreadCount)

	Register(api, huma.Operation{
		OperationID: "mailbox-recipients",
		Summary:     "Search mailbox recipients",
		Description: "Searches active non-bot Vikunja users by name or username for composing internal mail. Email addresses are never returned.",
		Method:      http.MethodGet,
		Path:        "/mailbox/recipients",
		Tags:        tags,
	}, mailboxRecipients)
}

func init() { AddRouteRegistrar(RegisterMailboxRoutes) }

func mailboxList(ctx context.Context, in *struct {
	Folder     string `query:"folder" default:"inbox" enum:"inbox,sent" doc:"Mailbox folder to list."`
	Q          string `query:"q" doc:"Optional search across subject and body."`
	UnreadOnly bool   `query:"unread_only" default:"false" doc:"For inbox, return only unread messages."`
	Page       int    `query:"page" default:"1" minimum:"1" doc:"1-based page number."`
	PerPage    int    `query:"per_page" default:"30" minimum:"1" maximum:"100" doc:"Messages per page."`
}) (*mailboxMessagesBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	messages, total, err := models.ListMailboxMessages(s, a, in.Folder, in.Q, in.UnreadOnly, in.Page, in.PerPage)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &mailboxMessagesBody{Body: NewPaginated(messages, total, in.Page, in.PerPage)}, nil
}

func mailboxRead(ctx context.Context, in *struct {
	MessageID int64 `path:"message" minimum:"1"`
}) (*mailboxMessageBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	message, found, err := models.GetMailboxMessage(s, a, in.MessageID)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if !found {
		_ = s.Rollback()
		return nil, huma.Error404NotFound("mailbox message not found")
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &mailboxMessageBody{Body: *message}, nil
}

func mailboxSend(ctx context.Context, in *struct {
	Body mailboxCreateBody
}) (*mailboxMessageBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	message, err := models.CreateMailboxMessage(s, a, &models.MailboxMessage{
		RecipientID: in.Body.RecipientID,
		ReplyToID:   in.Body.ReplyToID,
		Subject:     in.Body.Subject,
		Body:        in.Body.Body,
	})
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := models.CopyMailboxAttachments(
		s,
		a,
		message.ID,
		in.Body.ForwardAttachmentIDs,
	); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	if len(in.Body.ForwardAttachmentIDs) > 0 {
		hydrated, found, err := models.GetMailboxMessage(
			s,
			a,
			message.ID,
		)
		if err != nil {
			_ = s.Rollback()
			return nil, translateDomainError(err)
		}
		if found {
			message = hydrated
		}
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &mailboxMessageBody{Body: *message}, nil
}

func mailboxMarkRead(ctx context.Context, in *struct {
	MessageID int64 `path:"message" minimum:"1"`
	Body      mailboxReadBody
}) (*mailboxMessageBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	message, found, err := models.SetMailboxMessageRead(s, a, in.MessageID, in.Body.Read)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if !found {
		_ = s.Rollback()
		return nil, huma.Error404NotFound("mailbox message not found")
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &mailboxMessageBody{Body: *message}, nil
}

func mailboxDelete(ctx context.Context, in *struct {
	MessageID int64 `path:"message" minimum:"1"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	found, err := models.DeleteMailboxMessage(s, a, in.MessageID)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if !found {
		_ = s.Rollback()
		return nil, huma.Error404NotFound("mailbox message not found")
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}

func mailboxUnreadCount(ctx context.Context, _ *struct{}) (*mailboxUnreadBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	count, err := models.CountUnreadMailboxMessages(s, a)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	response := &mailboxUnreadBody{}
	response.Body.Count = count
	return response, nil
}

func mailboxRecipients(ctx context.Context, in *struct {
	Q string `query:"q" doc:"Optional search by name or username."`
}) (*mailboxRecipientsBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	recipients, err := models.SearchMailboxRecipients(s, a, in.Q)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &mailboxRecipientsBody{Body: recipients}, nil
}
