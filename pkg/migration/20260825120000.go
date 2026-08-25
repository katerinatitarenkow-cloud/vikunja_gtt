// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package migration

import (
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type project20260825120000 struct {
	ID          int64 `xorm:"bigint autoincr not null unique pk"`
	IsCompleted bool  `xorm:"not null default false"`
}

func (project20260825120000) TableName() string {
	return "projects"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260825120000",
		Description: "add completed state to projects",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync2(new(project20260825120000))
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
