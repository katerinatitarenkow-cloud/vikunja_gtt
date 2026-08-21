// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package migration

import (
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type clientCustomField20260820140000 struct {
	ID        int64     `xorm:"bigint autoincr not null unique pk"`
	ProjectID int64     `xorm:"bigint not null index"`
	Name      string    `xorm:"varchar(500) not null"`
	Value     string    `xorm:"text null"`
	Position  int       `xorm:"int not null default 0 index"`
	Created   time.Time `xorm:"created not null"`
	Updated   time.Time `xorm:"updated not null"`
}

func (*clientCustomField20260820140000) TableName() string { return "client_custom_fields" }

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260820140000",
		Description: "add custom name-value fields to CRM client cards",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync2(&clientCustomField20260820140000{}) //nolint:forbidigo // brand-new CRM table
		},
		Rollback: func(tx *xorm.Engine) error { return nil },
	})
}
