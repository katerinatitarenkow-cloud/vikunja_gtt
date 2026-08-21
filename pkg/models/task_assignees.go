// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/builder"
	"xorm.io/xorm"
)

// TaskAssginee represents an assignment of a user to a task
type TaskAssginee struct {
	ID      int64     `xorm:"bigint autoincr not null unique pk" json:"-"`
	TaskID  int64     `xorm:"bigint INDEX not null" json:"-" param:"projecttask"`
	UserID  int64     `xorm:"bigint INDEX not null" json:"user_id" param:"user" doc:"The id of the user to assign to the task. The user must have access to the task's project."`
	Created time.Time `xorm:"created not null" json:"created" readOnly:"true" doc:"A timestamp when this assignment was created. You cannot change this value."`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

// TableName makes a pretty table name
func (*TaskAssginee) TableName() string {
	return "task_assignees"
}

// TaskAssigneeWithUser is a helper type to deal with user joins
type TaskAssigneeWithUser struct {
	TaskID    int64
	user.User `xorm:"extends"`
}

func getRawTaskAssigneesForTasks(s *xorm.Session, taskIDs []int64) (taskAssignees []*TaskAssigneeWithUser, err error) {
	taskAssignees = []*TaskAssigneeWithUser{}
	err = s.Table("task_assignees").
		Select("task_id, users.*").
		In("task_id", taskIDs).
		Join("INNER", "users", "task_assignees.user_id = users.id").
		Find(&taskAssignees)
	return
}

func ensureTaskAssigneeSource(s *xorm.Session, taskID, userID int64, sourceType string, sourceID int64) (created bool, err error) {
	exists, err := s.Where("task_id = ? AND user_id = ? AND source_type = ? AND source_id = ?", taskID, userID, sourceType, sourceID).
		Exist(&TaskAssigneeSource{})
	if err != nil || exists {
		return false, err
	}
	_, err = s.Insert(&TaskAssigneeSource{TaskID: taskID, UserID: userID, SourceType: sourceType, SourceID: sourceID})
	return err == nil, err
}

func taskAssigneeSources(s *xorm.Session, taskID, userID int64) ([]*TaskAssigneeSource, error) {
	rows := []*TaskAssigneeSource{}
	err := s.Where("task_id = ? AND user_id = ?", taskID, userID).Find(&rows)
	return rows, err
}

func hasTaskAssigneeSource(s *xorm.Session, taskID, userID int64, sourceType string, sourceID int64) (bool, error) {
	return s.Where("task_id = ? AND user_id = ? AND source_type = ? AND source_id = ?", taskID, userID, sourceType, sourceID).
		Exist(&TaskAssigneeSource{})
}

func hasAnyTaskAssigneeSource(s *xorm.Session, taskID, userID int64) (bool, error) {
	return s.Where("task_id = ? AND user_id = ?", taskID, userID).Exist(&TaskAssigneeSource{})
}

// removeTaskAssigneeSourceAndMaybeMaterialized removes one reason for an
// assignment. The ordinary Vikunja task_assignees row is deleted only when no
// direct or group source remains.
func removeTaskAssigneeSourceAndMaybeMaterialized(s *xorm.Session, taskID, userID int64, sourceType string, sourceID int64, a web.Auth) error {
	if _, err := s.Where("task_id = ? AND user_id = ? AND source_type = ? AND source_id = ?", taskID, userID, sourceType, sourceID).
		Delete(&TaskAssigneeSource{}); err != nil {
		return err
	}
	hasSource, err := hasAnyTaskAssigneeSource(s, taskID, userID)
	if err != nil || hasSource {
		return err
	}
	exists, err := s.Where("task_id = ? AND user_id = ?", taskID, userID).Exist(&TaskAssginee{})
	if err != nil || !exists {
		return err
	}
	if _, err = s.Where("task_id = ? AND user_id = ?", taskID, userID).Delete(&TaskAssginee{}); err != nil {
		return err
	}
	if err = updateProjectByTaskID(s, taskID); err != nil {
		return err
	}
	doer := doerFromAuth(s, a)
	task, err := GetTaskByIDSimple(s, taskID)
	if err != nil {
		return err
	}
	events.DispatchOnCommit(s, &TaskAssigneeDeletedEvent{Task: &task, Assignee: &user.User{ID: userID}, Doer: doer})
	events.DispatchOnCommit(s, &TaskUpdatedEvent{Task: &task, Doer: doer})
	return nil
}

// Create or update a bunch of task assignees. The incoming list represents
// direct assignments. Work-group assignments are preserved independently.
func (t *Task) updateTaskAssignees(s *xorm.Session, assignees []*user.User, doer web.Auth) (err error) {
	currentRows, err := getRawTaskAssigneesForTasks(s, []int64{t.ID})
	if err != nil {
		return err
	}
	current := make(map[int64]*user.User, len(currentRows))
	for i := range currentRows {
		current[currentRows[i].ID] = &currentRows[i].User
	}

	desired := make(map[int64]*user.User, len(assignees))
	for _, assignee := range assignees {
		if assignee != nil && assignee.ID > 0 {
			desired[assignee.ID] = assignee
		}
	}

	// Remove direct sources that disappeared from the payload. Group sources
	// remain authoritative and keep their users assigned.
	directSources := []*TaskAssigneeSource{}
	if err := s.Where("task_id = ? AND source_type = ?", t.ID, TaskAssigneeSourceDirect).Find(&directSources); err != nil {
		return err
	}
	for _, source := range directSources {
		if _, keep := desired[source.UserID]; keep {
			continue
		}
		if err := removeTaskAssigneeSourceAndMaybeMaterialized(s, t.ID, source.UserID, TaskAssigneeSourceDirect, 0, doer); err != nil {
			return err
		}
	}

	project, err := GetProjectSimpleByID(s, t.ProjectID)
	if err != nil {
		return err
	}
	for userID := range desired {
		if _, alreadyMaterialized := current[userID]; alreadyMaterialized {
			// If there are no source rows this is a legacy/direct assignment.
			// Seed the direct source. If the user is already present only because
			// of a work group, leave it group-only; the regular task update sends
			// the complete visible assignee list on every field change.
			hasAny, err := hasAnyTaskAssigneeSource(s, t.ID, userID)
			if err != nil {
				return err
			}
			if !hasAny {
				if _, err := ensureTaskAssigneeSource(s, t.ID, userID, TaskAssigneeSourceDirect, 0); err != nil {
					return err
				}
			}
			continue
		}
		if err := t.addNewAssigneeByIDWithSource(s, userID, project, doer, TaskAssigneeSourceDirect, 0, false); err != nil {
			return err
		}
	}

	actualRows, err := getRawTaskAssigneesForTasks(s, []int64{t.ID})
	if err != nil {
		return err
	}
	actual := make([]*user.User, 0, len(actualRows))
	for i := range actualRows {
		actual = append(actual, &actualRows[i].User)
	}
	t.setTaskAssignees(actual)
	return updateProjectLastUpdated(s, &Project{ID: t.ProjectID})
}

// Small helper functions to set the new assignees in various places
func (t *Task) setTaskAssignees(assignees []*user.User) {
	if len(assignees) == 0 {
		t.Assignees = nil
		return
	}
	t.Assignees = assignees
}

// Delete a task assignee
// @Summary Delete an assignee
// @Description Un-assign a user from a task.
// @tags assignees
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param taskID path int true "Task ID"
// @Param userID path int true "Assignee user ID"
// @Success 200 {object} models.Message "The assignee was successfully deleted."
// @Failure 403 {object} web.HTTPError "Not allowed to delete the assignee."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{taskID}/assignees/{userID} [delete]
func (la *TaskAssginee) Delete(s *xorm.Session, a web.Auth) (err error) {
	// If this user was also assigned directly, remove only that direct source.
	// The materialized assignee remains while a work-group source still exists.
	directExists, err := hasTaskAssigneeSource(s, la.TaskID, la.UserID, TaskAssigneeSourceDirect, 0)
	if err != nil {
		return err
	}
	if directExists {
		return removeTaskAssigneeSourceAndMaybeMaterialized(s, la.TaskID, la.UserID, TaskAssigneeSourceDirect, 0, a)
	}

	// A user assigned only through a work group cannot be removed one person at
	// a time without breaking the semantic group assignment. Remove the group
	// assignment (or membership) first.
	groupSourceExists, err := s.Where("task_id = ? AND user_id = ? AND source_type = ?", la.TaskID, la.UserID, TaskAssigneeSourceWorkGroup).
		Exist(&TaskAssigneeSource{})
	if err != nil {
		return err
	}
	if groupSourceExists {
		return ErrInvalidData{Message: "this assignee is assigned through a work group; remove the group assignment first"}
	}

	// Compatibility for a legacy assignment created before task_assignee_sources
	// existed (or by an older integration).
	_, err = s.Delete(&TaskAssginee{TaskID: la.TaskID, UserID: la.UserID})
	if err != nil {
		return err
	}
	if err = updateProjectByTaskID(s, la.TaskID); err != nil {
		return err
	}
	doer := doerFromAuth(s, a)
	task, err := GetTaskByIDSimple(s, la.TaskID)
	if err != nil {
		return err
	}
	events.DispatchOnCommit(s, &TaskAssigneeDeletedEvent{Task: &task, Assignee: &user.User{ID: la.UserID}, Doer: doer})
	events.DispatchOnCommit(s, &TaskUpdatedEvent{Task: &task, Doer: doer})
	return nil
}

// Create adds a new assignee to a task
// @Summary Add a new assignee to a task
// @Description Adds a new assignee to a task. The assignee needs to have access to the project, the doer must be able to edit this task.
// @tags assignees
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param assignee body models.TaskAssginee true "The assingee object"
// @Param taskID path int true "Task ID"
// @Success 201 {object} models.TaskAssginee "The created assingee object."
// @Failure 400 {object} web.HTTPError "Invalid assignee object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{taskID}/assignees [put]
func (la *TaskAssginee) Create(s *xorm.Session, a web.Auth) (err error) {

	// Get the project to perform later checks
	project, err := GetProjectSimpleByTaskID(s, la.TaskID)
	if err != nil {
		return
	}

	task := &Task{ID: la.TaskID}
	return task.addNewAssigneeByID(s, la.UserID, project, a)
}

func (t *Task) addNewAssigneeByID(s *xorm.Session, newAssigneeID int64, project *Project, auth web.Auth) (err error) {
	return t.addNewAssigneeByIDWithSource(s, newAssigneeID, project, auth, TaskAssigneeSourceDirect, 0, true)
}

func (t *Task) addNewAssigneeByIDWithSource(s *xorm.Session, newAssigneeID int64, project *Project, auth web.Auth, sourceType string, sourceID int64, failIfSameSourceExists bool) (err error) {
	newAssignee, err := user.GetUserByID(s, newAssigneeID)
	if err != nil {
		return err
	}
	canRead, _, err := project.CanRead(s, newAssignee)
	if err != nil {
		return err
	}
	if !canRead {
		return ErrUserDoesNotHaveAccessToProject{project.ID, newAssigneeID}
	}

	sourceExists, err := hasTaskAssigneeSource(s, t.ID, newAssigneeID, sourceType, sourceID)
	if err != nil {
		return err
	}
	if sourceExists {
		if failIfSameSourceExists {
			return &ErrUserAlreadyAssigned{UserID: newAssigneeID, TaskID: t.ID}
		}
		return nil
	}

	assignmentExists, err := s.Where("task_id = ? AND user_id = ?", t.ID, newAssigneeID).Exist(&TaskAssginee{})
	if err != nil {
		return err
	}
	if assignmentExists {
		// Preserve an old direct assignment when a group is added on top of it.
		hasAny, err := hasAnyTaskAssigneeSource(s, t.ID, newAssigneeID)
		if err != nil {
			return err
		}
		if !hasAny && sourceType == TaskAssigneeSourceWorkGroup {
			if _, err := ensureTaskAssigneeSource(s, t.ID, newAssigneeID, TaskAssigneeSourceDirect, 0); err != nil {
				return err
			}
		}
		_, err = ensureTaskAssigneeSource(s, t.ID, newAssigneeID, sourceType, sourceID)
		return err
	}

	if _, err := s.Insert(&TaskAssginee{TaskID: t.ID, UserID: newAssigneeID}); err != nil {
		return err
	}
	if _, err := ensureTaskAssigneeSource(s, t.ID, newAssigneeID, sourceType, sourceID); err != nil {
		return err
	}

	sub := &Subscription{UserID: newAssigneeID, EntityType: SubscriptionEntityTask, EntityID: t.ID}
	if err := sub.Create(s, newAssignee); err != nil && !IsErrSubscriptionAlreadyExists(err) {
		return err
	}

	doer := doerFromAuth(s, auth)
	task, err := GetTaskSimple(s, &Task{ID: t.ID})
	if err != nil {
		return err
	}
	events.DispatchOnCommit(s, &TaskAssigneeCreatedEvent{Task: &task, Assignee: newAssignee, Doer: doer})
	events.DispatchOnCommit(s, &TaskUpdatedEvent{Task: &task, Doer: doer})
	return updateProjectLastUpdated(s, &Project{ID: t.ProjectID})
}

// ReadAll gets all assignees for a task
// @Summary Get all assignees for a task
// @Description Returns an array with all assignees for this task.
// @tags assignees
// @Accept json
// @Produce json
// @Param page query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page."
// @Param s query string false "Search assignees by their username."
// @Param taskID path int true "Task ID"
// @Security JWTKeyAuth
// @Success 200 {array} user.User "The assignees"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{taskID}/assignees [get]
func (la *TaskAssginee) ReadAll(s *xorm.Session, a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	task, err := GetProjectSimpleByTaskID(s, la.TaskID)
	if err != nil {
		return nil, 0, 0, err
	}

	can, _, err := task.CanRead(s, a)
	if err != nil {
		return nil, 0, 0, err
	}
	if !can {
		return nil, 0, 0, ErrGenericForbidden{}
	}
	limit, start := getLimitFromPageIndex(page, perPage)
	var taskAssignees []*user.User
	query := s.Table("task_assignees").
		Select("users.*").
		Join("INNER", "users", "task_assignees.user_id = users.id").
		Where(builder.And(
			builder.Eq{"task_id": la.TaskID},
			db.ILIKE("users.username", search),
		))
	if limit > 0 {
		query = query.Limit(limit, start)
	}
	err = query.Find(&taskAssignees)
	if err != nil {
		return nil, 0, 0, err
	}

	numberOfTotalItems, err = s.Table("task_assignees").
		Join("INNER", "users", "task_assignees.user_id = users.id").
		Where(builder.And(
			builder.Eq{"task_id": la.TaskID},
			db.ILIKE("users.username", search),
		)).
		Count(&TaskAssginee{})
	return taskAssignees, len(taskAssignees), numberOfTotalItems, err
}

// BulkAssignees is a helper struct used to update multiple assignees at once.
type BulkAssignees struct {
	// A project with all assignees
	Assignees []*user.User `json:"assignees" doc:"The full set of users to assign to the task. This replaces the task's current assignees: users not in this list are unassigned. Pass an empty array to unassign everyone. Each user must have access to the task's project."`
	TaskID    int64        `json:"-" param:"projecttask"`

	web.CRUDable    `json:"-"`
	web.Permissions `json:"-"`
}

// Create adds new assignees to a task
// @Summary Add multiple new assignees to a task
// @Description Adds multiple new assignees to a task. The assignee needs to have access to the project, the doer must be able to edit this task. Every user not in the project will be unassigned from the task, pass an empty array to unassign everyone.
// @tags assignees
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param assignee body models.BulkAssignees true "The array of assignees"
// @Param taskID path int true "Task ID"
// @Success 201 {object} models.TaskAssginee "The created assingees object."
// @Failure 400 {object} web.HTTPError "Invalid assignee object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{taskID}/assignees/bulk [post]
func (ba *BulkAssignees) Create(s *xorm.Session, a web.Auth) (err error) {
	task, err := GetTaskByIDSimple(s, ba.TaskID)
	if err != nil {
		return
	}
	assignees, err := getRawTaskAssigneesForTasks(s, []int64{task.ID})
	if err != nil {
		return err
	}
	for i := range assignees {
		task.Assignees = append(task.Assignees, &assignees[i].User)
	}

	err = task.updateTaskAssignees(s, ba.Assignees, a)
	return
}
