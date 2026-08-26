package migration

import (
"time"

"src.techknowlogick.com/xormigrate"
"xorm.io/xorm"
)

type taskChecklistItemRepair20260826000000 struct {
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

func (*taskChecklistItemRepair20260826000000) TableName() string {
return "task_checklist_items"
}

func init() {
migrations = append(migrations, &xormigrate.Migration{
ID:          "20260826000000",
Description: "repair missing task checklist items table",
Migrate: func(tx *xorm.Engine) error {
return tx.Sync2(&taskChecklistItemRepair20260826000000{})
},
Rollback: func(tx *xorm.Engine) error {
// Repair migration: do not delete user checklist data on rollback.
return nil
},
})
}
