// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package models

import (
	"io"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

const (
	ClientTypePerson  = "person"
	ClientTypeFOP     = "fop"
	ClientTypeCompany = "company"

	ClientStatusPotential = "potential"
	ClientStatusActive    = "active"
	ClientStatusInactive  = "inactive"
	ClientStatusVIP       = "vip"
)

var validClientTypes = map[string]struct{}{
	ClientTypePerson: {}, ClientTypeFOP: {}, ClientTypeCompany: {},
}
var validClientStatuses = map[string]struct{}{
	ClientStatusPotential: {}, ClientStatusActive: {}, ClientStatusInactive: {}, ClientStatusVIP: {},
}
var validAddressTypes = map[string]struct{}{
	"legal": {}, "actual": {}, "postal": {}, "delivery": {}, "object": {},
}

var validClientDecisionRoles = map[string]struct{}{
	"":               {},
	"leader":         {},
	"decision_maker": {},
	"technical":      {},
	"procurement":    {},
	"accountant":     {},
	"user":           {},
	"other":          {},
}

// ClientProfile is a one-to-one CRM card attached to a Vikunja project.
// Patch 11.3 completes the CRM card with legal data, addresses, contacts and proposal.
type ClientProfile struct {
	ProjectID int64 `xorm:"bigint not null unique pk" json:"project_id" readOnly:"true"`

	ClientType  string `xorm:"varchar(20) not null default 'company'" json:"client_type"`
	DisplayName string `xorm:"varchar(500) not null" json:"display_name"`
	ContactName string `xorm:"varchar(500) null" json:"contact_name"`
	Status      string `xorm:"varchar(20) not null default 'potential' index" json:"status"`
	Source      string `xorm:"varchar(50) null" json:"source"`

	ResponsibleUserID int64      `xorm:"bigint null index" json:"responsible_user_id"`
	Responsible       *user.User `xorm:"-" json:"responsible,omitempty" readOnly:"true"`

	Phone                  string `xorm:"varchar(100) null" json:"phone"`
	PhoneSecondary         string `xorm:"varchar(100) null" json:"phone_secondary"`
	Email                  string `xorm:"varchar(320) null" json:"email"`
	EmailSecondary         string `xorm:"varchar(320) null" json:"email_secondary"`
	Telegram               string `xorm:"varchar(250) null" json:"telegram"`
	Viber                  string `xorm:"varchar(250) null" json:"viber"`
	WhatsApp               string `xorm:"varchar(250) null" json:"whatsapp"`
	Website                string `xorm:"varchar(1000) null" json:"website"`
	PreferredContactMethod string `xorm:"varchar(50) null" json:"preferred_contact_method"`
	PreferredLanguage      string `xorm:"varchar(50) null" json:"preferred_language"`

	TaxID         string `xorm:"varchar(100) null" json:"tax_id"`
	LegalName     string `xorm:"varchar(500) null" json:"legal_name"`
	DirectorName  string `xorm:"varchar(500) null" json:"director_name"`
	OwnershipForm string `xorm:"varchar(250) null" json:"ownership_form"`
	Industry      string `xorm:"varchar(500) null" json:"industry"`
	EmployeeCount int    `xorm:"int null" json:"employee_count"`
	Requisites    string `xorm:"text null" json:"requisites"`
	IBAN          string `xorm:"varchar(100) null" json:"iban"`
	Bank          string `xorm:"varchar(500) null" json:"bank"`
	MFO           string `xorm:"varchar(100) null" json:"mfo"`
	VATNumber     string `xorm:"varchar(100) null" json:"vat_number"`
	TaxSystem     string `xorm:"varchar(250) null" json:"tax_system"`

	CommercialProposalFileID int64       `xorm:"bigint null index" json:"-"`
	CommercialProposal       *files.File `xorm:"-" json:"commercial_proposal,omitempty" readOnly:"true"`

	Addresses      []*ClientAddress       `xorm:"-" json:"addresses"`
	ContactPersons []*ClientContactPerson `xorm:"-" json:"contact_persons"`
	AddedAt        time.Time              `xorm:"-" json:"added_at" readOnly:"true"`
	Updated        time.Time              `xorm:"updated not null" json:"updated" readOnly:"true"`
}

func (*ClientProfile) TableName() string { return "client_profiles" }

// ClientAddress stores one of the five fixed CRM address types. The unique
// project/type pair guarantees that a project cannot accidentally get two
// competing delivery or legal addresses.
type ClientAddress struct {
	ID         int64   `xorm:"bigint autoincr not null unique pk" json:"id" readOnly:"true"`
	ProjectID  int64   `xorm:"bigint not null index unique(client_address_type)" json:"project_id" readOnly:"true"`
	Type       string  `xorm:"varchar(30) not null unique(client_address_type)" json:"type"`
	Country    string  `xorm:"varchar(250) null" json:"country"`
	Region     string  `xorm:"varchar(250) null" json:"region"`
	City       string  `xorm:"varchar(250) null" json:"city"`
	Street     string  `xorm:"varchar(500) null" json:"street"`
	House      string  `xorm:"varchar(100) null" json:"house"`
	Office     string  `xorm:"varchar(100) null" json:"office"`
	PostalCode string  `xorm:"varchar(50) null" json:"postal_code"`
	Latitude   float64 `xorm:"double null" json:"latitude"`
	Longitude  float64 `xorm:"double null" json:"longitude"`
}

func (*ClientAddress) TableName() string { return "client_addresses" }

// ClientContactPerson is a structured person connected to the client/project.
// It deliberately carries no permissions: contacts are CRM data, not Vikunja users.
type ClientContactPerson struct {
	ID        int64 `xorm:"bigint autoincr not null unique pk" json:"id" readOnly:"true"`
	ProjectID int64 `xorm:"bigint not null index" json:"project_id" readOnly:"true"`

	FullName               string `xorm:"varchar(500) not null" json:"full_name"`
	Position               string `xorm:"varchar(250) null" json:"position"`
	Department             string `xorm:"varchar(250) null" json:"department"`
	Phone                  string `xorm:"varchar(100) null" json:"phone"`
	Email                  string `xorm:"varchar(320) null" json:"email"`
	Telegram               string `xorm:"varchar(250) null" json:"telegram"`
	Viber                  string `xorm:"varchar(250) null" json:"viber"`
	WhatsApp               string `xorm:"varchar(250) null" json:"whatsapp"`
	Birthday               string `xorm:"varchar(10) null" json:"birthday"`
	PreferredContactMethod string `xorm:"varchar(50) null" json:"preferred_contact_method"`
	DecisionRole           string `xorm:"varchar(50) null index" json:"decision_role"`
	Notes                  string `xorm:"text null" json:"notes"`
	PositionIndex          int    `xorm:"int not null default 0" json:"position_index"`
}

func (*ClientContactPerson) TableName() string { return "client_contact_persons" }

func defaultClientProfile(project *Project) *ClientProfile {
	return &ClientProfile{
		ProjectID:      project.ID,
		ClientType:     ClientTypeCompany,
		DisplayName:    project.Title,
		Status:         ClientStatusPotential,
		AddedAt:        project.Created,
		Addresses:      []*ClientAddress{},
		ContactPersons: []*ClientContactPerson{},
	}
}

func GetClientProfile(s *xorm.Session, projectID int64) (*ClientProfile, error) {
	project, err := GetProjectSimpleByID(s, projectID)
	if err != nil {
		return nil, err
	}

	profile := &ClientProfile{ProjectID: projectID}
	has, err := s.ID(projectID).Get(profile)
	if err != nil {
		return nil, err
	}
	if !has {
		return defaultClientProfile(project), nil
	}

	profile.AddedAt = project.Created
	profile.Addresses = []*ClientAddress{}
	if err := s.Where("project_id = ?", projectID).OrderBy("id asc").Find(&profile.Addresses); err != nil {
		return nil, err
	}
	profile.ContactPersons = []*ClientContactPerson{}
	if err := s.Where("project_id = ?", projectID).OrderBy("position_index asc, id asc").Find(&profile.ContactPersons); err != nil {
		return nil, err
	}
	if profile.CommercialProposalFileID > 0 {
		proposal := &files.File{ID: profile.CommercialProposalFileID}
		has, err := s.ID(profile.CommercialProposalFileID).Get(proposal)
		if err != nil {
			return nil, err
		}
		if has {
			profile.CommercialProposal = proposal
		}
	}
	if profile.ResponsibleUserID > 0 {
		users, err := user.GetUsersByIDs(s, []int64{profile.ResponsibleUserID})
		if err != nil {
			return nil, err
		}
		profile.Responsible = users[profile.ResponsibleUserID]
	}
	return profile, nil
}

func validateClientProfile(profile *ClientProfile) error {
	profile.ClientType = strings.TrimSpace(profile.ClientType)
	profile.DisplayName = strings.TrimSpace(profile.DisplayName)
	profile.ContactName = strings.TrimSpace(profile.ContactName)
	profile.Status = strings.TrimSpace(profile.Status)
	profile.Source = strings.TrimSpace(profile.Source)

	if _, ok := validClientTypes[profile.ClientType]; !ok {
		return ErrInvalidData{Message: "invalid client type"}
	}
	if profile.DisplayName == "" {
		return ErrInvalidData{Message: "client name is required"}
	}
	if _, ok := validClientStatuses[profile.Status]; !ok {
		return ErrInvalidData{Message: "invalid client status"}
	}
	return nil
}

func validateClientResponsible(s *xorm.Session, projectID, userID int64) (*user.User, error) {
	if userID <= 0 {
		return nil, nil
	}
	target, err := user.GetUserByID(s, userID)
	if err != nil {
		return nil, err
	}
	canRead, _, err := (&Project{ID: projectID}).CanRead(s, target)
	if err != nil {
		return nil, err
	}
	if !canRead {
		return nil, ErrUserDoesNotHaveAccessToProject{ProjectID: projectID, UserID: userID}
	}
	return target, nil
}

func SaveClientProfile(s *xorm.Session, a web.Auth, projectID int64, input *ClientProfile) (*ClientProfile, error) {
	project, err := GetProjectSimpleByID(s, projectID)
	if err != nil {
		return nil, err
	}
	input.ProjectID = projectID
	if err := validateClientProfile(input); err != nil {
		return nil, err
	}

	responsible, err := validateClientResponsible(s, projectID, input.ResponsibleUserID)
	if err != nil {
		return nil, err
	}

	existing := &ClientProfile{ProjectID: projectID}
	has, err := s.ID(projectID).Get(existing)
	if err != nil {
		return nil, err
	}
	oldResponsibleID := int64(0)
	oldStatus := ""
	if has {
		oldResponsibleID = existing.ResponsibleUserID
		oldStatus = existing.Status
	}

	// Read-only relationship/derived fields are never persisted from request data.
	addresses := input.Addresses
	contactPersons := input.ContactPersons
	input.Responsible = nil
	input.Addresses = nil
	input.ContactPersons = nil
	input.CommercialProposal = nil
	input.AddedAt = time.Time{}
	// The proposal is managed through its dedicated upload/delete endpoint. A normal
	// profile save must never clear it just because the JSON body omits the hidden id.
	if has {
		input.CommercialProposalFileID = existing.CommercialProposalFileID
	}

	if has {
		if _, err := s.ID(projectID).AllCols().Update(input); err != nil {
			return nil, err
		}
	} else if _, err := s.Insert(input); err != nil {
		return nil, err
	}

	// Addresses are replaced as one small set (maximum five rows). Keeping this
	// atomic with the profile save makes the UI and API deterministic.
	if _, err := s.Where("project_id = ?", projectID).Delete(&ClientAddress{}); err != nil {
		return nil, err
	}
	seenAddressTypes := map[string]struct{}{}
	for _, address := range addresses {
		if address == nil {
			continue
		}
		address.Type = strings.TrimSpace(address.Type)
		if _, ok := validAddressTypes[address.Type]; !ok {
			return nil, ErrInvalidData{Message: "invalid client address type"}
		}
		if _, duplicate := seenAddressTypes[address.Type]; duplicate {
			return nil, ErrInvalidData{Message: "duplicate client address type"}
		}
		seenAddressTypes[address.Type] = struct{}{}
		address.ID = 0
		address.ProjectID = projectID
		address.Country = strings.TrimSpace(address.Country)
		address.Region = strings.TrimSpace(address.Region)
		address.City = strings.TrimSpace(address.City)
		address.Street = strings.TrimSpace(address.Street)
		address.House = strings.TrimSpace(address.House)
		address.Office = strings.TrimSpace(address.Office)
		address.PostalCode = strings.TrimSpace(address.PostalCode)
		if _, err := s.Insert(address); err != nil {
			return nil, err
		}
	}

	// Contact persons are a small ordered collection and are replaced atomically with
	// the profile. The response returns the new ids, so the frontend always stays in sync.
	if _, err := s.Where("project_id = ?", projectID).Delete(&ClientContactPerson{}); err != nil {
		return nil, err
	}
	for index, person := range contactPersons {
		if person == nil {
			continue
		}
		person.ID = 0
		person.ProjectID = projectID
		person.PositionIndex = index
		person.FullName = strings.TrimSpace(person.FullName)
		person.Position = strings.TrimSpace(person.Position)
		person.Department = strings.TrimSpace(person.Department)
		person.Phone = strings.TrimSpace(person.Phone)
		person.Email = strings.TrimSpace(person.Email)
		person.Telegram = strings.TrimSpace(person.Telegram)
		person.Viber = strings.TrimSpace(person.Viber)
		person.WhatsApp = strings.TrimSpace(person.WhatsApp)
		person.Birthday = strings.TrimSpace(person.Birthday)
		person.PreferredContactMethod = strings.TrimSpace(person.PreferredContactMethod)
		person.DecisionRole = strings.TrimSpace(person.DecisionRole)
		person.Notes = strings.TrimSpace(person.Notes)

		if person.FullName == "" {
			return nil, ErrInvalidData{Message: "contact person name is required"}
		}
		if _, ok := validClientDecisionRoles[person.DecisionRole]; !ok {
			return nil, ErrInvalidData{Message: "invalid contact person decision role"}
		}
		if person.Birthday != "" {
			if _, err := time.Parse("2006-01-02", person.Birthday); err != nil {
				return nil, ErrInvalidData{Message: "contact person birthday must use YYYY-MM-DD"}
			}
		}
		if _, err := s.Insert(person); err != nil {
			return nil, err
		}
	}

	// Project title remains the canonical name in menus. Keeping it synchronized
	// avoids the client card and the project navigation drifting apart.
	if project.Title != input.DisplayName {
		project.Title = input.DisplayName
		project.Updated = time.Now()
		if _, err := s.ID(project.ID).Cols("title", "updated").Update(project); err != nil {
			return nil, err
		}
		events.DispatchOnCommit(s, &ProjectUpdatedEvent{Project: project, Doer: doerFromAuth(s, a)})
	}

	doer := doerFromAuth(s, a)
	if has && oldStatus != input.Status {
		if err := CreateSystemClientActivity(s, projectID, ClientActivityStatusChanged, doer, "client", projectID, "", &ClientActivityMetadata{
			OldValue: oldStatus,
			NewValue: input.Status,
		}); err != nil {
			return nil, err
		}
	}

	if oldResponsibleID != input.ResponsibleUserID {
		oldName := ""
		if oldResponsibleID > 0 {
			users, err := user.GetUsersByIDs(s, []int64{oldResponsibleID})
			if err != nil {
				return nil, err
			}
			if oldUser, ok := users[oldResponsibleID]; ok {
				oldName = oldUser.GetName()
			}
		}
		newName := ""
		if responsible != nil {
			newName = responsible.GetName()
		}
		if err := CreateSystemClientActivity(s, projectID, ClientActivityResponsibleChanged, doer, "user", input.ResponsibleUserID, "", &ClientActivityMetadata{
			OldValue: oldName,
			NewValue: newName,
		}); err != nil {
			return nil, err
		}
	}

	if responsible != nil && oldResponsibleID != input.ResponsibleUserID {
		if err := notifications.Notify(responsible, &ClientResponsibleAssignedNotification{
			Doer:        doer,
			Project:     project,
			Responsible: responsible,
		}, s); err != nil {
			return nil, err
		}
	}

	return GetClientProfile(s, projectID)
}

// SetClientCommercialProposal replaces the project's single CRM proposal file.
// The caller must already have verified project write permission.
func SetClientCommercialProposal(s *xorm.Session, a web.Auth, projectID int64, reader io.ReadSeeker, filename string, size uint64) (*ClientProfile, error) {
	project, err := GetProjectSimpleByID(s, projectID)
	if err != nil {
		return nil, err
	}

	existing := &ClientProfile{ProjectID: projectID}
	has, err := s.ID(projectID).Get(existing)
	if err != nil {
		return nil, err
	}
	if !has {
		existing = defaultClientProfile(project)
		if _, err := s.Insert(existing); err != nil {
			return nil, err
		}
	}

	newFile, err := files.CreateWithSession(s, reader, filename, size, a)
	if err != nil {
		return nil, err
	}
	oldFileID := existing.CommercialProposalFileID
	oldFileName := ""
	if oldFileID > 0 {
		oldFileMeta := &files.File{ID: oldFileID}
		hasOldFile, err := s.ID(oldFileID).Get(oldFileMeta)
		if err != nil {
			_ = newFile.Delete(s)
			return nil, err
		}
		if hasOldFile {
			oldFileName = oldFileMeta.Name
		}
	}
	if _, err := s.ID(projectID).Cols("commercial_proposal_file_id").Update(&ClientProfile{CommercialProposalFileID: newFile.ID}); err != nil {
		_ = newFile.Delete(s)
		return nil, err
	}
	if oldFileID > 0 && oldFileID != newFile.ID {
		oldFile := &files.File{ID: oldFileID}
		if err := oldFile.Delete(s); err != nil && !files.IsErrFileDoesNotExist(err) {
			return nil, err
		}
	}
	eventType := ClientActivityProposalUploaded
	if oldFileID > 0 {
		eventType = ClientActivityProposalReplaced
	}
	if err := CreateSystemClientActivity(s, projectID, eventType, doerFromAuth(s, a), "commercial_proposal", newFile.ID, "", &ClientActivityMetadata{
		FileName: newFile.Name,
		OldValue: oldFileName,
		NewValue: newFile.Name,
	}); err != nil {
		return nil, err
	}
	return GetClientProfile(s, projectID)
}

// DeleteClientCommercialProposal removes the proposal reference and its stored file.
func DeleteClientCommercialProposal(s *xorm.Session, a web.Auth, projectID int64) error {
	profile := &ClientProfile{ProjectID: projectID}
	has, err := s.ID(projectID).Get(profile)
	if err != nil || !has {
		return err
	}
	fileID := profile.CommercialProposalFileID
	if fileID == 0 {
		return nil
	}
	f := &files.File{ID: fileID}
	fileName := ""
	if hasFile, err := s.ID(fileID).Get(f); err != nil {
		return err
	} else if hasFile {
		fileName = f.Name
	}
	if _, err := s.ID(projectID).Cols("commercial_proposal_file_id").Update(&ClientProfile{CommercialProposalFileID: 0}); err != nil {
		return err
	}
	if err := f.Delete(s); err != nil && !files.IsErrFileDoesNotExist(err) {
		return err
	}
	return CreateSystemClientActivity(s, projectID, ClientActivityProposalDeleted, doerFromAuth(s, a), "commercial_proposal", fileID, "", &ClientActivityMetadata{FileName: fileName})
}

func DeleteClientDataForProject(s *xorm.Session, projectID int64) error {
	profile := &ClientProfile{ProjectID: projectID}
	_, err := s.ID(projectID).Get(profile)
	if err != nil {
		return err
	}
	if _, err := s.Where("project_id = ?", projectID).Delete(&ClientContactPerson{}); err != nil {
		return err
	}
	if _, err := s.Where("project_id = ?", projectID).Delete(&ClientAddress{}); err != nil {
		return err
	}
	if _, err := s.Where("project_id = ?", projectID).Delete(&ClientCustomField{}); err != nil {
		return err
	}
	if _, err := s.Where("project_id = ?", projectID).Delete(&ClientActivityEvent{}); err != nil {
		return err
	}
	if profile.CommercialProposalFileID > 0 {
		f := &files.File{ID: profile.CommercialProposalFileID}
		if err := f.Delete(s); err != nil && !files.IsErrFileDoesNotExist(err) {
			return err
		}
	}
	_, err = s.ID(projectID).Delete(&ClientProfile{})
	return err
}
