// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package apiv2

import (
	"context"
	"net/http"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	webfiles "code.vikunja.io/api/pkg/web/files"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
)

type mailboxAttachmentUploadInput struct {
	MessageID int64 `path:"message" minimum:"1" doc:"Mailbox message id."`

	RawBody huma.MultipartFormFiles[struct {
		Files []huma.FormFile `form:"files" required:"true" doc:"One or more files."`
	}]
}

type mailboxAttachmentUploadError struct {
	Message string `json:"message"`
}

type mailboxAttachmentUploadResult struct {
	Success []*models.MailboxAttachment    `json:"success"`
	Errors  []mailboxAttachmentUploadError `json:"errors"`
}

type mailboxAttachmentUploadBody struct {
	Body mailboxAttachmentUploadResult
}

func RegisterMailboxAttachmentRoutes(api huma.API) {
	tags := []string{"mailbox"}

	Register(api, huma.Operation{
		OperationID: "mailbox-attachments-upload",
		Summary:     "Upload mailbox attachments",
		Description: "Adds one or more files to a message sent by the authenticated user.",
		Method:      http.MethodPost,
		Path:        "/mailbox/messages/{message}/attachments",
		Tags:        tags,
		MaxBodyBytes: (int64(config.GetMaxFileSizeInMBytes()) + 2) *
			1024 *
			1024,
	}, mailboxAttachmentsUpload)

	Register(api, huma.Operation{
		OperationID: "mailbox-attachments-download",
		Summary:     "Download mailbox attachment",
		Description: "Downloads an attachment if the authenticated user can read the message.",
		Method:      http.MethodGet,
		Path:        "/mailbox/messages/{message}/attachments/{attachment}",
		Tags:        tags,
		Responses: map[string]*huma.Response{
			"200": {
				Description: "Attachment bytes.",
				Content: map[string]*huma.MediaType{
					"application/octet-stream": {
						Schema: &huma.Schema{
							Type:   huma.TypeString,
							Format: "binary",
						},
					},
				},
			},
		},
	}, mailboxAttachmentsDownload)
}

func init() {
	AddRouteRegistrar(RegisterMailboxAttachmentRoutes)
}

func mailboxAttachmentsUpload(
	ctx context.Context,
	in *mailboxAttachmentUploadInput,
) (*mailboxAttachmentUploadBody, error) {
	a, err := authFromCtx(ctx)

	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	formFiles := in.RawBody.Data().Files

	uploads := make(
		[]*models.MailboxAttachmentToUpload,
		0,
		len(formFiles),
	)

	for _, file := range formFiles {
		uploads = append(
			uploads,
			&models.MailboxAttachmentToUpload{
				Reader:   file,
				Filename: file.Filename,
				Size:     uint64(file.Size),
			},
		)
	}

	success, failures, err :=
		models.UploadMailboxAttachments(
			s,
			a,
			in.MessageID,
			uploads,
		)

	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	result := mailboxAttachmentUploadResult{
		Success: success,
	}

	for _, failure := range failures {
		result.Errors = append(
			result.Errors,
			mailboxAttachmentUploadError{
				Message: failure.Error(),
			},
		)
	}

	return &mailboxAttachmentUploadBody{
		Body: result,
	}, nil
}

func mailboxAttachmentsDownload(
	ctx context.Context,
	in *struct {
		MessageID    int64 `path:"message" minimum:"1"`
		AttachmentID int64 `path:"attachment" minimum:"1"`
	},
) (*huma.StreamResponse, error) {
	a, err := authFromCtx(ctx)

	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	attachment, found, err :=
		models.LoadMailboxAttachmentForDownload(
			s,
			a,
			in.MessageID,
			in.AttachmentID,
		)

	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	if !found {
		_ = s.Rollback()
		return nil,
			huma.Error404NotFound(
				"mailbox attachment not found",
			)
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	return &huma.StreamResponse{
		Body: func(hctx huma.Context) {
			c := humaecho.Unwrap(hctx)

			defer func() {
				_ = attachment.File.File.Close()
			}()

			webfiles.WriteFileDownload(
				(*c).Response(),
				(*c).Request(),
				attachment.File,
			)
		},
	}, nil
}
