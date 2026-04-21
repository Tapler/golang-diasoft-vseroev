package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite
	apiURL string
	db     *sql.DB
}

type CreateEventRequest struct {
	Title        string `json:"title"`
	StartTime    string `json:"start_time"`
	Duration     int64  `json:"duration"`
	Description  string `json:"description,omitempty"`
	UserID       int64  `json:"user_id"`
	NotifyBefore int64  `json:"notify_before,omitempty"`
}

type EventResponse struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	StartTime    string `json:"start_time"`
	Duration     int64  `json:"duration"`
	Description  string `json:"description,omitempty"`
	UserID       int64  `json:"user_id"`
	NotifyBefore int64  `json:"notify_before,omitempty"`
}

type EventListResponse struct {
	Events []EventResponse `json:"events"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (s *IntegrationTestSuite) SetupSuite() {
	apiURL := os.Getenv("CALENDAR_API_URL")
	if apiURL == "" {
		apiURL = "http://calendar:8080"
	}
	s.apiURL = apiURL

	dbDSN := os.Getenv("DATABASE_DSN")
	if dbDSN == "" {
		dbDSN = "postgres://calendar:calendar@postgres:5432/calendar?sslmode=disable"
	}

	var err error
	s.db, err = sql.Open("postgres", dbDSN)
	require.NoError(s.T(), err)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for {
		err = s.db.PingContext(ctx)
		if err == nil {
			break
		}
		select {
		case <-ctx.Done():
			require.NoError(s.T(), fmt.Errorf("failed to connect to database: %w", err))
		case <-time.After(1 * time.Second):
		}
	}

	s.waitForAPI(ctx)
}

func (s *IntegrationTestSuite) waitForAPI(ctx context.Context) {
	client := &http.Client{Timeout: 5 * time.Second}
	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.apiURL+"/api/v1/events/day/2025-01-01", nil)
		if err == nil {
			resp, err := client.Do(req)
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode < 500 {
					return
				}
			}
		}

		select {
		case <-ctx.Done():
			require.NoError(s.T(), fmt.Errorf("API not ready"))
		case <-time.After(1 * time.Second):
		}
	}
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *IntegrationTestSuite) SetupTest() {
	_, err := s.db.Exec("TRUNCATE events, notifications CASCADE")
	require.NoError(s.T(), err)
}

func (s *IntegrationTestSuite) TestCreateEvent_Success() {
	startTime := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	reqBody := CreateEventRequest{
		Title:       "Test Event",
		StartTime:   startTime,
		Duration:    3600,
		Description: "Test Description",
		UserID:      1,
	}

	body, err := json.Marshal(reqBody)
	require.NoError(s.T(), err)

	resp, err := http.Post(
		s.apiURL+"/api/v1/events",
		"application/json",
		bytes.NewBuffer(body),
	)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var eventResp EventResponse
	err = json.NewDecoder(resp.Body).Decode(&eventResp)
	require.NoError(s.T(), err)

	require.NotEmpty(s.T(), eventResp.ID)
	require.Equal(s.T(), reqBody.Title, eventResp.Title)
	require.Equal(s.T(), reqBody.UserID, eventResp.UserID)

	var count int
	err = s.db.QueryRow("SELECT COUNT(*) FROM events WHERE id = $1", eventResp.ID).Scan(&count)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, count)
}

func (s *IntegrationTestSuite) TestCreateEvent_ValidationError() {
	tests := []struct {
		name     string
		request  CreateEventRequest
		wantCode int
	}{
		{
			name: "empty title",
			request: CreateEventRequest{
				Title:     "",
				StartTime: time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
				Duration:  3600,
				UserID:    1,
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "invalid duration",
			request: CreateEventRequest{
				Title:     "Test",
				StartTime: time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
				Duration:  30,
				UserID:    1,
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "invalid user_id",
			request: CreateEventRequest{
				Title:     "Test",
				StartTime: time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
				Duration:  3600,
				UserID:    0,
			},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			body, err := json.Marshal(tt.request)
			require.NoError(s.T(), err)

			resp, err := http.Post(
				s.apiURL+"/api/v1/events",
				"application/json",
				bytes.NewBuffer(body),
			)
			require.NoError(s.T(), err)
			defer resp.Body.Close()

			require.Equal(s.T(), tt.wantCode, resp.StatusCode)

			var errResp ErrorResponse
			err = json.NewDecoder(resp.Body).Decode(&errResp)
			require.NoError(s.T(), err)
			require.NotEmpty(s.T(), errResp.Error)
		})
	}
}

func (s *IntegrationTestSuite) TestCreateEvent_DateBusy() {
	startTime := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	reqBody := CreateEventRequest{
		Title:     "First Event",
		StartTime: startTime,
		Duration:  3600,
		UserID:    1,
	}

	body, err := json.Marshal(reqBody)
	require.NoError(s.T(), err)

	resp, err := http.Post(
		s.apiURL+"/api/v1/events",
		"application/json",
		bytes.NewBuffer(body),
	)
	require.NoError(s.T(), err)
	resp.Body.Close()
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	resp2, err := http.Post(
		s.apiURL+"/api/v1/events",
		"application/json",
		bytes.NewBuffer(body),
	)
	require.NoError(s.T(), err)
	defer resp2.Body.Close()

	require.Equal(s.T(), http.StatusConflict, resp2.StatusCode)

	var errResp ErrorResponse
	err = json.NewDecoder(resp2.Body).Decode(&errResp)
	require.NoError(s.T(), err)
	require.Contains(s.T(), errResp.Error, "busy")
}

func (s *IntegrationTestSuite) TestListEventsForDay() {
	now := time.Now().UTC()
	today := now.Format("2006-01-02")

	event1 := CreateEventRequest{
		Title:     "Event 1",
		StartTime: now.Add(1 * time.Hour).Format(time.RFC3339),
		Duration:  3600,
		UserID:    1,
	}
	event2 := CreateEventRequest{
		Title:     "Event 2",
		StartTime: now.Add(3 * time.Hour).Format(time.RFC3339),
		Duration:  3600,
		UserID:    1,
	}
	tomorrow := CreateEventRequest{
		Title:     "Tomorrow Event",
		StartTime: now.Add(25 * time.Hour).Format(time.RFC3339),
		Duration:  3600,
		UserID:    1,
	}

	for _, req := range []CreateEventRequest{event1, event2, tomorrow} {
		body, err := json.Marshal(req)
		require.NoError(s.T(), err)
		resp, err := http.Post(s.apiURL+"/api/v1/events", "application/json", bytes.NewBuffer(body))
		require.NoError(s.T(), err)
		resp.Body.Close()
		require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	}

	resp, err := http.Get(s.apiURL + "/api/v1/events/day/" + today)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var listResp EventListResponse
	err = json.NewDecoder(resp.Body).Decode(&listResp)
	require.NoError(s.T(), err)

	require.Len(s.T(), listResp.Events, 2)
}

func (s *IntegrationTestSuite) TestListEventsForWeek() {
	now := time.Now().UTC()
	today := now.Format("2006-01-02")

	events := []CreateEventRequest{
		{
			Title:     "Day 1",
			StartTime: now.Add(1 * time.Hour).Format(time.RFC3339),
			Duration:  3600,
			UserID:    1,
		},
		{
			Title:     "Day 3",
			StartTime: now.Add(48 * time.Hour).Format(time.RFC3339),
			Duration:  3600,
			UserID:    1,
		},
		{
			Title:     "Day 5",
			StartTime: now.Add(96 * time.Hour).Format(time.RFC3339),
			Duration:  3600,
			UserID:    1,
		},
		{
			Title:     "Next Week",
			StartTime: now.Add(8 * 24 * time.Hour).Format(time.RFC3339),
			Duration:  3600,
			UserID:    1,
		},
	}

	for _, req := range events {
		body, err := json.Marshal(req)
		require.NoError(s.T(), err)
		resp, err := http.Post(s.apiURL+"/api/v1/events", "application/json", bytes.NewBuffer(body))
		require.NoError(s.T(), err)
		resp.Body.Close()
		require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	}

	resp, err := http.Get(s.apiURL + "/api/v1/events/week/" + today)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var listResp EventListResponse
	err = json.NewDecoder(resp.Body).Decode(&listResp)
	require.NoError(s.T(), err)

	require.Len(s.T(), listResp.Events, 3)
}

func (s *IntegrationTestSuite) TestListEventsForMonth() {
	now := time.Now().UTC()
	today := now.Format("2006-01-02")

	// Создаем события с фиксированными днями месяца, чтобы гарантировать попадание в текущий месяц
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 12, 0, 0, 0, time.UTC)

	events := []CreateEventRequest{
		{
			Title:     "Week 1",
			StartTime: startOfMonth.AddDate(0, 0, 5).Format(time.RFC3339), // 6-е число месяца
			Duration:  3600,
			UserID:    1,
		},
		{
			Title:     "Week 2",
			StartTime: startOfMonth.AddDate(0, 0, 14).Format(time.RFC3339), // 15-е число месяца
			Duration:  3600,
			UserID:    1,
		},
		{
			Title:     "Week 3",
			StartTime: startOfMonth.AddDate(0, 0, 24).Format(time.RFC3339), // 25-е число месяца
			Duration:  3600,
			UserID:    1,
		},
		{
			Title:     "Next Month",
			StartTime: startOfMonth.AddDate(0, 1, 5).Format(time.RFC3339), // 6-е число следующего месяца
			Duration:  3600,
			UserID:    1,
		},
	}

	for _, req := range events {
		body, err := json.Marshal(req)
		require.NoError(s.T(), err)
		resp, err := http.Post(s.apiURL+"/api/v1/events", "application/json", bytes.NewBuffer(body))
		require.NoError(s.T(), err)
		resp.Body.Close()
		require.Equal(s.T(), http.StatusCreated, resp.StatusCode)
	}

	resp, err := http.Get(s.apiURL + "/api/v1/events/month/" + today)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var listResp EventListResponse
	err = json.NewDecoder(resp.Body).Decode(&listResp)
	require.NoError(s.T(), err)

	require.Len(s.T(), listResp.Events, 3)
}

func TestIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests")
	}

	suite.Run(t, new(IntegrationTestSuite))
}
