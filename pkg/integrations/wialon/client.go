// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Package wialon contains the server-side Wialon Remote API adapter used by
// the fleet map. Keeping the adapter out of the HTTP route makes it reusable
// later by task/geofence listeners without exposing the Wialon token to the
// browser.
package wialon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultAPIURL        = "https://hst-api.wialon.com"
	unitListFlags        = 1 + 1024 + 2097152 // base + last message/location + connection status
	locationMessageFlag  = 1
	locationMessageMask  = 65281 // 0xFF01: data messages with location
	allMessagesLoadCount = uint64(0xffffffff)
)

// Config defines a Wialon Remote API connection.
type Config struct {
	APIURL         string
	Token          string
	Timeout        time.Duration
	TrackMaxPoints int
}

// Client is safe for concurrent use. Wialon's message loader is session-scoped,
// so track operations are serialized while unit-list reads can happen in parallel.
type Client struct {
	endpoint       string
	token          string
	http           *http.Client
	trackMaxPoints int

	sessionMu sync.Mutex
	sid       string
	trackMu   sync.Mutex
}

// Unit is the subset of a Wialon unit required by the fleet map and future task bindings.
type Unit struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Position   *Position `json:"position,omitempty"`
	Connected  bool      `json:"connected"`
	LastUpdate int64     `json:"last_update,omitempty"`
}

// Position is the last known GPS position of a unit.
type Position struct {
	Time       int64   `json:"time"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Altitude   float64 `json:"altitude"`
	Speed      int     `json:"speed"`
	Course     int     `json:"course"`
	Satellites int     `json:"satellites"`
}

// TrackPoint is one point of the unit's historical path.
type TrackPoint struct {
	Time      int64   `json:"time"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Speed     int     `json:"speed"`
	Course    int     `json:"course"`
}

// Track contains a route reconstructed from Wialon location messages.
type Track struct {
	UnitID   int64        `json:"unit_id"`
	From     int64        `json:"from"`
	To       int64        `json:"to"`
	Points   []TrackPoint `json:"points"`
	Original int          `json:"original_point_count"`
}

// APIError is returned when Wialon reports a Remote API error code.
type APIError struct {
	Code int
}

func (e *APIError) Error() string { return fmt.Sprintf("wialon api error %d", e.Code) }

// apiBool accepts both JSON booleans and Wialon installations that serialize
// boolean state fields as numeric 0/1 values.
type apiBool bool

func (b *apiBool) UnmarshalJSON(data []byte) error {
	switch strings.TrimSpace(string(data)) {
	case "true", "1":
		*b = true
		return nil
	case "false", "0", "null":
		*b = false
		return nil
	default:
		return fmt.Errorf("invalid Wialon boolean value %q", string(data))
	}
}

// NewClient creates a Wialon Remote API client.
func NewClient(cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.Token) == "" {
		return nil, errors.New("wialon token is not configured")
	}

	apiURL := strings.TrimSpace(cfg.APIURL)
	if apiURL == "" {
		apiURL = defaultAPIURL
	}
	endpoint, err := normalizeEndpoint(apiURL)
	if err != nil {
		return nil, err
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.TrackMaxPoints <= 0 {
		cfg.TrackMaxPoints = 5000
	}

	return &Client{
		endpoint:       endpoint,
		token:          cfg.Token,
		http:           &http.Client{Timeout: cfg.Timeout},
		trackMaxPoints: cfg.TrackMaxPoints,
	}, nil
}

func normalizeEndpoint(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid Wialon API URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", errors.New("Wialon API URL must use http or https")
	}
	if u.Host == "" {
		return "", errors.New("Wialon API URL must include a host")
	}
	if !strings.HasSuffix(u.Path, "/wialon/ajax.html") {
		u.Path = strings.TrimRight(u.Path, "/") + "/wialon/ajax.html"
	}
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}

// Ping verifies token authorization and keeps/reuses a Wialon session.
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.ensureSession(ctx)
	return err
}

// ListUnits returns all units visible to the configured Wialon token, including
// last known location and connection state.
func (c *Client) ListUnits(ctx context.Context) ([]Unit, error) {
	params := map[string]any{
		"spec": map[string]any{
			"itemsType":     "avl_unit",
			"propName":      "sys_name",
			"propValueMask": "*",
			"sortType":      "sys_name",
			"propType":      "property",
			"or_logic":      false,
		},
		"force": 1,
		"flags": unitListFlags,
		"from":  0,
		"to":    0,
	}

	var response struct {
		Items []struct {
			ID      int64   `json:"id"`
			Name    string  `json:"nm"`
			NetConn apiBool `json:"netconn"`
			Pos     *struct {
				Time       int64   `json:"t"`
				Latitude   float64 `json:"y"`
				Longitude  float64 `json:"x"`
				Altitude   float64 `json:"z"`
				Speed      int     `json:"s"`
				Course     int     `json:"c"`
				Satellites int     `json:"sc"`
			} `json:"pos"`
		} `json:"items"`
	}
	if err := c.callWithSession(ctx, "core/search_items", params, &response); err != nil {
		return nil, fmt.Errorf("core/search_items: %w", err)
	}

	units := make([]Unit, 0, len(response.Items))
	for _, item := range response.Items {
		u := Unit{ID: item.ID, Name: item.Name, Connected: bool(item.NetConn)}
		if item.Pos != nil {
			u.Position = &Position{
				Time: item.Pos.Time, Latitude: item.Pos.Latitude, Longitude: item.Pos.Longitude,
				Altitude: item.Pos.Altitude, Speed: item.Pos.Speed, Course: item.Pos.Course, Satellites: item.Pos.Satellites,
			}
			u.LastUpdate = item.Pos.Time
		}
		units = append(units, u)
	}
	return units, nil
}

// LoadTrack reconstructs a unit track from Wialon data messages with GPS data.
// The full interval is requested from Wialon, then evenly downsampled for the UI
// so a long route cannot create an unbounded browser payload.
func (c *Client) LoadTrack(ctx context.Context, unitID, from, to int64) (*Track, error) {
	if unitID <= 0 {
		return nil, errors.New("unit id must be positive")
	}
	if from <= 0 || to <= 0 || from >= to {
		return nil, errors.New("invalid track interval")
	}

	c.trackMu.Lock()
	defer c.trackMu.Unlock()

	params := map[string]any{
		"itemId":    unitID,
		"timeFrom":  from,
		"timeTo":    to,
		"flags":     locationMessageFlag,
		"flagsMask": locationMessageMask,
		"loadCount": allMessagesLoadCount,
	}
	var response struct {
		Count    int `json:"count"`
		Messages []struct {
			Time int64 `json:"t"`
			Pos  *struct {
				Latitude  float64 `json:"y"`
				Longitude float64 `json:"x"`
				Speed     int     `json:"s"`
				Course    int     `json:"c"`
			} `json:"pos"`
		} `json:"messages"`
	}
	if err := c.callWithSession(ctx, "messages/load_interval", params, &response); err != nil {
		return nil, fmt.Errorf("messages/load_interval: %w", err)
	}
	// Best effort. The next load replaces the loader anyway, but unloading keeps
	// the session clean and mirrors Wialon's recommended loader lifecycle.
	_ = c.callWithSession(ctx, "messages/unload", map[string]any{}, &struct{}{})

	points := make([]TrackPoint, 0, len(response.Messages))
	for _, msg := range response.Messages {
		if msg.Pos == nil {
			continue
		}
		points = append(points, TrackPoint{
			Time: msg.Time, Latitude: msg.Pos.Latitude, Longitude: msg.Pos.Longitude,
			Speed: msg.Pos.Speed, Course: msg.Pos.Course,
		})
	}
	original := len(points)
	points = downsample(points, c.trackMaxPoints)

	return &Track{UnitID: unitID, From: from, To: to, Points: points, Original: original}, nil
}

func downsample(points []TrackPoint, max int) []TrackPoint {
	if max <= 0 || len(points) <= max {
		return points
	}
	if max == 1 {
		return []TrackPoint{points[len(points)-1]}
	}
	out := make([]TrackPoint, 0, max)
	last := len(points) - 1
	for i := 0; i < max; i++ {
		idx := int(float64(i) * float64(last) / float64(max-1))
		out = append(out, points[idx])
	}
	return out
}

func (c *Client) ensureSession(ctx context.Context) (string, error) {
	c.sessionMu.Lock()
	defer c.sessionMu.Unlock()
	if c.sid != "" {
		return c.sid, nil
	}

	var login struct {
		EID string `json:"eid"`
	}
	if err := c.call(ctx, "token/login", map[string]any{"token": c.token, "fl": 1}, "", &login); err != nil {
		return "", fmt.Errorf("token/login: %w", err)
	}
	if login.EID == "" {
		return "", errors.New("wialon token login returned an empty session id")
	}
	c.sid = login.EID
	return c.sid, nil
}

func (c *Client) invalidateSession(sid string) {
	c.sessionMu.Lock()
	if c.sid == sid {
		c.sid = ""
	}
	c.sessionMu.Unlock()
}

func (c *Client) callWithSession(ctx context.Context, svc string, params any, out any) error {
	sid, err := c.ensureSession(ctx)
	if err != nil {
		return err
	}
	err = c.call(ctx, svc, params, sid, out)
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.Code == 1 {
		c.invalidateSession(sid)
		sid, loginErr := c.ensureSession(ctx)
		if loginErr != nil {
			return loginErr
		}
		return c.call(ctx, svc, params, sid, out)
	}
	return err
}

func (c *Client) call(ctx context.Context, svc string, params any, sid string, out any) error {
	payload, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("marshal Wialon params: %w", err)
	}
	// Wialon Remote API expects svc, params and sid in the URL query string
	// even though the HTTP method is POST. This is the format used by the
	// official Wialon documentation and Postman collection.
	requestURL, err := url.Parse(c.endpoint)
	if err != nil {
		return fmt.Errorf("parse Wialon endpoint: %w", err)
	}
	query := requestURL.Query()
	query.Set("svc", svc)
	query.Set("params", string(payload))
	if sid != "" {
		query.Set("sid", sid)
	}
	requestURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL.String(), http.NoBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("Wialon request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return fmt.Errorf("read Wialon response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Wialon HTTP status %s", resp.Status)
	}

	var errorEnvelope struct {
		Error json.RawMessage `json:"error"`
	}
	if json.Unmarshal(body, &errorEnvelope) == nil && len(errorEnvelope.Error) > 0 {
		var code int
		if err := json.Unmarshal(errorEnvelope.Error, &code); err == nil && code != 0 {
			return &APIError{Code: code}
		}
		var codeString string
		if err := json.Unmarshal(errorEnvelope.Error, &codeString); err == nil {
			if parsed, parseErr := strconv.Atoi(codeString); parseErr == nil && parsed != 0 {
				return &APIError{Code: parsed}
			}
		}
	}

	if out == nil || len(body) == 0 || string(body) == "null" || string(body) == "0" {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decode Wialon response for %s: %w", svc, err)
	}
	return nil
}
