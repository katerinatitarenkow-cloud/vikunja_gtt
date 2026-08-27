package apiv2

import (
	"context"
	"net/http"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/danielgtaylor/huma/v2"
)

type prospectBody struct {
	Body models.Prospect
}

type prospectsBody struct {
	Body Paginated[*models.Prospect]
}

type prospectCampaignBody struct {
	Body models.ProspectCampaign
}

type prospectCampaignsBody struct {
	Body []*models.ProspectCampaign
}

type prospectCallBody struct {
	Body models.ProspectCall
}

type prospectCallsBody struct {
	Body []*models.ProspectCall
}

func RegisterCRMProspectRoutes(api huma.API) {
	tags := []string{"crm-call-center"}

	Register(api, huma.Operation{
		OperationID:   "crm-prospect-campaign-create",
		Summary:       "Create a prospect campaign",
		Method:        http.MethodPost,
		Path:          "/crm/prospect-campaigns",
		Tags:          tags,
		DefaultStatus: http.StatusCreated,
	}, crmProspectCampaignCreate)

	Register(api, huma.Operation{
		OperationID: "crm-prospect-campaign-list",
		Summary:     "List prospect campaigns",
		Method:      http.MethodGet,
		Path:        "/crm/prospect-campaigns",
		Tags:        tags,
	}, crmProspectCampaignList)

	Register(api, huma.Operation{
		OperationID:   "crm-prospect-create",
		Summary:       "Create a potential client",
		Method:        http.MethodPost,
		Path:          "/crm/prospects",
		Tags:          tags,
		DefaultStatus: http.StatusCreated,
	}, crmProspectCreate)

	Register(api, huma.Operation{
		OperationID: "crm-prospect-list",
		Summary:     "List potential clients",
		Description: "Server-paginated call-center list with search and filters.",
		Method:      http.MethodGet,
		Path:        "/crm/prospects",
		Tags:        tags,
	}, crmProspectList)

	Register(api, huma.Operation{
		OperationID: "crm-prospect-read",
		Summary:     "Read a potential client",
		Method:      http.MethodGet,
		Path:        "/crm/prospects/{prospect}",
		Tags:        tags,
	}, crmProspectRead)

	Register(api, huma.Operation{
		OperationID: "crm-prospect-update",
		Summary:     "Update a potential client",
		Method:      http.MethodPut,
		Path:        "/crm/prospects/{prospect}",
		Tags:        tags,
	}, crmProspectUpdate)

	Register(api, huma.Operation{
		OperationID:   "crm-prospect-call-create",
		Summary:       "Record a call with a potential client",
		Method:        http.MethodPost,
		Path:          "/crm/prospects/{prospect}/calls",
		Tags:          tags,
		DefaultStatus: http.StatusCreated,
	}, crmProspectCallCreate)

	Register(api, huma.Operation{
		OperationID: "crm-prospect-call-list",
		Summary:     "List calls with a potential client",
		Method:      http.MethodGet,
		Path:        "/crm/prospects/{prospect}/calls",
		Tags:        tags,
	}, crmProspectCallList)
}

func init() {
	AddRouteRegistrar(RegisterCRMProspectRoutes)
}

func crmProspectCampaignCreate(
	ctx context.Context,
	in *struct {
		Body models.ProspectCampaign
	},
) (*prospectCampaignBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	campaign, err := models.CreateProspectCampaign(
		s,
		a,
		&in.Body,
	)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	return &prospectCampaignBody{
		Body: *campaign,
	}, nil
}

func crmProspectCampaignList(
	ctx context.Context,
	in *struct {
		IncludeArchived bool `query:"include_archived" default:"false"`
	},
) (*prospectCampaignsBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	campaigns, err := models.ListProspectCampaigns(
		s,
		a,
		in.IncludeArchived,
	)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	return &prospectCampaignsBody{
		Body: campaigns,
	}, nil
}

func crmProspectCreate(
	ctx context.Context,
	in *struct {
		Body models.Prospect
	},
) (*prospectBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	prospect, err := models.CreateProspect(
		s,
		a,
		&in.Body,
	)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	return &prospectBody{
		Body: *prospect,
	}, nil
}

func crmProspectList(
	ctx context.Context,
	in *struct {
		Q                 string `query:"q"`
		Status            string `query:"status"`
		CampaignID        int64  `query:"campaign_id" minimum:"0"`
		ResponsibleUserID int64  `query:"responsible_user_id" minimum:"0"`
		Page              int    `query:"page" default:"1" minimum:"1"`
		PerPage           int    `query:"per_page" default:"50" minimum:"1" maximum:"200"`
	},
) (*prospectsBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	prospects, total, err := models.ListProspects(
		s,
		a,
		in.Q,
		in.Status,
		in.CampaignID,
		in.ResponsibleUserID,
		in.Page,
		in.PerPage,
	)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	return &prospectsBody{
		Body: NewPaginated(
			prospects,
			total,
			in.Page,
			in.PerPage,
		),
	}, nil
}

func crmProspectRead(
	ctx context.Context,
	in *struct {
		ProspectID int64 `path:"prospect" minimum:"1"`
	},
) (*prospectBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	prospect, found, err := models.GetProspect(
		s,
		a,
		in.ProspectID,
	)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if !found {
		_ = s.Rollback()
		return nil, huma.Error404NotFound(
			"prospect not found",
		)
	}

	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	return &prospectBody{
		Body: *prospect,
	}, nil
}

func crmProspectUpdate(
	ctx context.Context,
	in *struct {
		ProspectID int64 `path:"prospect" minimum:"1"`
		Body       models.Prospect
	},
) (*prospectBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	prospect, found, err := models.UpdateProspect(
		s,
		a,
		in.ProspectID,
		&in.Body,
	)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if !found {
		_ = s.Rollback()
		return nil, huma.Error404NotFound(
			"prospect not found",
		)
	}

	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	return &prospectBody{
		Body: *prospect,
	}, nil
}

func crmProspectCallCreate(
	ctx context.Context,
	in *struct {
		ProspectID int64 `path:"prospect" minimum:"1"`
		Body       models.ProspectCall
	},
) (*prospectCallBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	call, found, err := models.CreateProspectCall(
		s,
		a,
		in.ProspectID,
		&in.Body,
	)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if !found {
		_ = s.Rollback()
		return nil, huma.Error404NotFound(
			"prospect not found",
		)
	}

	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	return &prospectCallBody{
		Body: *call,
	}, nil
}

func crmProspectCallList(
	ctx context.Context,
	in *struct {
		ProspectID int64 `path:"prospect" minimum:"1"`
	},
) (*prospectCallsBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	calls, found, err := models.ListProspectCalls(
		s,
		a,
		in.ProspectID,
	)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if !found {
		_ = s.Rollback()
		return nil, huma.Error404NotFound(
			"prospect not found",
		)
	}

	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	return &prospectCallsBody{
		Body: calls,
	}, nil
}
