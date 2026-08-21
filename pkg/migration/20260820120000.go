// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package migration

import (
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

// Patch 11.3 finishes the CRM project card with a commercial-proposal file
// reference and an arbitrary number of structured contact persons.
type clientProfile20260820120000 struct {
	ProjectID int64 `xorm:"bigint not null unique pk"`

	ClientType  string `xorm:"varchar(20) not null default 'company'"`
	DisplayName string `xorm:"varchar(500) not null"`
	ContactName string `xorm:"varchar(500) null"`
	Status      string `xorm:"varchar(20) not null default 'potential' index"`
	Source      string `xorm:"varchar(50) null"`

	ResponsibleUserID int64 `xorm:"bigint null index"`

	Phone                  string `xorm:"varchar(100) null"`
	PhoneSecondary         string `xorm:"varchar(100) null"`
	Email                  string `xorm:"varchar(320) null"`
	EmailSecondary         string `xorm:"varchar(320) null"`
	Telegram               string `xorm:"varchar(250) null"`
	Viber                  string `xorm:"varchar(250) null"`
	WhatsApp               string `xorm:"varchar(250) null"`
	Website                string `xorm:"varchar(1000) null"`
	PreferredContactMethod string `xorm:"varchar(50) null"`
	PreferredLanguage      string `xorm:"varchar(50) null"`

	TaxID         string `xorm:"varchar(100) null"`
	LegalName     string `xorm:"varchar(500) null"`
	DirectorName  string `xorm:"varchar(500) null"`
	OwnershipForm string `xorm:"varchar(250) null"`
	Industry      string `xorm:"varchar(500) null"`
	EmployeeCount int    `xorm:"int null"`
	Requisites    string `xorm:"text null"`
	IBAN          string `xorm:"varchar(100) null"`
	Bank          string `xorm:"varchar(500) null"`
	MFO           string `xorm:"varchar(100) null"`
	VATNumber     string `xorm:"varchar(100) null"`
	TaxSystem     string `xorm:"varchar(250) null"`

	CommercialProposalFileID int64 `xorm:"bigint null index"`

	Updated time.Time `xorm:"updated not null"`
}

func (*clientProfile20260820120000) TableName() string { return "client_profiles" }

type clientContactPerson20260820120000 struct {
	ID        int64 `xorm:"bigint autoincr not null unique pk"`
	ProjectID int64 `xorm:"bigint not null index"`

	FullName               string `xorm:"varchar(500) not null"`
	Position               string `xorm:"varchar(250) null"`
	Department             string `xorm:"varchar(250) null"`
	Phone                  string `xorm:"varchar(100) null"`
	Email                  string `xorm:"varchar(320) null"`
	Telegram               string `xorm:"varchar(250) null"`
	Viber                  string `xorm:"varchar(250) null"`
	WhatsApp               string `xorm:"varchar(250) null"`
	Birthday               string `xorm:"varchar(10) null"`
	PreferredContactMethod string `xorm:"varchar(50) null"`
	DecisionRole           string `xorm:"varchar(50) null index"`
	Notes                  string `xorm:"text null"`
	PositionIndex          int    `xorm:"int not null default 0"`
}

func (*clientContactPerson20260820120000) TableName() string { return "client_contact_persons" }

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260820120000",
		Description: "finish CRM client profiles with contact persons and commercial proposal",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync2( //nolint:forbidigo // extends one CRM table and creates one new table
				&clientProfile20260820120000{},
				&clientContactPerson20260820120000{},
			)
		},
		Rollback: func(tx *xorm.Engine) error { return nil },
	})
}
