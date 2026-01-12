package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/dch/mcp-google-calendar/internal/domain"
	"github.com/dch/mcp-google-calendar/pkg/errors"
)

type mockEventReadRepository struct {
	events []domain.Event
	err    error
}

func (m *mockEventReadRepository) ListEvents(ctx context.Context, params domain.GetEventsParams) ([]domain.Event, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.events, nil
}

func TestGetEventsUseCase_Execute(t *testing.T) {
	// テスト用の基準時刻
	now := time.Now()
	timeMin := now
	timeMax := now.Add(24 * time.Hour)

	// TODO(human): テストケースを定義してください
	// 以下のケースを実装することで、ユースケースの動作を網羅的に検証できます
	tests := []struct {
		name        string
		params      domain.GetEventsParams
		events      []domain.Event
		repoErr     error
		expectedErr bool
	}{
		{
			name: "successful execution",
			params: domain.GetEventsParams{
				CalendarID: "cal1",
				TimeMin:    timeMin,
				TimeMax:    timeMax,
			},
			events: []domain.Event{
				{
					ID: "event1",
					Title: "Event 1",
					Description: "First event",
					Start: domain.DateTime{
						DateTime: now.Add(1 * time.Hour).Format(time.RFC3339),
					},
					End: domain.DateTime{
						DateTime: now.Add(2 * time.Hour).Format(time.RFC3339),
					},
				},
				{
					ID: "event2",
					Title: "Event 2",
					Description: "Second event",
					Start: domain.DateTime{
						DateTime: now.Add(3 * time.Hour).Format(time.RFC3339),
					},
					End: domain.DateTime{
						DateTime: now.Add(4 * time.Hour).Format(time.RFC3339),
					},
				},
			},
			repoErr:     nil,
			expectedErr: false,
		},
		{
			name: "calendar ID missing",
			params: domain.GetEventsParams{
				CalendarID: "",
				TimeMin:    timeMin,
				TimeMax:    timeMax,
			},
			events:      nil,
			repoErr:     nil,
			expectedErr: true,
		},
		{
			name: "repository error",
			params: domain.GetEventsParams{
				CalendarID: "cal1",
				TimeMin:    timeMin,
				TimeMax:    timeMax,
			},
			events:      nil,
			repoErr:     errors.ErrNotFound,
			expectedErr: true,
		},
		{
			name: "invalid event data from repository",
			params: domain.GetEventsParams{
				CalendarID: "cal1",
				TimeMin:    timeMin,
				TimeMax:    timeMax,
			},
			events: []domain.Event{
				{
					ID:    "event1",
					Title: "", // タイトルが空でバリデーションエラーになる
					Start: domain.DateTime{
						DateTime: now.Add(1 * time.Hour).Format(time.RFC3339),
					},
					End: domain.DateTime{
						DateTime: now.Add(2 * time.Hour).Format(time.RFC3339),
					},
				},
			},
			repoErr:     nil,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockEventReadRepository{
				events: tt.events,
				err:    tt.repoErr,
			}
			uc := NewGetEventsUseCase(repo)

			events, err := uc.Execute(context.Background(), tt.params)

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

			if len(events) != len(tt.events) {
				t.Errorf("expected %d events, got %d", len(tt.events), len(events))
			}

			for i, ev := range events {
				if ev.ID != tt.events[i].ID {
					t.Errorf("expected event ID %s, got %s", tt.events[i].ID, ev.ID)
				}
				if ev.Title != tt.events[i].Title {
					t.Errorf("expected event title %s, got %s", tt.events[i].Title, ev.Title)
				}
			}
		})
	}

	// 未使用変数の警告を避けるため
	_ = timeMin
	_ = timeMax
}
