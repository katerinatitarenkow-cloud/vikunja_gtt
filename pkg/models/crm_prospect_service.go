package models

import (
	"strings"
	"time"
	"unicode"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/builder"
	"xorm.io/xorm"
)

func prospectAuthUser(a web.Auth) (*user.User, bool) {
	u, ok := a.(*user.User)
	return u, ok && u != nil && u.ID > 0
}

func NormalizeProspectPhone(phone string) string {
	phone = strings.TrimSpace(phone)

	var digits strings.Builder
	for _, r := range phone {
		if unicode.IsDigit(r) {
			digits.WriteRune(r)
		}
	}

	normalized := digits.String()

	if len(normalized) == 10 && strings.HasPrefix(normalized, "0") {
		normalized = "38" + normalized
	}

	return normalized
}

func validateProspectCampaign(
	s *xorm.Session,
	campaignID int64,
) error {
	if campaignID <= 0 {
		return nil
	}

	has, err := s.ID(campaignID).Exist(&ProspectCampaign{})
	if err != nil {
		return err
	}
	if !has {
		return ErrInvalidData{
			Message: "prospect campaign does not exist",
		}
	}

	return nil
}

func validateProspectResponsible(
	s *xorm.Session,
	userID int64,
) error {
	if userID <= 0 {
		return nil
	}

	target, err := user.GetUserByID(s, userID)
	if err != nil {
		return err
	}
	if target.IsBot() {
		return ErrInvalidData{
			Message: "bot users cannot be prospect responsibles",
		}
	}

	return nil
}

func prepareProspect(
	s *xorm.Session,
	prospect *Prospect,
) error {
	prospect.CompanyName = strings.TrimSpace(prospect.CompanyName)
	prospect.ContactName = strings.TrimSpace(prospect.ContactName)
	prospect.Phone = strings.TrimSpace(prospect.Phone)
	prospect.PhoneSecondary = strings.TrimSpace(prospect.PhoneSecondary)
	prospect.Email = strings.TrimSpace(prospect.Email)
	prospect.Website = strings.TrimSpace(prospect.Website)
	prospect.Region = strings.TrimSpace(prospect.Region)
	prospect.City = strings.TrimSpace(prospect.City)
	prospect.Notes = strings.TrimSpace(prospect.Notes)
	prospect.Status = strings.TrimSpace(prospect.Status)

	if prospect.Status == "" {
		prospect.Status = ProspectStatusNew
	}

	if err := validateProspectStatus(prospect.Status); err != nil {
		return err
	}

	if prospect.CompanyName == "" &&
		prospect.ContactName == "" &&
		prospect.Phone == "" &&
		prospect.Email == "" {
		return ErrInvalidData{
			Message: "prospect requires at least a company, contact, phone or email",
		}
	}

	prospect.PhoneNormalized = NormalizeProspectPhone(prospect.Phone)

	if err := validateProspectCampaign(s, prospect.CampaignID); err != nil {
		return err
	}

	if err := validateProspectResponsible(s, prospect.ResponsibleUserID); err != nil {
		return err
	}

	return nil
}

func hydrateProspectResponsibles(
	s *xorm.Session,
	prospects []*Prospect,
) error {
	ids := make([]int64, 0, len(prospects))
	seen := map[int64]struct{}{}

	for _, prospect := range prospects {
		if prospect.ResponsibleUserID <= 0 {
			continue
		}
		if _, exists := seen[prospect.ResponsibleUserID]; exists {
			continue
		}

		seen[prospect.ResponsibleUserID] = struct{}{}
		ids = append(ids, prospect.ResponsibleUserID)
	}

	if len(ids) == 0 {
		return nil
	}

	users, err := user.GetUsersByIDs(s, ids)
	if err != nil {
		return err
	}

	for _, prospect := range prospects {
		prospect.Responsible = users[prospect.ResponsibleUserID]
	}

	return nil
}

func CreateProspectCampaign(
	s *xorm.Session,
	a web.Auth,
	input *ProspectCampaign,
) (*ProspectCampaign, error) {
	current, ok := prospectAuthUser(a)
	if !ok {
		return nil, ErrGenericForbidden{}
	}

	campaign := &ProspectCampaign{
		Name:            strings.TrimSpace(input.Name),
		Description:     strings.TrimSpace(input.Description),
		CreatedByUserID: current.ID,
		Archived:        false,
	}

	if campaign.Name == "" {
		return nil, ErrInvalidData{
			Message: "prospect campaign name is required",
		}
	}

	if _, err := s.Insert(campaign); err != nil {
		return nil, err
	}

	return campaign, nil
}

func ListProspectCampaigns(
	s *xorm.Session,
	a web.Auth,
	includeArchived bool,
) ([]*ProspectCampaign, error) {
	if _, ok := prospectAuthUser(a); !ok {
		return nil, ErrGenericForbidden{}
	}

	query := s.Table((&ProspectCampaign{}).TableName())

	if !includeArchived {
		query = query.Where("archived = ?", false)
	}

	campaigns := []*ProspectCampaign{}

	if err := query.
		OrderBy("created desc, id desc").
		Find(&campaigns); err != nil {
		return nil, err
	}

	return campaigns, nil
}

func CreateProspect(
	s *xorm.Session,
	a web.Auth,
	input *Prospect,
) (*Prospect, error) {
	if _, ok := prospectAuthUser(a); !ok {
		return nil, ErrGenericForbidden{}
	}

	prospect := &Prospect{
		CampaignID:        input.CampaignID,
		CompanyName:       input.CompanyName,
		ContactName:       input.ContactName,
		Phone:             input.Phone,
		PhoneSecondary:    input.PhoneSecondary,
		Email:             input.Email,
		Website:           input.Website,
		Region:            input.Region,
		City:              input.City,
		Notes:             input.Notes,
		Status:            input.Status,
		ResponsibleUserID: input.ResponsibleUserID,
	}

	if err := prepareProspect(s, prospect); err != nil {
		return nil, err
	}

	if _, err := s.Insert(prospect); err != nil {
		return nil, err
	}

	if err := hydrateProspectResponsibles(
		s,
		[]*Prospect{prospect},
	); err != nil {
		return nil, err
	}

	return prospect, nil
}

func prospectListSession(
	s *xorm.Session,
	search string,
	status string,
	campaignID int64,
	responsibleUserID int64,
) (*xorm.Session, error) {
	query := s.Table((&Prospect{}).TableName())

	status = strings.TrimSpace(status)
	if status != "" {
		if err := validateProspectStatus(status); err != nil {
			return nil, err
		}
		query = query.Where("status = ?", status)
	}

	if campaignID > 0 {
		query = query.Where("campaign_id = ?", campaignID)
	}

	if responsibleUserID > 0 {
		query = query.Where(
			"responsible_user_id = ?",
			responsibleUserID,
		)
	}

	search = strings.TrimSpace(search)
	if search != "" {
		pattern := "%" + search + "%"
		normalizedPhone := NormalizeProspectPhone(search)

		conditions := []builder.Cond{
			builder.Like{"company_name", pattern},
			builder.Like{"contact_name", pattern},
			builder.Like{"phone", pattern},
			builder.Like{"email", pattern},
			builder.Like{"region", pattern},
			builder.Like{"city", pattern},
		}

		if normalizedPhone != "" {
			conditions = append(
				conditions,
				builder.Like{
					"phone_normalized",
					"%" + normalizedPhone + "%",
				},
			)
		}

		query = query.And(builder.Or(conditions...))
	}

	return query, nil
}

func ListProspects(
	s *xorm.Session,
	a web.Auth,
	search string,
	status string,
	campaignID int64,
	responsibleUserID int64,
	page int,
	perPage int,
) ([]*Prospect, int64, error) {
	if _, ok := prospectAuthUser(a); !ok {
		return nil, 0, ErrGenericForbidden{}
	}

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 200 {
		perPage = 50
	}

	countQuery, err := prospectListSession(
		s,
		search,
		status,
		campaignID,
		responsibleUserID,
	)
	if err != nil {
		return nil, 0, err
	}

	total, err := countQuery.Count(&Prospect{})
	if err != nil {
		return nil, 0, err
	}

	listQuery, err := prospectListSession(
		s,
		search,
		status,
		campaignID,
		responsibleUserID,
	)
	if err != nil {
		return nil, 0, err
	}

	prospects := []*Prospect{}

	if err := listQuery.
		OrderBy(`
CASE WHEN next_contact_at IS NULL THEN 1 ELSE 0 END ASC,
next_contact_at ASC,
created DESC,
id DESC
`).
		Limit(perPage, (page-1)*perPage).
		Find(&prospects); err != nil {
		return nil, 0, err
	}

	if err := hydrateProspectResponsibles(
		s,
		prospects,
	); err != nil {
		return nil, 0, err
	}

	return prospects, total, nil
}

func GetProspect(
	s *xorm.Session,
	a web.Auth,
	prospectID int64,
) (*Prospect, bool, error) {
	if _, ok := prospectAuthUser(a); !ok {
		return nil, false, ErrGenericForbidden{}
	}

	prospect := &Prospect{ID: prospectID}

	has, err := s.ID(prospectID).Get(prospect)
	if err != nil {
		return nil, false, err
	}
	if !has {
		return nil, false, nil
	}

	if err := hydrateProspectResponsibles(
		s,
		[]*Prospect{prospect},
	); err != nil {
		return nil, false, err
	}

	return prospect, true, nil
}

func UpdateProspect(
	s *xorm.Session,
	a web.Auth,
	prospectID int64,
	input *Prospect,
) (*Prospect, bool, error) {
	if _, ok := prospectAuthUser(a); !ok {
		return nil, false, ErrGenericForbidden{}
	}

	stored := &Prospect{ID: prospectID}

	has, err := s.ID(prospectID).Get(stored)
	if err != nil {
		return nil, false, err
	}
	if !has {
		return nil, false, nil
	}

	stored.CampaignID = input.CampaignID
	stored.CompanyName = input.CompanyName
	stored.ContactName = input.ContactName
	stored.Phone = input.Phone
	stored.PhoneSecondary = input.PhoneSecondary
	stored.Email = input.Email
	stored.Website = input.Website
	stored.Region = input.Region
	stored.City = input.City
	stored.Notes = input.Notes
	stored.Status = input.Status
	stored.ResponsibleUserID = input.ResponsibleUserID

	if err := prepareProspect(s, stored); err != nil {
		return nil, true, err
	}

	stored.Updated = time.Now()

	_, err = s.ID(prospectID).
		Cols(
			"campaign_id",
			"company_name",
			"contact_name",
			"phone",
			"phone_normalized",
			"phone_secondary",
			"email",
			"website",
			"region",
			"city",
			"notes",
			"status",
			"responsible_user_id",
			"updated",
		).
		Update(stored)
	if err != nil {
		return nil, true, err
	}

	if err := hydrateProspectResponsibles(
		s,
		[]*Prospect{stored},
	); err != nil {
		return nil, true, err
	}

	return stored, true, nil
}

func statusForProspectCallOutcome(
	outcome string,
) string {
	switch outcome {
	case ProspectCallOutcomeNoAnswer,
		ProspectCallOutcomeBusy:
		return ProspectStatusNoAnswer

	case ProspectCallOutcomeConversation:
		return ProspectStatusInProgress

	case ProspectCallOutcomeCallback:
		return ProspectStatusCallback

	case ProspectCallOutcomeInterested:
		return ProspectStatusInterested

	case ProspectCallOutcomeNotInterested:
		return ProspectStatusNotInterested

	case ProspectCallOutcomeWrongNumber:
		return ProspectStatusInvalid
	}

	return ProspectStatusInProgress
}

func CreateProspectCall(
	s *xorm.Session,
	a web.Auth,
	prospectID int64,
	input *ProspectCall,
) (*ProspectCall, bool, error) {
	current, ok := prospectAuthUser(a)
	if !ok {
		return nil, false, ErrGenericForbidden{}
	}

	prospect := &Prospect{ID: prospectID}

	has, err := s.ID(prospectID).Get(prospect)
	if err != nil {
		return nil, false, err
	}
	if !has {
		return nil, false, nil
	}

	call := &ProspectCall{
		ProspectID:      prospectID,
		UserID:          current.ID,
		OccurredAt:      input.OccurredAt,
		Outcome:         strings.TrimSpace(input.Outcome),
		DurationMinutes: input.DurationMinutes,
		Note:            strings.TrimSpace(input.Note),
		NextContactAt:   input.NextContactAt,
	}

	if call.OccurredAt.IsZero() {
		call.OccurredAt = time.Now()
	}

	if err := validateProspectCallOutcome(
		call.Outcome,
	); err != nil {
		return nil, true, err
	}

	if call.DurationMinutes < 0 {
		return nil, true, ErrInvalidData{
			Message: "call duration cannot be negative",
		}
	}

	if call.Outcome == ProspectCallOutcomeCallback &&
		call.NextContactAt == nil {
		return nil, true, ErrInvalidData{
			Message: "callback outcome requires next contact time",
		}
	}

	if _, err := s.Insert(call); err != nil {
		return nil, true, err
	}

	now := time.Now()

	prospect.Status = statusForProspectCallOutcome(
		call.Outcome,
	)
	prospect.LastContactAt = &call.OccurredAt
	prospect.NextContactAt = call.NextContactAt
	prospect.Updated = now

	_, err = s.ID(prospectID).
		Cols(
			"status",
			"last_contact_at",
			"next_contact_at",
			"updated",
		).
		Update(prospect)
	if err != nil {
		return nil, true, err
	}

	return call, true, nil
}

func ListProspectCalls(
	s *xorm.Session,
	a web.Auth,
	prospectID int64,
) ([]*ProspectCall, bool, error) {
	if _, ok := prospectAuthUser(a); !ok {
		return nil, false, ErrGenericForbidden{}
	}

	has, err := s.ID(prospectID).Exist(&Prospect{})
	if err != nil {
		return nil, false, err
	}
	if !has {
		return nil, false, nil
	}

	calls := []*ProspectCall{}

	if err := s.
		Where("prospect_id = ?", prospectID).
		OrderBy("occurred_at desc, id desc").
		Find(&calls); err != nil {
		return nil, true, err
	}

	return calls, true, nil
}
