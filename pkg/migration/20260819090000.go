// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package migration

import (
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type accessGroup20260819090000 struct {
	ID          int64  `xorm:"bigint autoincr not null unique pk"`
	Name        string `xorm:"varchar(250) not null unique"`
	Description string `xorm:"text null"`
	SystemKey   string `xorm:"varchar(50) null index"`
}

func (*accessGroup20260819090000) TableName() string { return "access_groups" }

type accessGroupPermission20260819090000 struct {
	ID         int64  `xorm:"bigint autoincr not null unique pk"`
	GroupID    int64  `xorm:"bigint not null index unique(access_group_permission)"`
	Permission string `xorm:"varchar(100) not null unique(access_group_permission)"`
}

func (*accessGroupPermission20260819090000) TableName() string { return "access_group_permissions" }

type accessGroupMember20260819090000 struct {
	ID      int64 `xorm:"bigint autoincr not null unique pk"`
	GroupID int64 `xorm:"bigint not null index unique(access_group_member)"`
	UserID  int64 `xorm:"bigint not null index unique(access_group_member)"`
}

func (*accessGroupMember20260819090000) TableName() string { return "access_group_members" }

type userProfile20260819090000 struct {
	UserID int64  `xorm:"bigint not null unique pk"`
	Phone  string `xorm:"varchar(100) null"`
	Notes  string `xorm:"text null"`
}

func (*userProfile20260819090000) TableName() string { return "user_profiles" }

type wialonSettings20260819090000 struct {
	ID             int64  `xorm:"bigint not null unique pk"`
	Enabled        bool   `xorm:"not null default false"`
	APIURL         string `xorm:"varchar(500) not null"`
	Token          string `xorm:"text null"`
	TimeoutSeconds int    `xorm:"not null default 30"`
	TrackMaxPoints int    `xorm:"not null default 5000"`
}

func (*wialonSettings20260819090000) TableName() string { return "wialon_settings" }

type userAccessSeed20260819090000 struct {
	ID      int64 `xorm:"bigint pk 'id'"`
	IsAdmin bool  `xorm:"'is_admin'"`
}

func (*userAccessSeed20260819090000) TableName() string { return "users" }

var accessPermissions20260819090000 = []string{
	"projects.view", "projects.manage", "tasks.view", "tasks.manage",
	"labels.view", "labels.manage", "teams.view", "teams.manage",
	"kanban.use", "time_tracking.use", "wialon.view",
}

func ensureSystemAccessGroup20260819090000(tx *xorm.Engine, name, key, description string) (*accessGroup20260819090000, error) {
	group := &accessGroup20260819090000{}
	has, err := tx.Where("system_key = ?", key).Get(group)
	if err != nil {
		return nil, err
	}
	if !has {
		group = &accessGroup20260819090000{Name: name, SystemKey: key, Description: description}
		if _, err := tx.Insert(group); err != nil {
			return nil, err
		}
	}
	for _, permission := range accessPermissions20260819090000 {
		exists, err := tx.Where("group_id = ? AND permission = ?", group.ID, permission).Exist(&accessGroupPermission20260819090000{})
		if err != nil {
			return nil, err
		}
		if !exists {
			if _, err := tx.Insert(&accessGroupPermission20260819090000{GroupID: group.ID, Permission: permission}); err != nil {
				return nil, err
			}
		}
	}
	return group, nil
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260819090000",
		Description: "add access groups, user personnel cards and UI-managed Wialon settings",
		Migrate: func(tx *xorm.Engine) error {
			if err := tx.Sync2( //nolint:forbidigo // brand-new tables
				&accessGroup20260819090000{},
				&accessGroupPermission20260819090000{},
				&accessGroupMember20260819090000{},
				&userProfile20260819090000{},
				&wialonSettings20260819090000{},
			); err != nil {
				return err
			}

			adminGroup, err := ensureSystemAccessGroup20260819090000(tx, "Админ", "admin", "Системные администраторы")
			if err != nil {
				return err
			}
			usersGroup, err := ensureSystemAccessGroup20260819090000(tx, "Пользователи", "users", "Обычные пользователи")
			if err != nil {
				return err
			}

			var users []*userAccessSeed20260819090000
			if err := tx.Find(&users); err != nil {
				return err
			}
			for _, u := range users {
				groupID := usersGroup.ID
				if u.IsAdmin {
					groupID = adminGroup.ID
				}
				exists, err := tx.Where("user_id = ?", u.ID).Exist(&accessGroupMember20260819090000{})
				if err != nil {
					return err
				}
				if !exists {
					if _, err := tx.Insert(&accessGroupMember20260819090000{GroupID: groupID, UserID: u.ID}); err != nil {
						return err
					}
				}
			}
			return nil
		},
		Rollback: func(tx *xorm.Engine) error { return nil },
	})
}
