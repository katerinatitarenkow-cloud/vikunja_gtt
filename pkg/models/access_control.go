// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package models

import (
	"errors"
	"sort"
	"strings"

	"code.vikunja.io/api/pkg/user"
	"xorm.io/xorm"
)

const (
	SystemGroupAdmin = "admin"
	SystemGroupUsers = "users"

	PermissionProjectsView   = "projects.view"
	PermissionProjectsManage = "projects.manage"
	PermissionTasksView      = "tasks.view"
	PermissionTasksManage    = "tasks.manage"
	PermissionLabelsView     = "labels.view"
	PermissionLabelsManage   = "labels.manage"
	PermissionTeamsView      = "teams.view"
	PermissionTeamsManage    = "teams.manage"
	PermissionKanbanUse      = "kanban.use"
	PermissionTimeTracking   = "time_tracking.use"
	PermissionWialonView     = "wialon.view"
)

// PermissionDefinition is the stable global capability catalogue used by the
// admin UI. Project-level Vikunja permissions continue to apply independently.
type PermissionDefinition struct {
	Key      string `json:"key"`
	Category string `json:"category"`
}

var accessPermissionDefinitions = []PermissionDefinition{
	{Key: PermissionProjectsView, Category: "projects"},
	{Key: PermissionProjectsManage, Category: "projects"},
	{Key: PermissionTasksView, Category: "tasks"},
	{Key: PermissionTasksManage, Category: "tasks"},
	{Key: PermissionLabelsView, Category: "labels"},
	{Key: PermissionLabelsManage, Category: "labels"},
	{Key: PermissionTeamsView, Category: "teams"},
	{Key: PermissionTeamsManage, Category: "teams"},
	{Key: PermissionKanbanUse, Category: "kanban"},
	{Key: PermissionTimeTracking, Category: "time_tracking"},
	{Key: PermissionWialonView, Category: "wialon"},
}

func AccessPermissionDefinitions() []PermissionDefinition {
	out := make([]PermissionDefinition, len(accessPermissionDefinitions))
	copy(out, accessPermissionDefinitions)
	return out
}

func AllAccessPermissionKeys() []string {
	out := make([]string, 0, len(accessPermissionDefinitions))
	for _, definition := range accessPermissionDefinitions {
		out = append(out, definition.Key)
	}
	return out
}

// DefaultUserPermissionKeys deliberately preserve the pre-RBAC behaviour for
// existing and newly-created regular users. Admins can then narrow access by
// editing the Users group or moving a user to a custom group.
func DefaultUserPermissionKeys() []string { return AllAccessPermissionKeys() }

func validAccessPermissionSet() map[string]struct{} {
	valid := make(map[string]struct{}, len(accessPermissionDefinitions))
	for _, definition := range accessPermissionDefinitions {
		valid[definition.Key] = struct{}{}
	}
	return valid
}

// AccessGroup is a global feature-permission group, separate from Vikunja's
// project-sharing Team model.
type AccessGroup struct {
	ID          int64  `xorm:"bigint autoincr not null unique pk" json:"id"`
	Name        string `xorm:"varchar(250) not null unique" json:"name"`
	Description string `xorm:"text null" json:"description"`
	SystemKey   string `xorm:"varchar(50) null index" json:"system_key,omitempty"`
}

func (*AccessGroup) TableName() string { return "access_groups" }

// AccessGroupPermission stores one capability per row so permissions remain
// queryable and easy to extend without a schema migration.
type AccessGroupPermission struct {
	ID         int64  `xorm:"bigint autoincr not null unique pk" json:"id"`
	GroupID    int64  `xorm:"bigint not null index unique(access_group_permission)" json:"group_id"`
	Permission string `xorm:"varchar(100) not null unique(access_group_permission)" json:"permission"`
}

func (*AccessGroupPermission) TableName() string { return "access_group_permissions" }

// AccessGroupMember is intentionally many-to-many: a person may combine a
// department group with a role group. Effective permissions are the union.
type AccessGroupMember struct {
	ID      int64 `xorm:"bigint autoincr not null unique pk" json:"id"`
	GroupID int64 `xorm:"bigint not null index unique(access_group_member)" json:"group_id"`
	UserID  int64 `xorm:"bigint not null index unique(access_group_member)" json:"user_id"`
}

func (*AccessGroupMember) TableName() string { return "access_group_members" }

// UserProfile contains admin-maintained personnel-card fields that should not
// be exposed by the normal /user endpoint.
type UserProfile struct {
	UserID int64  `xorm:"bigint not null unique pk" json:"user_id"`
	Phone  string `xorm:"varchar(100) null" json:"phone"`
	Notes  string `xorm:"text null" json:"notes"`
}

func (*UserProfile) TableName() string { return "user_profiles" }

// WialonSettings is the UI-managed server-side connection configuration. The
// token is never returned by public/read APIs.
type WialonSettings struct {
	ID             int64  `xorm:"bigint not null unique pk" json:"-"`
	Enabled        bool   `xorm:"not null default false" json:"enabled"`
	APIURL         string `xorm:"varchar(500) not null" json:"api_url"`
	Token          string `xorm:"text null" json:"-"`
	TimeoutSeconds int    `xorm:"not null default 30" json:"timeout_seconds"`
	TrackMaxPoints int    `xorm:"not null default 5000" json:"track_max_points"`
}

func (*WialonSettings) TableName() string { return "wialon_settings" }

type AccessGroupView struct {
	AccessGroup
	Permissions []string `json:"permissions"`
	MemberCount int64    `json:"member_count"`
}

type AccessUserView struct {
	ID       int64             `json:"id"`
	Username string            `json:"username"`
	Name     string            `json:"name"`
	Email    string            `json:"email"`
	Phone    string            `json:"phone"`
	Notes    string            `json:"notes"`
	IsAdmin  bool              `json:"is_admin"`
	Status   user.Status       `json:"status"`
	Created  int64             `json:"created"`
	Updated  int64             `json:"updated"`
	Groups   []AccessGroupView `json:"groups"`
}

func normalizePermissions(permissions []string) ([]string, error) {
	valid := validAccessPermissionSet()
	seen := map[string]struct{}{}
	out := make([]string, 0, len(permissions)+4)
	for _, permission := range permissions {
		permission = strings.TrimSpace(permission)
		if permission == "" {
			continue
		}
		if _, ok := valid[permission]; !ok {
			return nil, ErrInvalidData{Message: "unknown access permission: " + permission}
		}
		if _, ok := seen[permission]; ok {
			continue
		}
		seen[permission] = struct{}{}
		out = append(out, permission)
	}
	// Manage permissions imply the corresponding read permission so a group
	// cannot be configured into an unusable write-without-read state.
	dependencies := map[string]string{
		PermissionProjectsManage: PermissionProjectsView,
		PermissionTasksManage:    PermissionTasksView,
		PermissionLabelsManage:   PermissionLabelsView,
		PermissionTeamsManage:    PermissionTeamsView,
	}
	for manage, view := range dependencies {
		if _, hasManage := seen[manage]; hasManage {
			if _, hasView := seen[view]; !hasView {
				seen[view] = struct{}{}
				out = append(out, view)
			}
		}
	}
	sort.Strings(out)
	return out, nil
}

func getSystemAccessGroup(s *xorm.Session, key string) (*AccessGroup, error) {
	group := &AccessGroup{}
	has, err := s.Where("system_key = ?", key).Get(group)
	if err != nil {
		return nil, err
	}
	if has {
		return group, nil
	}

	name := "Пользователи"
	description := "Обычные пользователи"
	if key == SystemGroupAdmin {
		name = "Админ"
		description = "Системные администраторы"
	} else if key != SystemGroupUsers {
		return nil, errors.New("unknown system access group: " + key)
	}
	group = &AccessGroup{Name: name, Description: description, SystemKey: key}
	if _, err := s.Insert(group); err != nil {
		return nil, err
	}
	if err := SetAccessGroupPermissions(s, group.ID, AllAccessPermissionKeys()); err != nil {
		return nil, err
	}
	return group, nil
}

func groupPermissions(s *xorm.Session, groupID int64) ([]string, error) {
	var rows []*AccessGroupPermission
	if err := s.Where("group_id = ?", groupID).OrderBy("permission ASC").Find(&rows); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.Permission)
	}
	return out, nil
}

func ListAccessGroups(s *xorm.Session) ([]AccessGroupView, error) {
	var groups []*AccessGroup
	if err := s.OrderBy("id ASC").Find(&groups); err != nil {
		return nil, err
	}
	out := make([]AccessGroupView, 0, len(groups))
	for _, group := range groups {
		permissions, err := groupPermissions(s, group.ID)
		if err != nil {
			return nil, err
		}
		count, err := s.Where("group_id = ?", group.ID).Count(&AccessGroupMember{})
		if err != nil {
			return nil, err
		}
		out = append(out, AccessGroupView{AccessGroup: *group, Permissions: permissions, MemberCount: count})
	}
	return out, nil
}

func SetAccessGroupPermissions(s *xorm.Session, groupID int64, permissions []string) error {
	normalized, err := normalizePermissions(permissions)
	if err != nil {
		return err
	}
	if _, err := s.Where("group_id = ?", groupID).Delete(&AccessGroupPermission{}); err != nil {
		return err
	}
	for _, permission := range normalized {
		if _, err := s.Insert(&AccessGroupPermission{GroupID: groupID, Permission: permission}); err != nil {
			return err
		}
	}
	return nil
}

func CreateAccessGroup(s *xorm.Session, name, description string, permissions []string) (*AccessGroupView, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidData{Message: "group name is required"}
	}
	group := &AccessGroup{Name: name, Description: strings.TrimSpace(description)}
	if _, err := s.Insert(group); err != nil {
		return nil, err
	}
	if err := SetAccessGroupPermissions(s, group.ID, permissions); err != nil {
		return nil, err
	}
	perms, err := groupPermissions(s, group.ID)
	if err != nil {
		return nil, err
	}
	return &AccessGroupView{AccessGroup: *group, Permissions: perms}, nil
}

func UpdateAccessGroup(s *xorm.Session, id int64, name, description *string, permissions *[]string) (*AccessGroupView, error) {
	group := &AccessGroup{ID: id}
	has, err := s.Get(group)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrInvalidData{Message: "access group does not exist"}
	}

	if group.SystemKey == SystemGroupAdmin {
		// The instance-admin group is a display/synchronisation group. Actual
		// admin access is still protected by user.is_admin and cannot be weakened.
		all := AllAccessPermissionKeys()
		permissions = &all
	}
	cols := make([]string, 0, 2)
	if name != nil && group.SystemKey == "" {
		trimmed := strings.TrimSpace(*name)
		if trimmed == "" {
			return nil, ErrInvalidData{Message: "group name is required"}
		}
		group.Name = trimmed
		cols = append(cols, "name")
	}
	if description != nil {
		group.Description = strings.TrimSpace(*description)
		cols = append(cols, "description")
	}
	if len(cols) > 0 {
		if _, err := s.ID(group.ID).Cols(cols...).Update(group); err != nil {
			return nil, err
		}
	}
	if permissions != nil {
		if err := SetAccessGroupPermissions(s, group.ID, *permissions); err != nil {
			return nil, err
		}
	}
	perms, err := groupPermissions(s, group.ID)
	if err != nil {
		return nil, err
	}
	count, err := s.Where("group_id = ?", group.ID).Count(&AccessGroupMember{})
	if err != nil {
		return nil, err
	}
	return &AccessGroupView{AccessGroup: *group, Permissions: perms, MemberCount: count}, nil
}

func DeleteAccessGroup(s *xorm.Session, id int64) error {
	group := &AccessGroup{ID: id}
	has, err := s.Get(group)
	if err != nil {
		return err
	}
	if !has {
		return ErrInvalidData{Message: "access group does not exist"}
	}
	if group.SystemKey != "" {
		return ErrInvalidData{Message: "system access groups cannot be deleted"}
	}

	var affected []*AccessGroupMember
	if err := s.Where("group_id = ?", id).Find(&affected); err != nil {
		return err
	}
	if _, err := s.Where("group_id = ?", id).Delete(&AccessGroupMember{}); err != nil {
		return err
	}
	if _, err := s.Where("group_id = ?", id).Delete(&AccessGroupPermission{}); err != nil {
		return err
	}
	if _, err := s.ID(id).Delete(&AccessGroup{}); err != nil {
		return err
	}

	// Never leave a regular user with an accidental empty permission set after
	// a custom group is deleted. Users who no longer belong to any group fall
	// back to the system Users group immediately.
	for _, membership := range affected {
		remaining, err := s.Where("user_id = ?", membership.UserID).Exist(&AccessGroupMember{})
		if err != nil {
			return err
		}
		if remaining {
			continue
		}
		u := &user.User{ID: membership.UserID}
		has, err := s.Get(u)
		if err != nil {
			return err
		}
		if has {
			if err := AssignDefaultAccessGroup(s, u); err != nil {
				return err
			}
		}
	}
	return nil
}

func UserAccessGroups(s *xorm.Session, userID int64) ([]AccessGroupView, error) {
	var memberships []*AccessGroupMember
	if err := s.Where("user_id = ?", userID).Find(&memberships); err != nil {
		return nil, err
	}
	out := make([]AccessGroupView, 0, len(memberships))
	for _, membership := range memberships {
		group := &AccessGroup{ID: membership.GroupID}
		has, err := s.Get(group)
		if err != nil {
			return nil, err
		}
		if !has {
			continue
		}
		permissions, err := groupPermissions(s, group.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, AccessGroupView{AccessGroup: *group, Permissions: permissions})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func UserAccessPermissions(s *xorm.Session, u *user.User) ([]string, error) {
	if u.IsAdmin {
		return AllAccessPermissionKeys(), nil
	}
	groups, err := UserAccessGroups(s, u.ID)
	if err != nil {
		return nil, err
	}
	if len(groups) == 0 {
		if err := AssignDefaultAccessGroup(s, u); err != nil {
			return nil, err
		}
		groups, err = UserAccessGroups(s, u.ID)
		if err != nil {
			return nil, err
		}
	}
	seen := map[string]struct{}{}
	out := make([]string, 0)
	for _, group := range groups {
		for _, permission := range group.Permissions {
			if _, ok := seen[permission]; ok {
				continue
			}
			seen[permission] = struct{}{}
			out = append(out, permission)
		}
	}
	sort.Strings(out)
	return out, nil
}

func UserHasAccessPermission(s *xorm.Session, u *user.User, permission string) (bool, error) {
	if u.IsAdmin {
		return true, nil
	}
	permissions, err := UserAccessPermissions(s, u)
	if err != nil {
		return false, err
	}
	for _, p := range permissions {
		if p == permission {
			return true, nil
		}
	}
	return false, nil
}

// AssignDefaultAccessGroup keeps compatibility with normal Vikunja account
// creation while giving administrators an immediately editable group model.
func AssignDefaultAccessGroup(s *xorm.Session, u *user.User) error {
	key := SystemGroupUsers
	if u.IsAdmin {
		key = SystemGroupAdmin
	}
	group, err := getSystemAccessGroup(s, key)
	if err != nil {
		return err
	}
	existing, err := s.Where("user_id = ? AND group_id = ?", u.ID, group.ID).Exist(&AccessGroupMember{})
	if err != nil {
		return err
	}
	if existing {
		return nil
	}
	_, err = s.Insert(&AccessGroupMember{UserID: u.ID, GroupID: group.ID})
	return err
}

// SyncAdminAccessGroup mirrors the legacy is_admin flag to the system Admin
// group without using the group itself as an alternative privilege-escalation path.
func SyncAdminAccessGroup(s *xorm.Session, u *user.User) error {
	adminGroup, err := getSystemAccessGroup(s, SystemGroupAdmin)
	if err != nil {
		return err
	}
	usersGroup, err := getSystemAccessGroup(s, SystemGroupUsers)
	if err != nil {
		return err
	}
	if u.IsAdmin {
		if _, err := s.Where("user_id = ? AND group_id = ?", u.ID, usersGroup.ID).Delete(&AccessGroupMember{}); err != nil {
			return err
		}
		exists, err := s.Where("user_id = ? AND group_id = ?", u.ID, adminGroup.ID).Exist(&AccessGroupMember{})
		if err != nil {
			return err
		}
		if !exists {
			_, err = s.Insert(&AccessGroupMember{UserID: u.ID, GroupID: adminGroup.ID})
		}
		return err
	}
	if _, err := s.Where("user_id = ? AND group_id = ?", u.ID, adminGroup.ID).Delete(&AccessGroupMember{}); err != nil {
		return err
	}
	exists, err := s.Where("user_id = ?", u.ID).Exist(&AccessGroupMember{})
	if err != nil {
		return err
	}
	if !exists {
		_, err = s.Insert(&AccessGroupMember{UserID: u.ID, GroupID: usersGroup.ID})
	}
	return err
}

// SetUserAccessGroups assigns regular/custom groups. The system Admin group is
// controlled exclusively by is_admin. An empty list falls back to Users.
func SetUserAccessGroups(s *xorm.Session, u *user.User, groupIDs []int64) error {
	adminGroup, err := getSystemAccessGroup(s, SystemGroupAdmin)
	if err != nil {
		return err
	}
	usersGroup, err := getSystemAccessGroup(s, SystemGroupUsers)
	if err != nil {
		return err
	}

	desired := map[int64]struct{}{}
	for _, id := range groupIDs {
		if id < 1 || id == adminGroup.ID {
			continue
		}
		group := &AccessGroup{ID: id}
		has, err := s.Get(group)
		if err != nil {
			return err
		}
		if !has {
			return ErrInvalidData{Message: "access group does not exist"}
		}
		desired[id] = struct{}{}
	}
	if !u.IsAdmin && len(desired) == 0 {
		desired[usersGroup.ID] = struct{}{}
	}

	if _, err := s.Where("user_id = ? AND group_id <> ?", u.ID, adminGroup.ID).Delete(&AccessGroupMember{}); err != nil {
		return err
	}
	for id := range desired {
		if _, err := s.Insert(&AccessGroupMember{UserID: u.ID, GroupID: id}); err != nil {
			return err
		}
	}
	return SyncAdminAccessGroup(s, u)
}

func GetUserProfile(s *xorm.Session, userID int64) (*UserProfile, error) {
	profile := &UserProfile{UserID: userID}
	has, err := s.Get(profile)
	if err != nil {
		return nil, err
	}
	if !has {
		return &UserProfile{UserID: userID}, nil
	}
	return profile, nil
}

func SetUserProfile(s *xorm.Session, userID int64, phone, notes string) error {
	profile := &UserProfile{UserID: userID, Phone: strings.TrimSpace(phone), Notes: strings.TrimSpace(notes)}
	has, err := s.Where("user_id = ?", userID).Exist(&UserProfile{})
	if err != nil {
		return err
	}
	if has {
		_, err = s.Where("user_id = ?", userID).Cols("phone", "notes").Update(profile)
		return err
	}
	_, err = s.Insert(profile)
	return err
}

func AccessUser(s *xorm.Session, u *user.User) (*AccessUserView, error) {
	profile, err := GetUserProfile(s, u.ID)
	if err != nil {
		return nil, err
	}
	groups, err := UserAccessGroups(s, u.ID)
	if err != nil {
		return nil, err
	}
	return &AccessUserView{
		ID: u.ID, Username: u.Username, Name: u.Name, Email: u.Email,
		Phone: profile.Phone, Notes: profile.Notes, IsAdmin: u.IsAdmin,
		Status: u.Status, Created: u.Created.Unix(), Updated: u.Updated.Unix(), Groups: groups,
	}, nil
}

func ListAccessUsers(s *xorm.Session, search string) ([]AccessUserView, error) {
	query := s.OrderBy("id ASC")
	search = strings.TrimSpace(search)
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("username LIKE ? OR name LIKE ? OR email LIKE ?", like, like, like)
	}
	var users []*user.User
	if err := query.Find(&users); err != nil {
		return nil, err
	}
	out := make([]AccessUserView, 0, len(users))
	for _, u := range users {
		view, err := AccessUser(s, u)
		if err != nil {
			return nil, err
		}
		out = append(out, *view)
	}
	return out, nil
}

func LoadWialonSettings(s *xorm.Session) (*WialonSettings, bool, error) {
	settings := &WialonSettings{ID: 1}
	has, err := s.Get(settings)
	return settings, has, err
}

func SaveWialonSettings(s *xorm.Session, settings *WialonSettings) error {
	settings.ID = 1
	settings.APIURL = strings.TrimSpace(settings.APIURL)
	if settings.APIURL == "" {
		settings.APIURL = "https://hst-api.wialon.com"
	}
	if settings.TimeoutSeconds < 1 || settings.TimeoutSeconds > 300 {
		return ErrInvalidData{Message: "Wialon timeout must be between 1 and 300 seconds"}
	}
	if settings.TrackMaxPoints < 100 || settings.TrackMaxPoints > 50000 {
		return ErrInvalidData{Message: "Wialon track max points must be between 100 and 50000"}
	}
	has, err := s.ID(1).Exist(&WialonSettings{})
	if err != nil {
		return err
	}
	if has {
		_, err = s.ID(1).AllCols().Update(settings)
		return err
	}
	_, err = s.Insert(settings)
	return err
}
