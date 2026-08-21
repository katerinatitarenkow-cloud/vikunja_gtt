// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package migration

import (
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260821150000",
		Description: "reconcile missing CRM tables after an inconsistent migration state",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync2( //nolint:forbidigo // repairs missing tables and columns without removing existing CRM data
				&clientProfile20260820120000{},
				&clientAddress20260820113000{},
				&clientContactPerson20260820120000{},
				&clientActivityEvent20260820130000{},
				&clientCustomField20260820140000{},
			)
		},
		Rollback: func(tx *xorm.Engine) error { return nil },
	})
}
