package gcal

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dch/mcp-google-calendar/internal/domain"
)

type mockCalendarService struct {
	listResp   []domain.Calendar
	insertResp *domain.Event
	listErr    error
	insertErr  error
}

func (m *mockCalendarService) ListCalendars(ctx context.Context) ([]domain.Calendar, error) {
	return m.listResp, m.listErr
}
func (m *mockCalendarService) CreateEvent(ctx context.Context, calendarID string, ev *domain.Event) (*domain.Event, error) {
	return m.insertResp, m.insertErr
}

var _ CalendarService = (*mockCalendarService)(nil)

func TestGoogleCalendarAdapter_ListCalendars(t *testing.T) {
	mockSvc := &mockCalendarService{
		listResp: []domain.Calendar{
			{ID: "primary", Title: "メイン", Description: "説明", TimeZone: "Asia/Tokyo"},
		},
	}

	adapter := &GoogleCalendarAdapter{
		service: mockSvc,
		limiter: NewRateLimiter(),
	}

	got, err := adapter.ListCalendars(context.Background())
	require.NoError(t, err)
	require.Equal(t, []domain.Calendar{
		{
			ID:          "primary",
			Title:       "メイン",
			Description: "説明",
			TimeZone:    "Asia/Tokyo",
		},
	}, got)
}

func TestGoogleCalendarAdapter_CreateEvent(t *testing.T) {
	wantEvent := &domain.Event{
		ID:    "evt123",
		Title: "会議",
		Start: domain.DateTime{DateTime: "2025-06-01T10:00:00+09:00"},
		End:   domain.DateTime{DateTime: "2025-06-01T11:00:00+09:00"},
	}
	mockSvc := &mockCalendarService{
		insertResp: &domain.Event{
			ID:    wantEvent.ID,
			Title: wantEvent.Title,
			Start: wantEvent.Start,
			End:   wantEvent.End,
		},
	}
	adapter := &GoogleCalendarAdapter{
		service: mockSvc,
		limiter: NewRateLimiter(),
	}

	got, err := adapter.CreateEvent(context.Background(), "primary", &domain.Event{
		Title: wantEvent.Title,
		Start: wantEvent.Start,
		End:   wantEvent.End,
	})
	require.NoError(t, err)
	require.Equal(t, wantEvent, got)
}
