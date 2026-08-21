// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package migration

import (
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type taskChecklistItem20260819165000 struct {
	ID            int64      `xorm:"bigint autoincr not null unique pk"`
	TaskID        int64      `xorm:"bigint not null index"`
	Title         string     `xorm:"text not null"`
	Done          bool       `xorm:"bool not null default false index"`
	CompletedByID int64      `xorm:"bigint null index"`
	CompletedAt   *time.Time `xorm:"DATETIME null"`
	Position      int64      `xorm:"bigint not null default 0 index"`
	Created       time.Time  `xorm:"created not null"`
	Updated       time.Time  `xorm:"updated not null"`
}

func (*taskChecklistItem20260819165000) TableName() string { return "task_checklist_items" }

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260819165000",
		Description: "add structured task checklist items with completion audit",
		Migrate: func(tx *xorm.Engine) error {
			// Brand-new table: Sync2 is safe here and creates the declared indexes.
			return tx.Sync2(&taskChecklistItem20260819165000{}) //nolint:forbidigo
		},
		Rollback: func(tx *xorm.Engine) error {
			return tx.DropTables(&taskChecklistItem20260819165000{})
		},
	})
}
