// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package models

import (
	"strings"
	"time"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// TaskChecklistItem is a structured checklist entry attached to a task.
// It deliberately is not a full Task: it is a lightweight, ordered unit of work
// which inherits permissions from its parent task.
type TaskChecklistItem struct {
	ID     int64  `xorm:"bigint autoincr not null unique pk" json:"id" readOnly:"true" doc:"The unique numeric id of this checklist item."`
	TaskID int64  `xorm:"bigint not null index" json:"task_id" readOnly:"true" doc:"The parent task id. Taken from the URL and controlled by the server."`
	Title  string `xorm:"text not null" json:"title" minLength:"1" maxLength:"1000" doc:"The checklist item text."`
	Done   bool   `xorm:"bool not null default false index" json:"done" doc:"Whether this checklist item is completed."`

	CompletedByID int64      `xorm:"bigint null index" json:"-"`
	CompletedBy   *user.User `xorm:"-" json:"completed_by,omitempty" readOnly:"true" doc:"The user who most recently completed this item. Null while the item is incomplete."`
	CompletedAt   *time.Time `xorm:"DATETIME null" json:"completed_at,omitempty" readOnly:"true" doc:"When this item was most recently completed. Null while the item is incomplete."`

	Position int64     `xorm:"bigint not null default 0 index" json:"position" readOnly:"true" doc:"Stable display order inside the parent task."`
	Created  time.Time `xorm:"created not null" json:"created" readOnly:"true" doc:"When this checklist item was created."`
	Updated  time.Time `xorm:"updated not null" json:"updated" readOnly:"true" doc:"When this checklist item was last changed."`
}

func (*TaskChecklistItem) TableName() string { return "task_checklist_items" }

// TaskChecklistState is the complete state returned to the task detail UI.
type TaskChecklistState struct {
	Items      []*TaskChecklistItem `json:"items" doc:"The ordered checklist items."`
	Total      int                  `json:"total" doc:"The total number of checklist items."`
	Completed  int                  `json:"completed" doc:"The number of completed checklist items."`
	TaskDone   bool                 `json:"task_done" doc:"Whether the parent task is currently completed."`
	TaskDoneAt *time.Time           `json:"task_done_at,omitempty" doc:"When the parent task was completed. Null while the parent task is incomplete."`
}

func loadChecklistCompletionUsers(s *xorm.Session, items []*TaskChecklistItem) error {
	ids := make([]int64, 0, len(items))
	for _, item := range items {
		if item.CompletedByID != 0 {
			ids = append(ids, item.CompletedByID)
		}
	}
	users, err := user.GetUsersByIDs(s, ids)
	if err != nil {
		return err
	}
	for _, item := range items {
		if completedBy, ok := users[item.CompletedByID]; ok {
			item.CompletedBy = completedBy
		}
	}
	return nil
}

// GetTaskChecklistState returns all structured checklist items and parent completion state.
func GetTaskChecklistState(s *xorm.Session, taskID int64) (*TaskChecklistState, error) {
	items := []*TaskChecklistItem{}
	if err := s.Where("task_id = ?", taskID).OrderBy("position asc, id asc").Find(&items); err != nil {
		return nil, err
	}
	if err := loadChecklistCompletionUsers(s, items); err != nil {
		return nil, err
	}

	task, err := GetTaskByIDSimple(s, taskID)
	if err != nil {
		return nil, err
	}

	completed := 0
	for _, item := range items {
		if item.Done {
			completed++
		}
	}

	var taskDoneAt *time.Time
	if !task.DoneAt.IsZero() {
		doneAt := task.DoneAt
		taskDoneAt = &doneAt
	}

	return &TaskChecklistState{
		Items:      items,
		Total:      len(items),
		Completed:  completed,
		TaskDone:   task.Done,
		TaskDoneAt: taskDoneAt,
	}, nil
}

// CreateTaskChecklistItem appends a pending item to a task.
func CreateTaskChecklistItem(s *xorm.Session, taskID int64, title string) (*TaskChecklistItem, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, ErrInvalidData{Message: "checklist item title must not be empty"}
	}

	type maxPositionRow struct {
		MaxPosition int64 `xorm:"max_position"`
	}
	row := &maxPositionRow{}
	if _, err := s.Table("task_checklist_items").Where("task_id = ?", taskID).Select("COALESCE(MAX(position), 0) AS max_position").Get(row); err != nil {
		return nil, err
	}
	maxPosition := row.MaxPosition

	item := &TaskChecklistItem{
		TaskID:   taskID,
		Title:    title,
		Done:     false,
		Position: maxPosition + 1,
	}
	if _, err := s.Insert(item); err != nil {
		return nil, err
	}
	return item, nil
}

// UpdateTaskChecklistItem updates the title/status and records the authenticated completer.
// The returned bool is false when the item does not exist under the supplied parent task.
func UpdateTaskChecklistItem(s *xorm.Session, a web.Auth, taskID, itemID int64, title string, done bool) (*TaskChecklistItem, bool, error) {
	saved := &TaskChecklistItem{}
	exists, err := s.Where("id = ? AND task_id = ?", itemID, taskID).Get(saved)
	if err != nil || !exists {
		return nil, exists, err
	}

	title = strings.TrimSpace(title)
	if title == "" {
		return nil, true, ErrInvalidData{Message: "checklist item title must not be empty"}
	}

	saved.Title = title
	if done != saved.Done {
		saved.Done = done
		if done {
			completedAt := time.Now()
			saved.CompletedAt = &completedAt
			if doer, err := user.GetFromAuth(a); err == nil && doer != nil {
				saved.CompletedByID = doer.ID
				saved.CompletedBy = doer
			} else {
				saved.CompletedByID = 0
				saved.CompletedBy = nil
			}
		} else {
			saved.CompletedAt = nil
			saved.CompletedByID = 0
			saved.CompletedBy = nil
		}
	}

	if _, err := s.ID(saved.ID).Cols("title", "done", "completed_by_id", "completed_at").Update(saved); err != nil {
		return nil, true, err
	}
	return saved, true, nil
}

// DeleteTaskChecklistItem deletes a single item from the supplied parent task.
func DeleteTaskChecklistItem(s *xorm.Session, taskID, itemID int64) (bool, error) {
	affected, err := s.Where("id = ? AND task_id = ?", itemID, taskID).Delete(&TaskChecklistItem{})
	return affected > 0, err
}

// SyncTaskDoneFromChecklist makes parent completion follow its structured checklist.
// Empty checklists do not change the task. Adding/reopening pending work reopens the task;
// completing the final item completes the task through the normal Task update path so
// buckets, timestamps, events and notifications continue to work.
func SyncTaskDoneFromChecklist(s *xorm.Session, a web.Auth, taskID int64) (*Task, error) {
	total, err := s.Where("task_id = ?", taskID).Count(&TaskChecklistItem{})
	if err != nil {
		return nil, err
	}

	task := &Task{ID: taskID}
	if err := task.ReadOne(s, a); err != nil {
		return nil, err
	}
	if total == 0 {
		return task, nil
	}

	doneCount, err := s.Where("task_id = ? AND done = ?", taskID, true).Count(&TaskChecklistItem{})
	if err != nil {
		return nil, err
	}
	shouldBeDone := doneCount == total
	if task.Done == shouldBeDone {
		return task, nil
	}

	// Use the fully hydrated task as the update payload. Task.updateSingleTask
	// reconciles assignees, reminders and favorites even for a field-scoped
	// update, so preserving those values is essential for an automatic action.
	update := *task
	update.Done = shouldBeDone
	if err := update.updateSingleTask(s, a, []string{"done"}); err != nil {
		return nil, err
	}

	// Repeating tasks immediately roll into their next occurrence. In that case
	// reset the structured checklist too, matching the next-occurrence semantics.
	if shouldBeDone && !update.Done {
		if _, err := s.Where("task_id = ?", taskID).Cols("done", "completed_by_id", "completed_at").Update(&TaskChecklistItem{
			Done:          false,
			CompletedByID: 0,
			CompletedAt:   nil,
		}); err != nil {
			return nil, err
		}
	}

	return &update, nil
}
