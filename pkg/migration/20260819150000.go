// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package migration

import (
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type workGroup20260819150000 struct {
	ID           int64     `xorm:"bigint autoincr not null unique pk"`
	Name         string    `xorm:"varchar(250) not null unique"`
	Description  string    `xorm:"text null"`
	LeaderUserID int64     `xorm:"bigint null index"`
	Created      time.Time `xorm:"created not null"`
	Updated      time.Time `xorm:"updated not null"`
}

func (*workGroup20260819150000) TableName() string { return "work_groups" }

type workGroupMember20260819150000 struct {
	ID      int64     `xorm:"bigint autoincr not null unique pk"`
	GroupID int64     `xorm:"bigint not null index unique(work_group_member)"`
	UserID  int64     `xorm:"bigint not null index unique(work_group_member)"`
	Created time.Time `xorm:"created not null"`
}

func (*workGroupMember20260819150000) TableName() string { return "work_group_members" }

type taskWorkGroupAssignee20260819150000 struct {
	ID      int64     `xorm:"bigint autoincr not null unique pk"`
	TaskID  int64     `xorm:"bigint not null index unique(task_work_group_assignee)"`
	GroupID int64     `xorm:"bigint not null index unique(task_work_group_assignee)"`
	Created time.Time `xorm:"created not null"`
}

func (*taskWorkGroupAssignee20260819150000) TableName() string { return "task_work_group_assignees" }

type taskAssigneeSource20260819150000 struct {
	ID         int64     `xorm:"bigint autoincr not null unique pk"`
	TaskID     int64     `xorm:"bigint not null index unique(task_assignee_source)"`
	UserID     int64     `xorm:"bigint not null index unique(task_assignee_source)"`
	SourceType string    `xorm:"varchar(30) not null unique(task_assignee_source)"`
	SourceID   int64     `xorm:"bigint not null default 0 unique(task_assignee_source)"`
	Created    time.Time `xorm:"created not null"`
}

func (*taskAssigneeSource20260819150000) TableName() string { return "task_assignee_sources" }

type existingTaskAssignee20260819150000 struct {
	TaskID int64 `xorm:"bigint 'task_id'"`
	UserID int64 `xorm:"bigint 'user_id'"`
}

func (*existingTaskAssignee20260819150000) TableName() string { return "task_assignees" }

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260819150000",
		Description: "add operational work groups and group task assignment sources",
		Migrate: func(tx *xorm.Engine) error {
			if err := tx.Sync2( //nolint:forbidigo // brand-new tables
				&workGroup20260819150000{},
				&workGroupMember20260819150000{},
				&taskWorkGroupAssignee20260819150000{},
				&taskAssigneeSource20260819150000{},
			); err != nil {
				return err
			}

			// Every pre-existing task assignment was made directly because work
			// groups did not exist yet. Seed those sources so later group removal
			// can never accidentally remove an older personal assignment.
			var existing []*existingTaskAssignee20260819150000
			if err := tx.Find(&existing); err != nil {
				return err
			}
			for _, row := range existing {
				has, err := tx.Where(
					"task_id = ? AND user_id = ? AND source_type = ? AND source_id = ?",
					row.TaskID, row.UserID, "direct", 0,
				).Exist(&taskAssigneeSource20260819150000{})
				if err != nil {
					return err
				}
				if !has {
					if _, err := tx.Insert(&taskAssigneeSource20260819150000{
						TaskID: row.TaskID, UserID: row.UserID, SourceType: "direct", SourceID: 0,
					}); err != nil {
						return err
					}
				}
			}
			return nil
		},
		Rollback: func(tx *xorm.Engine) error { return nil },
	})
}
