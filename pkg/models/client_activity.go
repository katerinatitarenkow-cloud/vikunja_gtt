// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package models

import (
	"encoding/json"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

const (
	ClientActivityCall               = "call"
	ClientActivityMessage            = "message"
	ClientActivityMeeting            = "meeting"
	ClientActivityManualNote         = "manual_note"
	ClientActivityDocumentSent       = "document_sent"
	ClientActivityProposalSent       = "commercial_proposal_sent"
	ClientActivityInvoiceSent        = "invoice_sent"
	ClientActivityTaskCreated        = "task_created"
	ClientActivityTaskCompleted      = "task_completed"
	ClientActivityTaskReopened       = "task_reopened"
	ClientActivityCommentCreated     = "comment_created"
	ClientActivityStatusChanged      = "status_changed"
	ClientActivityResponsibleChanged = "responsible_changed"
	ClientActivityProposalUploaded   = "commercial_proposal_uploaded"
	ClientActivityProposalReplaced   = "commercial_proposal_replaced"
	ClientActivityProposalDeleted    = "commercial_proposal_deleted"
	ClientActivityCustomFieldCreated = "custom_field_created"
	ClientActivityCustomFieldUpdated = "custom_field_updated"
	ClientActivityCustomFieldDeleted = "custom_field_deleted"
)

var clientManualActivityTypes = map[string]struct{}{
	ClientActivityCall:         {},
	ClientActivityMessage:      {},
	ClientActivityMeeting:      {},
	ClientActivityManualNote:   {},
	ClientActivityDocumentSent: {},
	ClientActivityProposalSent: {},
	ClientActivityInvoiceSent:  {},
}

// ClientActivityMetadata contains optional structured details shared by manual
// and automatic CRM events. New fields can be added without a database migration
// because the struct is serialized into MetadataJSON.
type ClientActivityMetadata struct {
	Direction         string `json:"direction,omitempty" doc:"Optional incoming/outgoing direction for calls and messages."`
	Channel           string `json:"channel,omitempty" doc:"Optional communication channel, for example Telegram, WhatsApp or email."`
	DurationMinutes   int    `json:"duration_minutes,omitempty" doc:"Optional duration in minutes for calls and meetings."`
	Result            string `json:"result,omitempty" doc:"Optional outcome or result of an activity."`
	ContactPersonID   int64  `json:"contact_person_id,omitempty" doc:"Optional CRM contact-person id associated with the activity."`
	ContactPersonName string `json:"contact_person_name,omitempty" doc:"Snapshot of the contact-person name at the time of the event."`
	TaskTitle         string `json:"task_title,omitempty" doc:"Snapshot of the related task title."`
	OldValue          string `json:"old_value,omitempty" doc:"Previous value for a change event."`
	NewValue          string `json:"new_value,omitempty" doc:"New value for a change event."`
	FileName          string `json:"file_name,omitempty" doc:"Snapshot of a related file name."`
	FieldName         string `json:"field_name,omitempty" doc:"Current custom-field name for CRM custom field events."`
	OldFieldName      string `json:"old_field_name,omitempty" doc:"Previous custom-field name when it changed."`
	NewFieldName      string `json:"new_field_name,omitempty" doc:"New custom-field name when it changed."`
}

// ClientActivityEvent is one item in the chronological CRM history for a
// project/client. ProjectID is intentionally used while Project == Client in the
// current architecture; EntityType/EntityID keep the event relation generic.
type ClientActivityEvent struct {
	ID              int64                   `xorm:"bigint autoincr not null unique pk" json:"id" readOnly:"true" doc:"Unique numeric event id."`
	ProjectID       int64                   `xorm:"bigint not null index" json:"project_id" readOnly:"true" doc:"Project/client this history event belongs to."`
	EventType       string                  `xorm:"varchar(80) not null index" json:"event_type" readOnly:"true" doc:"Stable machine-readable CRM event type."`
	ActorUserID     int64                   `xorm:"bigint null index" json:"actor_user_id" readOnly:"true" doc:"User who caused the event, or zero for a system actor."`
	Actor           *user.User              `xorm:"-" json:"actor,omitempty" readOnly:"true" doc:"User who caused the event when available."`
	OccurredAt      time.Time               `xorm:"DATETIME not null index" json:"occurred_at" readOnly:"true" doc:"When the customer interaction or system event occurred."`
	Title           string                  `xorm:"varchar(1000) null" json:"title" readOnly:"true" doc:"Optional user-entered title for manual activities."`
	Description     string                  `xorm:"text null" json:"description" readOnly:"true" doc:"Activity notes or a snapshot of relevant text such as a task comment."`
	EntityType      string                  `xorm:"varchar(80) null index" json:"entity_type" readOnly:"true" doc:"Type of linked entity, such as task or commercial_proposal."`
	EntityID        int64                   `xorm:"bigint null index" json:"entity_id" readOnly:"true" doc:"Identifier of the linked entity when one exists."`
	MetadataJSON    string                  `xorm:"text null" json:"-"`
	Metadata        *ClientActivityMetadata `xorm:"-" json:"metadata,omitempty" readOnly:"true" doc:"Structured event-specific details."`
	SystemGenerated bool                    `xorm:"bool not null default false index" json:"system_generated" readOnly:"true" doc:"Whether Vikunja generated this event automatically."`
	Created         time.Time               `xorm:"created not null" json:"created" readOnly:"true" doc:"When this history row was stored."`
}

func (*ClientActivityEvent) TableName() string { return "client_activity_events" }

// ClientActivityCreate is the writable shape for a manually recorded CRM activity.
type ClientActivityCreate struct {
	EventType   string                 `json:"event_type" enum:"call,message,meeting,manual_note,document_sent,commercial_proposal_sent,invoice_sent" doc:"Manual CRM activity type."`
	OccurredAt  time.Time              `json:"occurred_at" doc:"When the interaction happened. Zero means now."`
	Title       string                 `json:"title" maxLength:"1000" doc:"Short activity subject or title."`
	Description string                 `json:"description" doc:"Free-form notes about the interaction."`
	EntityType  string                 `json:"entity_type,omitempty" maxLength:"80" doc:"Optional linked entity type for future CRM modules."`
	EntityID    int64                  `json:"entity_id,omitempty" doc:"Optional linked entity id."`
	Metadata    ClientActivityMetadata `json:"metadata" doc:"Structured call, meeting, message or document details."`
}

func isManualClientActivityType(eventType string) bool {
	_, ok := clientManualActivityTypes[eventType]
	return ok
}

func prepareClientActivityMetadata(event *ClientActivityEvent) error {
	if event.Metadata == nil {
		event.MetadataJSON = ""
		return nil
	}
	encoded, err := json.Marshal(event.Metadata)
	if err != nil {
		return err
	}
	event.MetadataJSON = string(encoded)
	return nil
}

func hydrateClientActivityMetadata(event *ClientActivityEvent) error {
	if strings.TrimSpace(event.MetadataJSON) == "" {
		event.Metadata = nil
		return nil
	}
	metadata := &ClientActivityMetadata{}
	if err := json.Unmarshal([]byte(event.MetadataJSON), metadata); err != nil {
		return err
	}
	event.Metadata = metadata
	return nil
}

func hydrateClientActivityActors(s *xorm.Session, events []*ClientActivityEvent) error {
	ids := make([]int64, 0, len(events))
	for _, event := range events {
		if event.ActorUserID > 0 {
			ids = append(ids, event.ActorUserID)
		}
	}
	users, err := user.GetUsersByIDs(s, ids)
	if err != nil {
		return err
	}
	for _, event := range events {
		if actor, ok := users[event.ActorUserID]; ok {
			event.Actor = actor
		}
		if err := hydrateClientActivityMetadata(event); err != nil {
			return err
		}
	}
	return nil
}

// CreateClientActivityEvent stores a system/manual event inside the caller's
// transaction. This is used by tasks, comments and CRM profile operations so the
// history cannot commit separately from the action it describes.
func CreateClientActivityEvent(s *xorm.Session, event *ClientActivityEvent) error {
	if event.ProjectID <= 0 || strings.TrimSpace(event.EventType) == "" {
		return ErrInvalidData{Message: "client activity requires project and event type"}
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now()
	}
	if event.Actor != nil {
		event.ActorUserID = event.Actor.ID
	}
	if err := prepareClientActivityMetadata(event); err != nil {
		return err
	}
	_, err := s.Insert(event)
	return err
}

// CreateSystemClientActivity is the concise helper used by existing domain
// operations to append automatic CRM history entries transactionally.
func CreateSystemClientActivity(s *xorm.Session, projectID int64, eventType string, doer *user.User, entityType string, entityID int64, description string, metadata *ClientActivityMetadata) error {
	return CreateClientActivityEvent(s, &ClientActivityEvent{
		ProjectID:       projectID,
		EventType:       eventType,
		Actor:           doer,
		OccurredAt:      time.Now(),
		Description:     description,
		EntityType:      entityType,
		EntityID:        entityID,
		Metadata:        metadata,
		SystemGenerated: true,
	})
}

// CreateManualClientActivity validates and stores a user-entered CRM interaction.
func CreateManualClientActivity(s *xorm.Session, a web.Auth, projectID int64, input *ClientActivityCreate) (*ClientActivityEvent, error) {
	input.EventType = strings.TrimSpace(input.EventType)
	if !isManualClientActivityType(input.EventType) {
		return nil, ErrInvalidData{Message: "invalid manual client activity type"}
	}
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.EntityType = strings.TrimSpace(input.EntityType)
	input.Metadata.Direction = strings.TrimSpace(input.Metadata.Direction)
	input.Metadata.Channel = strings.TrimSpace(input.Metadata.Channel)
	input.Metadata.Result = strings.TrimSpace(input.Metadata.Result)
	input.Metadata.ContactPersonName = strings.TrimSpace(input.Metadata.ContactPersonName)
	if input.Metadata.DurationMinutes < 0 {
		return nil, ErrInvalidData{Message: "activity duration cannot be negative"}
	}
	if input.Metadata.ContactPersonID > 0 {
		contact := &ClientContactPerson{ID: input.Metadata.ContactPersonID, ProjectID: projectID}
		has, err := s.Where("id = ? AND project_id = ?", contact.ID, projectID).Get(contact)
		if err != nil {
			return nil, err
		}
		if !has {
			return nil, ErrInvalidData{Message: "contact person does not belong to this client"}
		}
		input.Metadata.ContactPersonName = contact.FullName
	}

	actor := doerFromAuth(s, a)
	event := &ClientActivityEvent{
		ProjectID:       projectID,
		EventType:       input.EventType,
		Actor:           actor,
		OccurredAt:      input.OccurredAt,
		Title:           input.Title,
		Description:     input.Description,
		EntityType:      input.EntityType,
		EntityID:        input.EntityID,
		Metadata:        &input.Metadata,
		SystemGenerated: false,
	}
	if err := CreateClientActivityEvent(s, event); err != nil {
		return nil, err
	}
	return event, nil
}

// GetClientActivityEvents returns newest events first and optionally filters by
// one stable event type. Pagination keeps long-lived customer histories bounded.
func GetClientActivityEvents(s *xorm.Session, projectID int64, eventType string, page, perPage int) ([]*ClientActivityEvent, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 200 {
		perPage = 50
	}
	trimmedType := strings.TrimSpace(eventType)
	countQuery := s.Where("project_id = ?", projectID)
	if trimmedType != "" {
		countQuery = countQuery.And("event_type = ?", trimmedType)
	}
	total, err := countQuery.Count(&ClientActivityEvent{})
	if err != nil {
		return nil, 0, err
	}

	findQuery := s.Where("project_id = ?", projectID)
	if trimmedType != "" {
		findQuery = findQuery.And("event_type = ?", trimmedType)
	}
	events := []*ClientActivityEvent{}
	if err := findQuery.OrderBy("occurred_at desc, id desc").Limit(perPage, (page-1)*perPage).Find(&events); err != nil {
		return nil, 0, err
	}
	if err := hydrateClientActivityActors(s, events); err != nil {
		return nil, 0, err
	}
	return events, total, nil
}

// DeleteManualClientActivity removes only user-entered history. Automatic audit
// facts are immutable from the CRM UI.
func DeleteManualClientActivity(s *xorm.Session, projectID, eventID int64) (bool, error) {
	event := &ClientActivityEvent{}
	has, err := s.Where("id = ? AND project_id = ?", eventID, projectID).Get(event)
	if err != nil || !has {
		return has, err
	}
	if event.SystemGenerated {
		return false, ErrInvalidData{Message: "automatic client history events cannot be deleted"}
	}
	affected, err := s.Where("id = ? AND project_id = ? AND system_generated = ?", eventID, projectID, false).Delete(&ClientActivityEvent{})
	return affected > 0, err
}
