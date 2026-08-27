// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package migration

import (
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type googleCalendarConnection20260827110500 struct {
	ID int64 `xorm:"bigint autoincr not null unique pk"`

	UserID int64 `xorm:"bigint not null unique index"`

	GoogleEmail string `xorm:"varchar(320)"`

	RefreshTokenEncrypted string `xorm:"text"`

	VikunjaCalendarID string `xorm:"varchar(1024)"`

	SelectedCalendarIDs string `xorm:"text"`

	OAuthStateHash      string    `xorm:"varchar(64)"`
	OAuthStateExpiresAt time.Time `xorm:"index"`

	ConnectedAt time.Time

	Created time.Time `xorm:"created not null"`
	Updated time.Time `xorm:"updated not null"`
}

func (*googleCalendarConnection20260827110500) TableName() string {
	return "google_calendar_connections"
}

func init() {
	migrations = append(
		migrations,
		&xormigrate.Migration{
			ID:          "20260827110500",
			Description: "add per-user Google Calendar connections",

			Migrate: func(tx *xorm.Engine) error {
				return tx.Sync2(
					&googleCalendarConnection20260827110500{},
				) //nolint:forbidigo
			},

			Rollback: func(tx *xorm.Engine) error {
				return tx.DropTables(
					&googleCalendarConnection20260827110500{},
				)
			},
		},
	)
}
