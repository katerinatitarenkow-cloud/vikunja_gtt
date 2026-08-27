package migration

import (
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type prospectCampaign20260827134000 struct {
	ID int64 `xorm:"bigint autoincr not null unique pk"`

	Name        string `xorm:"varchar(500) not null index"`
	Description string `xorm:"text null"`

	CreatedByUserID int64 `xorm:"bigint not null index"`
	Archived        bool  `xorm:"bool not null default false index"`

	Created time.Time `xorm:"created not null"`
	Updated time.Time `xorm:"updated not null"`
}

func (*prospectCampaign20260827134000) TableName() string {
	return "crm_prospect_campaigns"
}

type prospect20260827134000 struct {
	ID int64 `xorm:"bigint autoincr not null unique pk"`

	CampaignID int64 `xorm:"bigint null index"`

	CompanyName string `xorm:"varchar(500) null index"`
	ContactName string `xorm:"varchar(500) null"`

	Phone           string `xorm:"varchar(100) null index"`
	PhoneNormalized string `xorm:"varchar(50) null index"`
	PhoneSecondary  string `xorm:"varchar(100) null"`

	Email   string `xorm:"varchar(320) null index"`
	Website string `xorm:"varchar(1000) null"`

	Region string `xorm:"varchar(250) null index"`
	City   string `xorm:"varchar(250) null index"`

	Notes string `xorm:"text null"`

	Status string `xorm:"varchar(40) not null default 'new' index"`

	ResponsibleUserID int64 `xorm:"bigint null index"`

	LastContactAt *time.Time `xorm:"DATETIME null index"`
	NextContactAt *time.Time `xorm:"DATETIME null index"`

	ProjectID int64 `xorm:"bigint null index"`

	ImportRowNumber int    `xorm:"int null"`
	RawDataJSON     string `xorm:"text null"`

	Created time.Time `xorm:"created not null index"`
	Updated time.Time `xorm:"updated not null"`
}

func (*prospect20260827134000) TableName() string {
	return "crm_prospects"
}

type prospectCall20260827134000 struct {
	ID int64 `xorm:"bigint autoincr not null unique pk"`

	ProspectID int64 `xorm:"bigint not null index"`
	UserID     int64 `xorm:"bigint not null index"`

	OccurredAt time.Time `xorm:"DATETIME not null index"`

	Outcome string `xorm:"varchar(40) not null index"`

	DurationMinutes int    `xorm:"int not null default 0"`
	Note            string `xorm:"text null"`

	NextContactAt *time.Time `xorm:"DATETIME null index"`

	Created time.Time `xorm:"created not null"`
}

func (*prospectCall20260827134000) TableName() string {
	return "crm_prospect_calls"
}

func init() {
	migrations = append(
		migrations,
		&xormigrate.Migration{
			ID:          "20260827134000",
			Description: "add CRM call center prospects and campaigns",

			Migrate: func(tx *xorm.Engine) error {
				return tx.Sync2(
					&prospectCampaign20260827134000{},
					&prospect20260827134000{},
					&prospectCall20260827134000{},
				)
			},

			Rollback: func(tx *xorm.Engine) error {
				return tx.DropTables(
					&prospectCall20260827134000{},
					&prospect20260827134000{},
					&prospectCampaign20260827134000{},
				)
			},
		},
	)
}
