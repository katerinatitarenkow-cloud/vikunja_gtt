// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package migration

import (
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

// Patch 11.2 extends the CRM foundation from 20260820110000 with legal
// company data and a dedicated address table. Sync2 is used here because it
// only adds the missing columns to the already-existing client_profiles table.
type clientProfile20260820113000 struct {
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

	Updated time.Time `xorm:"updated not null"`
}

func (*clientProfile20260820113000) TableName() string { return "client_profiles" }

type clientAddress20260820113000 struct {
	ID         int64   `xorm:"bigint autoincr not null unique pk"`
	ProjectID  int64   `xorm:"bigint not null index unique(client_address_type)"`
	Type       string  `xorm:"varchar(30) not null unique(client_address_type)"`
	Country    string  `xorm:"varchar(250) null"`
	Region     string  `xorm:"varchar(250) null"`
	City       string  `xorm:"varchar(250) null"`
	Street     string  `xorm:"varchar(500) null"`
	House      string  `xorm:"varchar(100) null"`
	Office     string  `xorm:"varchar(100) null"`
	PostalCode string  `xorm:"varchar(50) null"`
	Latitude   float64 `xorm:"double null"`
	Longitude  float64 `xorm:"double null"`
}

func (*clientAddress20260820113000) TableName() string { return "client_addresses" }

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260820113000",
		Description: "extend CRM client profiles with legal data and addresses",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync2( //nolint:forbidigo // extends one existing CRM table and creates one new table
				&clientProfile20260820113000{},
				&clientAddress20260820113000{},
			)
		},
		Rollback: func(tx *xorm.Engine) error { return nil },
	})
}
