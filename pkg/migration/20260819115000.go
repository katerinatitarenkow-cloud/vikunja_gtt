// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package migration

import (
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type accessAdminBootstrapUser20260819115000 struct {
	ID       int64  `xorm:"bigint pk 'id'"`
	Username string `xorm:"varchar(250) 'username'"`
	IsAdmin  bool   `xorm:"'is_admin'"`
}

func (*accessAdminBootstrapUser20260819115000) TableName() string { return "users" }

type accessAdminBootstrapGroup20260819115000 struct {
	ID        int64  `xorm:"bigint pk 'id'"`
	SystemKey string `xorm:"varchar(50) 'system_key'"`
}

func (*accessAdminBootstrapGroup20260819115000) TableName() string { return "access_groups" }

type accessAdminBootstrapMember20260819115000 struct {
	ID      int64 `xorm:"bigint autoincr not null unique pk"`
	GroupID int64 `xorm:"bigint not null 'group_id'"`
	UserID  int64 `xorm:"bigint not null 'user_id'"`
}

func (*accessAdminBootstrapMember20260819115000) TableName() string { return "access_group_members" }

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260819115000",
		Description: "bootstrap custom core administrator on free installations",
		Migrate: func(tx *xorm.Engine) error {
			// Do not change an instance that already has an explicit administrator.
			adminCount, err := tx.Where("is_admin = ?", true).Count(&accessAdminBootstrapUser20260819115000{})
			if err != nil || adminCount > 0 {
				return err
			}

			// The Windows custom distribution is initially provisioned with the
			// local account named "admin". Promote only that explicit bootstrap
			// account; never guess another user on an existing multi-user instance.
			candidate := &accessAdminBootstrapUser20260819115000{}
			has, err := tx.Where("LOWER(username) = ?", "admin").Get(candidate)
			if err != nil || !has {
				return err
			}
			if _, err = tx.ID(candidate.ID).Cols("is_admin").Update(&accessAdminBootstrapUser20260819115000{IsAdmin: true}); err != nil {
				return err
			}

			adminGroup := &accessAdminBootstrapGroup20260819115000{}
			has, err = tx.Where("system_key = ?", "admin").Get(adminGroup)
			if err != nil || !has {
				return err
			}
			usersGroup := &accessAdminBootstrapGroup20260819115000{}
			hasUsers, err := tx.Where("system_key = ?", "users").Get(usersGroup)
			if err != nil {
				return err
			}
			if hasUsers {
				if _, err = tx.Where("user_id = ? AND group_id = ?", candidate.ID, usersGroup.ID).Delete(&accessAdminBootstrapMember20260819115000{}); err != nil {
					return err
				}
			}
			exists, err := tx.Where("user_id = ? AND group_id = ?", candidate.ID, adminGroup.ID).Exist(&accessAdminBootstrapMember20260819115000{})
			if err != nil {
				return err
			}
			if !exists {
				_, err = tx.Insert(&accessAdminBootstrapMember20260819115000{UserID: candidate.ID, GroupID: adminGroup.ID})
			}
			return err
		},
		Rollback: func(tx *xorm.Engine) error { return nil },
	})
}
