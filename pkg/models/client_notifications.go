// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package models

import (
	"strconv"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/i18n"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"
)

// ClientResponsibleAssignedNotification is sent when a user becomes the
// responsible employee for a client/project.
type ClientResponsibleAssignedNotification struct {
	Doer        *user.User `json:"doer"`
	Project     *Project   `json:"project"`
	Responsible *user.User `json:"responsible"`
}

func init() {
	notifications.Register(func() notifications.PersistedNotification { return &ClientResponsibleAssignedNotification{} })
}

func (n *ClientResponsibleAssignedNotification) ToTitle(lang string) string {
	return i18n.T(lang, "notifications.client.responsible.subject", n.Project.Title)
}

func (n *ClientResponsibleAssignedNotification) ToMail(lang string) *notifications.Mail {
	return notifications.NewMail().
		Greeting(i18n.T(lang, "notifications.greeting", n.Responsible.GetName())).
		Line(i18n.T(lang, "notifications.client.responsible.message", n.Doer.GetName(), n.Project.Title)).
		Action(i18n.T(lang, "notifications.client.responsible.open"), config.ServicePublicURL.GetString()+"projects/"+strconv.FormatInt(n.Project.ID, 10)+"/client").
		IncludeLinkToSettings(lang)
}

func (n *ClientResponsibleAssignedNotification) ToDB() interface{} { return n }
func (n *ClientResponsibleAssignedNotification) Name() string      { return "client.responsible.assigned" }
func (n *ClientResponsibleAssignedNotification) ProjectID() int64 {
	if n.Project == nil || n.Project.ID <= 0 {
		return notifications.ProjectIDUnresolved
	}
	return n.Project.ID
}
