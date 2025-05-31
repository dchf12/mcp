package usecase

import (
	"context"
	"fmt"
	"testing"

	"github.com/dch/mcp-google-calendar/internal/domain"
)

type mockCalendarRepository struct {
	calendars []domain.Calendar
	err       error
}

func (m *mockCalendarRepository) ListCalendars(ctx context.Context) ([]domain.Calendar, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.calendars, nil
}

func TestGetCalendarsUseCase_Execute(t *testing.T) {
	tests := []struct {
		name        string
		calendars   []domain.Calendar
		repoErr     error
		expectedErr bool
	}{
		{
			name: "successful execution",
			calendars: []domain.Calendar{
				{ID: "cal1", Title: "Calendar 1", Description: "First calendar", TimeZone: "UTC"},
				{ID: "cal2", Title: "Calendar 2", Description: "Second calendar", TimeZone: "UTC"},
			},
			repoErr:     nil,
			expectedErr: false,
		},
		{
			name:        "repository error",
			calendars:   nil,
			repoErr:     fmt.Errorf("repository error"),
			expectedErr: true,
		},
		{
			name: "validation error - missing ID",
			calendars: []domain.Calendar{
				{ID: "", Title: "Calendar 1", Description: "First calendar", TimeZone: "UTC"},
			},
			repoErr:     nil,
			expectedErr: true,
		},
		{
			name: "validation error - missing title",
			calendars: []domain.Calendar{
				{ID: "cal1", Title: "", Description: "First calendar", TimeZone: "UTC"},
			},
			repoErr:     nil,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCalendarRepository{
				calendars: tt.calendars,
				err:       tt.repoErr,
			}
			uc := NewGetCalendarsUseCase(repo)

			calendars, err := uc.Execute(context.Background())

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

			if len(calendars) != len(tt.calendars) {
				t.Errorf("expected %d calendars, got %d", len(tt.calendars), len(calendars))
			}

			for i, cal := range calendars {
				if cal.ID != tt.calendars[i].ID {
					t.Errorf("expected calendar ID %s, got %s", tt.calendars[i].ID, cal.ID)
				}
				if cal.Title != tt.calendars[i].Title {
					t.Errorf("expected calendar title %s, got %s", tt.calendars[i].Title, cal.Title)
				}
			}
		})
	}
}