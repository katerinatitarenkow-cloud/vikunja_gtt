// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package migration

import (
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type userMailboxMessage20260820150000 struct {
	ID          int64 `xorm:"bigint autoincr not null unique pk"`
	SenderID    int64 `xorm:"bigint not null index"`
	RecipientID int64 `xorm:"bigint not null index"`
	ReplyToID   int64 `xorm:"bigint null index"`

	Subject string `xorm:"varchar(500) not null"`
	Body    string `xorm:"text not null"`

	ReadAt           time.Time `xorm:"datetime null index"`
	SenderDeleted    bool      `xorm:"bool not null default false index"`
	RecipientDeleted bool      `xorm:"bool not null default false index"`

	Created time.Time `xorm:"created not null index"`
	Updated time.Time `xorm:"updated not null"`
}

func (*userMailboxMessage20260820150000) TableName() string { return "user_mailbox_messages" }

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260820150000",
		Description: "add private per-user mailbox messages",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync2(&userMailboxMessage20260820150000{}) //nolint:forbidigo // brand-new table
		},
		Rollback: func(tx *xorm.Engine) error { return nil },
	})
}
