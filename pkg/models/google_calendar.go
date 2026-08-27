package models

import (
	"crypto/subtle"
	"time"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

type GoogleCalendarConnection struct {
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id"`

	UserID int64 `xorm:"bigint not null unique index" json:"-"`

	GoogleEmail string `xorm:"varchar(320)" json:"google_email"`

	RefreshTokenEncrypted string `xorm:"text" json:"-"`

	VikunjaCalendarID string `xorm:"varchar(1024)" json:"vikunja_calendar_id"`

	SelectedCalendarIDs string `xorm:"text" json:"-"`

	OAuthStateHash      string    `xorm:"varchar(64)" json:"-"`
	OAuthStateExpiresAt time.Time `xorm:"index" json:"-"`

	ConnectedAt time.Time `json:"connected_at"`

	Created time.Time `xorm:"created not null" json:"created"`
	Updated time.Time `xorm:"updated not null" json:"updated"`
}

func (*GoogleCalendarConnection) TableName() string {
	return "google_calendar_connections"
}

func googleCalendarAuthUser(a web.Auth) (*user.User, bool) {
	u, ok := a.(*user.User)

	return u, ok && u != nil && u.ID > 0
}

func GetGoogleCalendarConnection(
	s *xorm.Session,
	a web.Auth,
) (*GoogleCalendarConnection, bool, error) {
	u, ok := googleCalendarAuthUser(a)

	if !ok {
		return nil, false, ErrGenericForbidden{}
	}

	connection := &GoogleCalendarConnection{}

	found, err := s.
		Where("user_id = ?", u.ID).
		Get(connection)

	if err != nil {
		return nil, false, err
	}

	return connection, found, nil
}

func SaveGoogleCalendarOAuthState(
	s *xorm.Session,
	a web.Auth,
	stateHash string,
	expiresAt time.Time,
) error {
	u, ok := googleCalendarAuthUser(a)

	if !ok {
		return ErrGenericForbidden{}
	}

	connection, found, err := GetGoogleCalendarConnection(s, a)
	if err != nil {
		return err
	}

	if !found {
		connection = &GoogleCalendarConnection{
			UserID:              u.ID,
			OAuthStateHash:      stateHash,
			OAuthStateExpiresAt: expiresAt,
		}

		_, err = s.Insert(connection)

		return err
	}

	connection.OAuthStateHash = stateHash
	connection.OAuthStateExpiresAt = expiresAt
	connection.Updated = time.Now()

	_, err = s.
		Where("user_id = ?", u.ID).
		Cols(
			"oauth_state_hash",
			"oauth_state_expires_at",
			"updated",
		).
		Update(connection)

	return err
}

func ValidateGoogleCalendarOAuthState(
	s *xorm.Session,
	a web.Auth,
	stateHash string,
	now time.Time,
) (bool, error) {
	connection, found, err := GetGoogleCalendarConnection(s, a)
	if err != nil || !found {
		return false, err
	}

	if connection.OAuthStateHash == "" {
		return false, nil
	}

	if connection.OAuthStateExpiresAt.IsZero() ||
		now.After(connection.OAuthStateExpiresAt) {
		return false, nil
	}

	valid := subtle.ConstantTimeCompare(
		[]byte(connection.OAuthStateHash),
		[]byte(stateHash),
	) == 1

	if !valid {
		return false, nil
	}

	connection.OAuthStateHash = ""
	connection.OAuthStateExpiresAt = time.Time{}
	connection.Updated = now

	_, err = s.
		Where("user_id = ?", connection.UserID).
		Cols(
			"oauth_state_hash",
			"oauth_state_expires_at",
			"updated",
		).
		Update(connection)

	return err == nil, err
}

func CompleteGoogleCalendarConnection(
	s *xorm.Session,
	a web.Auth,
	encryptedRefreshToken string,
) error {
	connection, found, err := GetGoogleCalendarConnection(s, a)
	if err != nil {
		return err
	}

	if !found {
		return ErrInvalidData{
			Message: "google calendar connection state was not found",
		}
	}

	now := time.Now()

	connection.RefreshTokenEncrypted = encryptedRefreshToken
	connection.ConnectedAt = now
	connection.Updated = now

	_, err = s.
		Where("user_id = ?", connection.UserID).
		Cols(
			"refresh_token_encrypted",
			"connected_at",
			"updated",
		).
		Update(connection)

	return err
}

func DeleteGoogleCalendarConnection(
	s *xorm.Session,
	a web.Auth,
) error {
	u, ok := googleCalendarAuthUser(a)

	if !ok {
		return ErrGenericForbidden{}
	}

	_, err := s.
		Where("user_id = ?", u.ID).
		Delete(&GoogleCalendarConnection{})

	return err
}
