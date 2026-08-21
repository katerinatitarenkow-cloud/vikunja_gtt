// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package models

import (
	"strings"
	"time"

	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// ClientCustomField is one user-defined name/value pair attached to a client project.
type ClientCustomField struct {
	ID        int64     `xorm:"bigint autoincr not null unique pk" json:"id" readOnly:"true"`
	ProjectID int64     `xorm:"bigint not null index" json:"project_id" readOnly:"true"`
	Name      string    `xorm:"varchar(500) not null" json:"name" maxLength:"500"`
	Value     string    `xorm:"text null" json:"value"`
	Position  int       `xorm:"int not null default 0 index" json:"position" readOnly:"true"`
	Created   time.Time `xorm:"created not null" json:"created" readOnly:"true"`
	Updated   time.Time `xorm:"updated not null" json:"updated" readOnly:"true"`
}

func (*ClientCustomField) TableName() string { return "client_custom_fields" }

func (f *ClientCustomField) loadProjectID(s *xorm.Session) error {
	if f.ProjectID > 0 {
		return nil
	}
	if f.ID <= 0 {
		return ErrInvalidData{Message: "custom field requires project id"}
	}
	stored := &ClientCustomField{ID: f.ID}
	has, err := s.ID(f.ID).Get(stored)
	if err != nil {
		return err
	}
	if !has {
		return ErrGenericForbidden{}
	}
	f.ProjectID = stored.ProjectID
	return nil
}

func (f *ClientCustomField) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	if _, isLinkShare := a.(*LinkSharing); isLinkShare {
		return false, 0, nil
	}
	if err := f.loadProjectID(s); err != nil {
		return false, 0, err
	}
	return (&Project{ID: f.ProjectID}).CanRead(s, a)
}

func (f *ClientCustomField) canWrite(s *xorm.Session, a web.Auth) (bool, error) {
	if _, isLinkShare := a.(*LinkSharing); isLinkShare {
		return false, nil
	}
	if err := f.loadProjectID(s); err != nil {
		return false, err
	}
	return (&Project{ID: f.ProjectID}).CanUpdate(s, a)
}

func (f *ClientCustomField) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	return f.canWrite(s, a)
}

func (f *ClientCustomField) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	return f.canWrite(s, a)
}

func (f *ClientCustomField) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	return f.canWrite(s, a)
}

func GetClientCustomFields(s *xorm.Session, projectID int64) ([]*ClientCustomField, error) {
	fields := []*ClientCustomField{}
	err := s.Where("project_id = ?", projectID).OrderBy("position asc, id asc").Find(&fields)
	return fields, err
}

func validateClientCustomField(field *ClientCustomField) error {
	field.Name = strings.TrimSpace(field.Name)
	field.Value = strings.TrimSpace(field.Value)
	if field.Name == "" {
		return ErrInvalidData{Message: "custom field name is required"}
	}
	if len([]rune(field.Name)) > 500 {
		return ErrInvalidData{Message: "custom field name is too long"}
	}
	return nil
}

func nextClientCustomFieldPosition(s *xorm.Session, projectID int64) (int, error) {
	last := &ClientCustomField{}
	has, err := s.Where("project_id = ?", projectID).OrderBy("position desc, id desc").Get(last)
	if err != nil {
		return 0, err
	}
	if !has {
		return 0, nil
	}
	return last.Position + 1, nil
}

func CreateClientCustomField(s *xorm.Session, a web.Auth, projectID int64, name, value string) (*ClientCustomField, error) {
	field := &ClientCustomField{ProjectID: projectID, Name: name, Value: value}
	if err := validateClientCustomField(field); err != nil {
		return nil, err
	}
	can, err := field.CanCreate(s, a)
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, ErrGenericForbidden{}
	}
	field.Position, err = nextClientCustomFieldPosition(s, projectID)
	if err != nil {
		return nil, err
	}
	if _, err := s.Insert(field); err != nil {
		return nil, err
	}
	if err := CreateSystemClientActivity(s, projectID, ClientActivityCustomFieldCreated, doerFromAuth(s, a), "custom_field", field.ID, "", &ClientActivityMetadata{
		FieldName: field.Name,
		NewValue:  field.Value,
	}); err != nil {
		return nil, err
	}
	return field, nil
}

func UpdateClientCustomField(s *xorm.Session, a web.Auth, projectID, fieldID int64, name, value string) (*ClientCustomField, bool, error) {
	stored := &ClientCustomField{}
	has, err := s.Where("id = ? AND project_id = ?", fieldID, projectID).Get(stored)
	if err != nil || !has {
		return nil, has, err
	}
	can, err := stored.CanUpdate(s, a)
	if err != nil {
		return nil, true, err
	}
	if !can {
		return nil, true, ErrGenericForbidden{}
	}

	oldName := stored.Name
	oldValue := stored.Value
	stored.Name = name
	stored.Value = value
	if err := validateClientCustomField(stored); err != nil {
		return nil, true, err
	}
	if stored.Name == oldName && stored.Value == oldValue {
		return stored, true, nil
	}
	stored.Updated = time.Now()
	if _, err := s.ID(fieldID).Cols("name", "value", "updated").Update(stored); err != nil {
		return nil, true, err
	}
	metadata := &ClientActivityMetadata{FieldName: stored.Name}
	if oldName != stored.Name {
		metadata.OldFieldName = oldName
		metadata.NewFieldName = stored.Name
	}
	if oldValue != stored.Value {
		metadata.OldValue = oldValue
		metadata.NewValue = stored.Value
	}
	if err := CreateSystemClientActivity(s, projectID, ClientActivityCustomFieldUpdated, doerFromAuth(s, a), "custom_field", fieldID, "", metadata); err != nil {
		return nil, true, err
	}
	return stored, true, nil
}

func DeleteClientCustomField(s *xorm.Session, a web.Auth, projectID, fieldID int64) (bool, error) {
	stored := &ClientCustomField{}
	has, err := s.Where("id = ? AND project_id = ?", fieldID, projectID).Get(stored)
	if err != nil || !has {
		return has, err
	}
	can, err := stored.CanDelete(s, a)
	if err != nil {
		return true, err
	}
	if !can {
		return true, ErrGenericForbidden{}
	}
	if _, err := s.ID(fieldID).Delete(&ClientCustomField{}); err != nil {
		return true, err
	}
	if err := CreateSystemClientActivity(s, projectID, ClientActivityCustomFieldDeleted, doerFromAuth(s, a), "custom_field", fieldID, "", &ClientActivityMetadata{
		FieldName: stored.Name,
		OldValue:  stored.Value,
	}); err != nil {
		return true, err
	}
	return true, nil
}
