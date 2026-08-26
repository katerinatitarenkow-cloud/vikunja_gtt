// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package migration

import (
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type mailboxAttachment20260826090000 struct {
	ID        int64     `xorm:"bigint autoincr not null unique pk"`
	MessageID int64     `xorm:"bigint not null index"`
	FileID    int64     `xorm:"bigint not null index"`
	Created   time.Time `xorm:"created not null index"`
}

func (*mailboxAttachment20260826090000) TableName() string {
	return "user_mailbox_attachments"
}

func init() {
	migrations = append(
		migrations,
		&xormigrate.Migration{
			ID:          "20260826090000",
			Description: "add attachments to internal mailbox messages",

			Migrate: func(tx *xorm.Engine) error {
				return tx.Sync2(
					&mailboxAttachment20260826090000{},
				) //nolint:forbidigo
			},

			Rollback: func(tx *xorm.Engine) error {
				return tx.DropTables(
					&mailboxAttachment20260826090000{},
				)
			},
		},
	)
}
