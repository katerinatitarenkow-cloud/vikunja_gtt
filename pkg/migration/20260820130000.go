// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package migration

import (
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

// clientActivityEvent20260820130000 is deliberately generic: entity_type/entity_id
// allow future deals, invoices, objects, equipment and Wialon entities to use the
// same client history without changing this schema.
type clientActivityEvent20260820130000 struct {
	ID              int64     `xorm:"bigint autoincr not null unique pk"`
	ProjectID       int64     `xorm:"bigint not null index"`
	EventType       string    `xorm:"varchar(80) not null index"`
	ActorUserID     int64     `xorm:"bigint null index"`
	OccurredAt      time.Time `xorm:"DATETIME not null index"`
	Title           string    `xorm:"varchar(1000) null"`
	Description     string    `xorm:"text null"`
	EntityType      string    `xorm:"varchar(80) null index"`
	EntityID        int64     `xorm:"bigint null index"`
	MetadataJSON    string    `xorm:"text null"`
	SystemGenerated bool      `xorm:"bool not null default false index"`
	Created         time.Time `xorm:"created not null"`
}

func (*clientActivityEvent20260820130000) TableName() string { return "client_activity_events" }

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260820130000",
		Description: "add universal CRM client activity history",
		Migrate: func(tx *xorm.Engine) error {
			// Brand-new table: a normal Sync2 is safe and creates the declared indexes.
			return tx.Sync2(&clientActivityEvent20260820130000{}) //nolint:forbidigo
		},
		Rollback: func(tx *xorm.Engine) error {
			return tx.DropTables(&clientActivityEvent20260820130000{})
		},
	})
}
