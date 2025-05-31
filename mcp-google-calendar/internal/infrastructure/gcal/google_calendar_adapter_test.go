package gcal

import (
	"context"
	"errors"
	"testing"

	"github.com/dch/mcp-google-calendar/internal/domain"
	"google.golang.org/api/calendar/v3"
)

type mockCalendarService struct {
	listCalendarsFunc func() (*calendar.CalendarList, error)
	createEventFunc   func(string, *calendar.Event) (*calendar.Event, error)
}

func (m *mockCalendarService) CalendarList() *calendar.CalendarListService {
	return &calendar.CalendarListService{}
}

func (m *mockCalendarService) Events() *calendar.EventsService {
	return &calendar.EventsService{}
}

func TestGoogleCalendarAdapter_ListCalendars(t *testing.T) {
	tests := []struct {
		name            string
		mockResponse    *calendar.CalendarList
		mockError       error
		expectError     bool
		expectCalendars []domain.Calendar
	}{
		{
			name: "successful response",
			mockResponse: &calendar.CalendarList{
				Items: []*calendar.CalendarListEntry{
					{
						Id:          "calendar1",
						Summary:     "Test Calendar 1",
						Description: "Test Description 1",
						TimeZone:    "Asia/Tokyo",
					},
				},
			},
			expectCalendars: []domain.Calendar{
				{
					ID:          "calendar1",
					Title:       "Test Calendar 1",
					Description: "Test Description 1",
					TimeZone:    "Asia/Tokyo",
				},
			},
		},
		{
			name:        "error response",
			mockError:   errors.New("API error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockCalendarService{
				listCalendarsFunc: func() (*calendar.CalendarList, error) {
					return tt.mockResponse, tt.mockError
				},
			}

			adapter := &GoogleCalendarAdapter{
				service: mock,
			}

			calendars, err := adapter.ListCalendars(context.Background())

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(calendars) != len(tt.expectCalendars) {
				t.Errorf("got %d calendars, want %d", len(calendars), len(tt.expectCalendars))
				return
			}

			for i, got := range calendars {
				want := tt.expectCalendars[i]
				if got.ID != want.ID || got.Title != want.Title {
					t.Errorf("calendar %d: got %+v, want %+v", i, got, want)
				}
			}
		})
	}
}

func TestGoogleCalendarAdapter_CreateEvent(t *testing.T) {
	tests := []struct {
		name         string
		calendarID   string
		event        *domain.Event
		mockResponse *calendar.Event
		mockError    error
		expectError  bool
	}{
		{
			name:       "successful creation",
			calendarID: "primary",
			event: &domain.Event{
				Title:       "Test Event",
				Description: "Test Description",
				Start: domain.DateTime{
					DateTime: "2025-06-01T10:00:00+09:00",
					TimeZone: "Asia/Tokyo",
				},
				End: domain.DateTime{
					DateTime: "2025-06-01T11:00:00+09:00",
					TimeZone: "Asia/Tokyo",
				},
			},
			mockResponse: &calendar.Event{
				Id:          "event1",
				Summary:     "Test Event",
				Description: "Test Description",
				Start: &calendar.EventDateTime{
					DateTime: "2025-06-01T10:00:00+09:00",
					TimeZone: "Asia/Tokyo",
				},
				End: &calendar.EventDateTime{
					DateTime: "2025-06-01T11:00:00+09:00",
					TimeZone: "Asia/Tokyo",
				},
			},
		},
		{
			name:        "error response",
			calendarID:  "primary",
			event:       &domain.Event{},
			mockError:   errors.New("API error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockCalendarService{
				createEventFunc: func(calID string, e *calendar.Event) (*calendar.Event, error) {
					if calID != tt.calendarID {
						t.Errorf("got calendar ID %q, want %q", calID, tt.calendarID)
					}
					return tt.mockResponse, tt.mockError
				},
			}

			adapter := &GoogleCalendarAdapter{
				service: mock,
			}

			createdEvent, err := adapter.CreateEvent(context.Background(), tt.calendarID, tt.event)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if createdEvent.ID != tt.mockResponse.Id {
				t.Errorf("got event ID %q, want %q", createdEvent.ID, tt.mockResponse.Id)
			}
		})
	}
}
