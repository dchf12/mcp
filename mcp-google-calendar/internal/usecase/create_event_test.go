package usecase

import (
	"context"
	"fmt"
	"testing"

	"github.com/dch/mcp-google-calendar/internal/domain"
)

type mockEventRepository struct {
	event *domain.Event
	err   error
}

func (m *mockEventRepository) CreateEvent(ctx context.Context, calendarID string, event *domain.Event) (*domain.Event, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.event, nil
}

func TestCreateEventUseCase_Execute(t *testing.T) {
	validEvent := &domain.Event{
		ID:          "event1",
		Title:       "Test Event",
		Description: "Test Description",
		Start: domain.DateTime{
			DateTime: "2023-01-01T10:00:00Z",
			TimeZone: "UTC",
		},
		End: domain.DateTime{
			DateTime: "2023-01-01T11:00:00Z",
			TimeZone: "UTC",
		},
	}

	tests := []struct {
		name        string
		calendarID  string
		inputEvent  *domain.Event
		repoEvent   *domain.Event
		repoErr     error
		expectedErr bool
	}{
		{
			name:        "successful execution",
			calendarID:  "cal1",
			inputEvent:  validEvent,
			repoEvent:   validEvent,
			repoErr:     nil,
			expectedErr: false,
		},
		{
			name:        "nil event",
			calendarID:  "cal1",
			inputEvent:  nil,
			repoEvent:   nil,
			repoErr:     nil,
			expectedErr: true,
		},
		{
			name:       "invalid event - missing title",
			calendarID: "cal1",
			inputEvent: &domain.Event{
				ID:          "event1",
				Title:       "",
				Description: "Test Description",
				Start: domain.DateTime{
					DateTime: "2023-01-01T10:00:00Z",
					TimeZone: "UTC",
				},
				End: domain.DateTime{
					DateTime: "2023-01-01T11:00:00Z",
					TimeZone: "UTC",
				},
			},
			repoEvent:   nil,
			repoErr:     nil,
			expectedErr: true,
		},
		{
			name:        "repository error",
			calendarID:  "cal1",
			inputEvent:  validEvent,
			repoEvent:   nil,
			repoErr:     fmt.Errorf("repository error"),
			expectedErr: true,
		},
		{
			name:       "invalid created event",
			calendarID: "cal1",
			inputEvent: validEvent,
			repoEvent: &domain.Event{
				ID:    "event1",
				Title: "",
				Start: domain.DateTime{
					DateTime: "2023-01-01T10:00:00Z",
					TimeZone: "UTC",
				},
				End: domain.DateTime{
					DateTime: "2023-01-01T11:00:00Z",
					TimeZone: "UTC",
				},
			},
			repoErr:     nil,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockEventRepository{
				event: tt.repoEvent,
				err:   tt.repoErr,
			}
			uc := NewCreateEventUseCase(repo)

			event, err := uc.Execute(context.Background(), tt.calendarID, tt.inputEvent)

			if tt.expectedErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if event == nil {
				t.Errorf("expected event but got nil")
				return
			}

			if event.ID != tt.repoEvent.ID {
				t.Errorf("expected event ID %s, got %s", tt.repoEvent.ID, event.ID)
			}
			if event.Title != tt.repoEvent.Title {
				t.Errorf("expected event title %s, got %s", tt.repoEvent.Title, event.Title)
			}
		})
	}
}