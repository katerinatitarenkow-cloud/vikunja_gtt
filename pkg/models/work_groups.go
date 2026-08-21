// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package models

import (
	"sort"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

const (
	TaskAssigneeSourceDirect    = "direct"
	TaskAssigneeSourceWorkGroup = "work_group"
)

// WorkGroup is an operational grouping of users. It deliberately has no
// permission fields: access rights are handled by AccessGroup (shown as roles
// in the UI), while work groups exist only for coordination and task assignment.
type WorkGroup struct {
	ID           int64     `xorm:"bigint autoincr not null unique pk" json:"id"`
	Name         string    `xorm:"varchar(250) not null unique" json:"name"`
	Description  string    `xorm:"text null" json:"description"`
	LeaderUserID int64     `xorm:"bigint null index" json:"leader_user_id"`
	Created      time.Time `xorm:"created not null" json:"created" readOnly:"true"`
	Updated      time.Time `xorm:"updated not null" json:"updated" readOnly:"true"`
}

func (*WorkGroup) TableName() string { return "work_groups" }

type WorkGroupMember struct {
	ID      int64     `xorm:"bigint autoincr not null unique pk" json:"id"`
	GroupID int64     `xorm:"bigint not null index unique(work_group_member)" json:"group_id"`
	UserID  int64     `xorm:"bigint not null index unique(work_group_member)" json:"user_id"`
	Created time.Time `xorm:"created not null" json:"created" readOnly:"true"`
}

func (*WorkGroupMember) TableName() string { return "work_group_members" }

// TaskWorkGroupAssignee keeps the semantic assignment to a group, in addition
// to the materialized user assignees used by the rest of Vikunja.
type TaskWorkGroupAssignee struct {
	ID      int64     `xorm:"bigint autoincr not null unique pk" json:"id"`
	TaskID  int64     `xorm:"bigint not null index unique(task_work_group_assignee)" json:"task_id"`
	GroupID int64     `xorm:"bigint not null index unique(task_work_group_assignee)" json:"group_id"`
	Created time.Time `xorm:"created not null" json:"created" readOnly:"true"`
}

func (*TaskWorkGroupAssignee) TableName() string { return "task_work_group_assignees" }

// TaskAssigneeSource records why a materialized task_assignees row exists.
// A user can be assigned directly and through one or more work groups at the
// same time; the task_assignees row is removed only when the last source goes.
type TaskAssigneeSource struct {
	ID         int64     `xorm:"bigint autoincr not null unique pk" json:"id"`
	TaskID     int64     `xorm:"bigint not null index unique(task_assignee_source)" json:"task_id"`
	UserID     int64     `xorm:"bigint not null index unique(task_assignee_source)" json:"user_id"`
	SourceType string    `xorm:"varchar(30) not null unique(task_assignee_source)" json:"source_type"`
	SourceID   int64     `xorm:"bigint not null default 0 unique(task_assignee_source)" json:"source_id"`
	Created    time.Time `xorm:"created not null" json:"created" readOnly:"true"`
}

func (*TaskAssigneeSource) TableName() string { return "task_assignee_sources" }

type WorkGroupView struct {
	WorkGroup
	Leader      *user.User   `json:"leader,omitempty"`
	Members     []*user.User `json:"members"`
	MemberCount int          `json:"member_count"`
	TaskCount   int64        `json:"task_count"`
}

type WorkGroupSummary struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	LeaderUserID int64  `json:"leader_user_id"`
}

type taskWorkGroupSummaryRow struct {
	TaskID       int64  `xorm:"'task_id'"`
	ID           int64  `xorm:"'id'"`
	Name         string `xorm:"'name'"`
	LeaderUserID int64  `xorm:"'leader_user_id'"`
}

func addWorkGroupsToTasks(s *xorm.Session, taskIDs []int64, taskMap map[int64]*Task) error {
	if len(taskIDs) == 0 {
		return nil
	}
	rows := []*taskWorkGroupSummaryRow{}
	err := s.Table("task_work_group_assignees").
		Select("task_work_group_assignees.task_id AS task_id, work_groups.id AS id, work_groups.name AS name, work_groups.leader_user_id AS leader_user_id").
		Join("INNER", "work_groups", "task_work_group_assignees.group_id = work_groups.id").
		In("task_work_group_assignees.task_id", taskIDs).
		OrderBy("work_groups.name ASC").
		Find(&rows)
	if err != nil {
		return err
	}
	for _, row := range rows {
		task, ok := taskMap[row.TaskID]
		if !ok {
			continue
		}
		task.WorkGroups = append(task.WorkGroups, WorkGroupSummary{ID: row.ID, Name: row.Name, LeaderUserID: row.LeaderUserID})
	}
	return nil
}

type WorkGroupTaskAssignmentResult struct {
	Group         WorkGroupView `json:"group"`
	AssignedUsers []*user.User  `json:"assigned_users"`
	SkippedUsers  []*user.User  `json:"skipped_users"`
}

func normalizeWorkGroupMemberIDs(memberIDs []int64, leaderUserID int64) []int64 {
	seen := make(map[int64]struct{}, len(memberIDs)+1)
	out := make([]int64, 0, len(memberIDs)+1)
	for _, id := range memberIDs {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	if leaderUserID > 0 {
		if _, ok := seen[leaderUserID]; !ok {
			out = append(out, leaderUserID)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func workGroupMembers(s *xorm.Session, groupID int64) ([]*user.User, error) {
	members := []*user.User{}
	err := s.Table("work_group_members").
		Select("users.*").
		Join("INNER", "users", "work_group_members.user_id = users.id").
		Where("work_group_members.group_id = ?", groupID).
		OrderBy("users.name ASC, users.username ASC").
		Find(&members)
	if err != nil {
		return nil, err
	}
	for _, member := range members {
		member.Email = ""
	}
	return members, nil
}

func workGroupView(s *xorm.Session, group *WorkGroup) (*WorkGroupView, error) {
	members, err := workGroupMembers(s, group.ID)
	if err != nil {
		return nil, err
	}
	var leader *user.User
	if group.LeaderUserID > 0 {
		leaders, lookupErr := user.GetUsersByIDs(s, []int64{group.LeaderUserID})
		if lookupErr != nil {
			return nil, lookupErr
		}
		leader = leaders[group.LeaderUserID]
	}
	taskCount, err := s.Where("group_id = ?", group.ID).Count(&TaskWorkGroupAssignee{})
	if err != nil {
		return nil, err
	}
	return &WorkGroupView{
		WorkGroup:   *group,
		Leader:      leader,
		Members:     members,
		MemberCount: len(members),
		TaskCount:   taskCount,
	}, nil
}

func GetWorkGroup(s *xorm.Session, id int64) (*WorkGroupView, error) {
	group := &WorkGroup{ID: id}
	has, err := s.Get(group)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrInvalidData{Message: "work group not found"}
	}
	return workGroupView(s, group)
}

func ListWorkGroups(s *xorm.Session, search string) ([]WorkGroupView, error) {
	groups := []*WorkGroup{}
	query := s.OrderBy("name ASC")
	if strings.TrimSpace(search) != "" {
		query = query.Where(db.ILIKE("name", strings.TrimSpace(search)))
	}
	if err := query.Find(&groups); err != nil {
		return nil, err
	}
	out := make([]WorkGroupView, 0, len(groups))
	for _, group := range groups {
		view, err := workGroupView(s, group)
		if err != nil {
			return nil, err
		}
		out = append(out, *view)
	}
	return out, nil
}

func validateWorkGroupUsers(s *xorm.Session, memberIDs []int64) error {
	if len(memberIDs) == 0 {
		return nil
	}
	usersByID, err := user.GetUsersByIDs(s, memberIDs)
	if err != nil {
		return err
	}
	for _, id := range memberIDs {
		if _, ok := usersByID[id]; !ok {
			return user.ErrUserDoesNotExist{UserID: id}
		}
	}
	return nil
}

func setWorkGroupMembers(s *xorm.Session, groupID int64, memberIDs []int64) error {
	if _, err := s.Where("group_id = ?", groupID).Delete(&WorkGroupMember{}); err != nil {
		return err
	}
	for _, userID := range memberIDs {
		if _, err := s.Insert(&WorkGroupMember{GroupID: groupID, UserID: userID}); err != nil {
			return err
		}
	}
	return nil
}

func workGroupMemberIDSet(s *xorm.Session, groupID int64) (map[int64]struct{}, error) {
	rows := []*WorkGroupMember{}
	if err := s.Where("group_id = ?", groupID).Find(&rows); err != nil {
		return nil, err
	}
	out := make(map[int64]struct{}, len(rows))
	for _, row := range rows {
		out[row.UserID] = struct{}{}
	}
	return out, nil
}

func workGroupAssignedTaskIDs(s *xorm.Session, groupID int64) ([]int64, error) {
	rows := []*TaskWorkGroupAssignee{}
	if err := s.Where("group_id = ?", groupID).Find(&rows); err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.TaskID)
	}
	return ids, nil
}

func syncAddedWorkGroupMember(s *xorm.Session, groupID, userID int64, a web.Auth) error {
	taskIDs, err := workGroupAssignedTaskIDs(s, groupID)
	if err != nil {
		return err
	}
	for _, taskID := range taskIDs {
		project, err := GetProjectSimpleByTaskID(s, taskID)
		if err != nil {
			return err
		}
		task := &Task{ID: taskID, ProjectID: project.ID}
		err = task.addNewAssigneeByIDWithSource(s, userID, project, a, TaskAssigneeSourceWorkGroup, groupID, false)
		if IsErrUserDoesNotHaveAccessToProject(err) || user.IsErrAccountDisabled(err) || user.IsErrAccountLocked(err) {
			// Group membership never grants project permissions. Ineligible or
			// inactive users stay members of the group but are not materialized as
			// task assignees.
			continue
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func syncRemovedWorkGroupMember(s *xorm.Session, groupID, userID int64, a web.Auth) error {
	taskIDs, err := workGroupAssignedTaskIDs(s, groupID)
	if err != nil {
		return err
	}
	for _, taskID := range taskIDs {
		if err := removeTaskAssigneeSourceAndMaybeMaterialized(s, taskID, userID, TaskAssigneeSourceWorkGroup, groupID, a); err != nil {
			return err
		}
	}
	return nil
}

func CreateWorkGroup(s *xorm.Session, name, description string, leaderUserID int64, memberIDs []int64, a web.Auth) (*WorkGroupView, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidData{Message: "work group name is required"}
	}
	members := normalizeWorkGroupMemberIDs(memberIDs, leaderUserID)
	if err := validateWorkGroupUsers(s, members); err != nil {
		return nil, err
	}
	group := &WorkGroup{Name: name, Description: strings.TrimSpace(description), LeaderUserID: leaderUserID}
	if _, err := s.Insert(group); err != nil {
		return nil, err
	}
	if err := setWorkGroupMembers(s, group.ID, members); err != nil {
		return nil, err
	}
	return workGroupView(s, group)
}

func UpdateWorkGroup(s *xorm.Session, id int64, name, description *string, leaderUserID *int64, memberIDs *[]int64, a web.Auth) (*WorkGroupView, error) {
	group := &WorkGroup{ID: id}
	has, err := s.Get(group)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrInvalidData{Message: "work group not found"}
	}
	oldMembers, err := workGroupMemberIDSet(s, id)
	if err != nil {
		return nil, err
	}

	if name != nil {
		trimmed := strings.TrimSpace(*name)
		if trimmed == "" {
			return nil, ErrInvalidData{Message: "work group name is required"}
		}
		group.Name = trimmed
	}
	if description != nil {
		group.Description = strings.TrimSpace(*description)
	}
	if leaderUserID != nil {
		group.LeaderUserID = *leaderUserID
	}

	newMembers := make([]int64, 0, len(oldMembers)+1)
	if memberIDs != nil {
		newMembers = normalizeWorkGroupMemberIDs(*memberIDs, group.LeaderUserID)
	} else {
		for id := range oldMembers {
			newMembers = append(newMembers, id)
		}
		newMembers = normalizeWorkGroupMemberIDs(newMembers, group.LeaderUserID)
	}
	if err := validateWorkGroupUsers(s, newMembers); err != nil {
		return nil, err
	}

	if _, err := s.ID(group.ID).Cols("name", "description", "leader_user_id").Update(group); err != nil {
		return nil, err
	}
	if err := setWorkGroupMembers(s, group.ID, newMembers); err != nil {
		return nil, err
	}

	newSet := make(map[int64]struct{}, len(newMembers))
	for _, userID := range newMembers {
		newSet[userID] = struct{}{}
		if _, existed := oldMembers[userID]; !existed {
			if err := syncAddedWorkGroupMember(s, group.ID, userID, a); err != nil {
				return nil, err
			}
		}
	}
	for userID := range oldMembers {
		if _, stillMember := newSet[userID]; stillMember {
			continue
		}
		if err := syncRemovedWorkGroupMember(s, group.ID, userID, a); err != nil {
			return nil, err
		}
	}

	return workGroupView(s, group)
}

func DeleteWorkGroup(s *xorm.Session, id int64, a web.Auth) error {
	if _, err := GetWorkGroup(s, id); err != nil {
		return err
	}
	taskIDs, err := workGroupAssignedTaskIDs(s, id)
	if err != nil {
		return err
	}
	for _, taskID := range taskIDs {
		sources := []*TaskAssigneeSource{}
		if err := s.Where("task_id = ? AND source_type = ? AND source_id = ?", taskID, TaskAssigneeSourceWorkGroup, id).Find(&sources); err != nil {
			return err
		}
		for _, source := range sources {
			if err := removeTaskAssigneeSourceAndMaybeMaterialized(s, taskID, source.UserID, TaskAssigneeSourceWorkGroup, id, a); err != nil {
				return err
			}
		}
		if _, err := s.Where("task_id = ? AND group_id = ?", taskID, id).Delete(&TaskWorkGroupAssignee{}); err != nil {
			return err
		}
	}
	if _, err := s.Where("group_id = ?", id).Delete(&WorkGroupMember{}); err != nil {
		return err
	}
	_, err = s.ID(id).Delete(&WorkGroup{})
	return err
}

func ListTaskWorkGroupAssignees(s *xorm.Session, taskID int64, a web.Auth) ([]WorkGroupView, error) {
	project, err := GetProjectSimpleByTaskID(s, taskID)
	if err != nil {
		return nil, err
	}
	can, _, err := project.CanRead(s, a)
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, ErrGenericForbidden{}
	}
	rows := []*TaskWorkGroupAssignee{}
	if err := s.Where("task_id = ?", taskID).OrderBy("id ASC").Find(&rows); err != nil {
		return nil, err
	}
	out := make([]WorkGroupView, 0, len(rows))
	for _, row := range rows {
		view, err := GetWorkGroup(s, row.GroupID)
		if err != nil {
			return nil, err
		}
		out = append(out, *view)
	}
	return out, nil
}

func AssignWorkGroupToTask(s *xorm.Session, taskID, groupID int64, a web.Auth) (*WorkGroupTaskAssignmentResult, error) {
	can, err := (&TaskAssginee{TaskID: taskID}).CanCreate(s, a)
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, ErrGenericForbidden{}
	}
	group, err := GetWorkGroup(s, groupID)
	if err != nil {
		return nil, err
	}
	exists, err := s.Where("task_id = ? AND group_id = ?", taskID, groupID).Exist(&TaskWorkGroupAssignee{})
	if err != nil {
		return nil, err
	}
	if !exists {
		if _, err := s.Insert(&TaskWorkGroupAssignee{TaskID: taskID, GroupID: groupID}); err != nil {
			return nil, err
		}
	}

	project, err := GetProjectSimpleByTaskID(s, taskID)
	if err != nil {
		return nil, err
	}
	assigned := make([]*user.User, 0, len(group.Members))
	skipped := make([]*user.User, 0)
	task := &Task{ID: taskID, ProjectID: project.ID}
	for _, member := range group.Members {
		err := task.addNewAssigneeByIDWithSource(s, member.ID, project, a, TaskAssigneeSourceWorkGroup, groupID, false)
		if IsErrUserDoesNotHaveAccessToProject(err) || user.IsErrAccountDisabled(err) || user.IsErrAccountLocked(err) {
			skipped = append(skipped, member)
			continue
		}
		if err != nil {
			return nil, err
		}
		assigned = append(assigned, member)
	}
	return &WorkGroupTaskAssignmentResult{Group: *group, AssignedUsers: assigned, SkippedUsers: skipped}, nil
}

func UnassignWorkGroupFromTask(s *xorm.Session, taskID, groupID int64, a web.Auth) error {
	can, err := (&TaskAssginee{TaskID: taskID}).CanDelete(s, a)
	if err != nil {
		return err
	}
	if !can {
		return ErrGenericForbidden{}
	}
	members := []*TaskAssigneeSource{}
	if err := s.Where("task_id = ? AND source_type = ? AND source_id = ?", taskID, TaskAssigneeSourceWorkGroup, groupID).Find(&members); err != nil {
		return err
	}
	for _, source := range members {
		if err := removeTaskAssigneeSourceAndMaybeMaterialized(s, taskID, source.UserID, TaskAssigneeSourceWorkGroup, groupID, a); err != nil {
			return err
		}
	}
	_, err = s.Where("task_id = ? AND group_id = ?", taskID, groupID).Delete(&TaskWorkGroupAssignee{})
	return err
}
