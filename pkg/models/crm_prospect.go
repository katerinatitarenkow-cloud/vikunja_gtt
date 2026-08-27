package models

import (
	"strings"
	"time"

	"code.vikunja.io/api/pkg/user"
)

const (
	ProspectStatusNew           = "new"
	ProspectStatusNoAnswer      = "no_answer"
	ProspectStatusInProgress    = "in_progress"
	ProspectStatusCallback      = "callback"
	ProspectStatusInterested    = "interested"
	ProspectStatusNotInterested = "not_interested"
	ProspectStatusInvalid       = "invalid"
	ProspectStatusArchived      = "archived"
)

const (
	ProspectCallOutcomeNoAnswer      = "no_answer"
	ProspectCallOutcomeBusy          = "busy"
	ProspectCallOutcomeConversation  = "conversation"
	ProspectCallOutcomeCallback      = "callback"
	ProspectCallOutcomeInterested    = "interested"
	ProspectCallOutcomeNotInterested = "not_interested"
	ProspectCallOutcomeWrongNumber   = "wrong_number"
)

var validProspectStatuses = map[string]struct{}{
	ProspectStatusNew:           {},
	ProspectStatusNoAnswer:      {},
	ProspectStatusInProgress:    {},
	ProspectStatusCallback:      {},
	ProspectStatusInterested:    {},
	ProspectStatusNotInterested: {},
	ProspectStatusInvalid:       {},
	ProspectStatusArchived:      {},
}

var validProspectCallOutcomes = map[string]struct{}{
	ProspectCallOutcomeNoAnswer:      {},
	ProspectCallOutcomeBusy:          {},
	ProspectCallOutcomeConversation:  {},
	ProspectCallOutcomeCallback:      {},
	ProspectCallOutcomeInterested:    {},
	ProspectCallOutcomeNotInterested: {},
	ProspectCallOutcomeWrongNumber:   {},
}

type ProspectCampaign struct {
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id"`

	Name        string `xorm:"varchar(500) not null index" json:"name"`
	Description string `xorm:"text null" json:"description"`

	CreatedByUserID int64 `xorm:"bigint not null index" json:"created_by_user_id"`
	Archived        bool  `xorm:"bool not null default false index" json:"archived"`

	Created time.Time `xorm:"created not null" json:"created"`
	Updated time.Time `xorm:"updated not null" json:"updated"`
}

func (*ProspectCampaign) TableName() string {
	return "crm_prospect_campaigns"
}

type Prospect struct {
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id"`

	CampaignID int64 `xorm:"bigint null index" json:"campaign_id"`

	CompanyName string `xorm:"varchar(500) null index" json:"company_name"`
	ContactName string `xorm:"varchar(500) null" json:"contact_name"`

	Phone           string `xorm:"varchar(100) null index" json:"phone"`
	PhoneNormalized string `xorm:"varchar(50) null index" json:"phone_normalized"`
	PhoneSecondary  string `xorm:"varchar(100) null" json:"phone_secondary"`

	Email   string `xorm:"varchar(320) null index" json:"email"`
	Website string `xorm:"varchar(1000) null" json:"website"`

	Region string `xorm:"varchar(250) null index" json:"region"`
	City   string `xorm:"varchar(250) null index" json:"city"`

	Notes string `xorm:"text null" json:"notes"`

	Status string `xorm:"varchar(40) not null default 'new' index" json:"status"`

	ResponsibleUserID int64      `xorm:"bigint null index" json:"responsible_user_id"`
	Responsible       *user.User `xorm:"-" json:"responsible,omitempty"`

	LastContactAt *time.Time `xorm:"DATETIME null index" json:"last_contact_at,omitempty"`
	NextContactAt *time.Time `xorm:"DATETIME null index" json:"next_contact_at,omitempty"`

	ProjectID int64 `xorm:"bigint null index" json:"project_id"`

	ImportRowNumber int    `xorm:"int null" json:"import_row_number,omitempty"`
	RawDataJSON     string `xorm:"text null" json:"-"`

	Created time.Time `xorm:"created not null index" json:"created"`
	Updated time.Time `xorm:"updated not null" json:"updated"`
}

func (*Prospect) TableName() string {
	return "crm_prospects"
}

type ProspectCall struct {
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id"`

	ProspectID int64 `xorm:"bigint not null index" json:"prospect_id"`
	UserID     int64 `xorm:"bigint not null index" json:"user_id"`

	OccurredAt time.Time `xorm:"DATETIME not null index" json:"occurred_at"`

	Outcome string `xorm:"varchar(40) not null index" json:"outcome"`

	DurationMinutes int    `xorm:"int not null default 0" json:"duration_minutes"`
	Note            string `xorm:"text null" json:"note"`

	NextContactAt *time.Time `xorm:"DATETIME null index" json:"next_contact_at,omitempty"`

	Created time.Time `xorm:"created not null" json:"created"`
}

func (*ProspectCall) TableName() string {
	return "crm_prospect_calls"
}

func validateProspectStatus(status string) error {
	status = strings.TrimSpace(status)

	if _, ok := validProspectStatuses[status]; !ok {
		return ErrInvalidData{
			Message: "invalid prospect status",
		}
	}

	return nil
}

func validateProspectCallOutcome(outcome string) error {
	outcome = strings.TrimSpace(outcome)

	if _, ok := validProspectCallOutcomes[outcome]; !ok {
		return ErrInvalidData{
			Message: "invalid prospect call outcome",
		}
	}

	return nil
}
