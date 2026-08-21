// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package migration

import (
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

// Patch 11.1: CRM foundation. Later CRM patches extend this table with
// legal/address/contact-person/document data without replacing project/task data.
type clientProfile20260820110000 struct {
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

	Updated time.Time `xorm:"updated not null"`
}

func (*clientProfile20260820110000) TableName() string { return "client_profiles" }

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260820110000",
		Description: "add CRM client profile foundation to projects",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync2(&clientProfile20260820110000{}) //nolint:forbidigo // brand-new table
		},
		Rollback: func(tx *xorm.Engine) error { return nil },
	})
}
