package gcal

import (
	"context"
	"fmt"

	"github.com/dch/mcp-google-calendar/internal/domain"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// GoogleCalendarAdapter は Google Calendar API とのインタラクションを担当します
type GoogleCalendarAdapter struct {
	service *calendar.Service
}

// New は新しい GoogleCalendarAdapter インスタンスを作成します
func New(ctx context.Context, opts ...option.ClientOption) (*GoogleCalendarAdapter, error) {
	service, err := calendar.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service: %w", err)
	}

	return &GoogleCalendarAdapter{
		service: service,
	}, nil
}

// ListCalendars は利用可能なカレンダーの一覧を取得します
func (a *GoogleCalendarAdapter) ListCalendars(ctx context.Context) ([]domain.Calendar, error) {
	list, err := a.service.CalendarList.List().Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list calendars: %w", err)
	}

	calendars := make([]domain.Calendar, len(list.Items))
	for i, item := range list.Items {
		calendars[i] = domain.Calendar{
			ID:          item.Id,
			Title:       item.Summary,
			Description: item.Description,
			TimeZone:    item.TimeZone,
		}
	}

	return calendars, nil
}

// CreateEvent は指定したカレンダーに新しいイベントを作成します
func (a *GoogleCalendarAdapter) CreateEvent(ctx context.Context, calendarID string, event *domain.Event) (*domain.Event, error) {
	if event == nil {
		return nil, fmt.Errorf("event cannot be nil")
	}

	calendarEvent := &calendar.Event{
		Summary:     event.Title,
		Description: event.Description,
		Start: &calendar.EventDateTime{
			DateTime: event.Start.DateTime,
			Date:     event.Start.Date,
			TimeZone: event.Start.TimeZone,
		},
		End: &calendar.EventDateTime{
			DateTime: event.End.DateTime,
			Date:     event.End.Date,
			TimeZone: event.End.TimeZone,
		},
	}

	if event.Location != nil {
		calendarEvent.Location = *event.Location
	}

	if len(event.Attendees) > 0 {
		attendees := make([]*calendar.EventAttendee, len(event.Attendees))
		for i, email := range event.Attendees {
			attendees[i] = &calendar.EventAttendee{
				Email: email,
			}
		}
		calendarEvent.Attendees = attendees
	}

	created, err := a.service.Events.Insert(calendarID, calendarEvent).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	return &domain.Event{
		ID:          created.Id,
		Title:       created.Summary,
		Description: created.Description,
		Start: domain.DateTime{
			DateTime: created.Start.DateTime,
			Date:     created.Start.Date,
			TimeZone: created.Start.TimeZone,
		},
		End: domain.DateTime{
			DateTime: created.End.DateTime,
			Date:     created.End.Date,
			TimeZone: created.End.TimeZone,
		},
	}, nil
}

// GetScopes は必要な OAuth スコープを返します
func (a *GoogleCalendarAdapter) GetScopes() []string {
	return []string{
		calendar.CalendarReadonlyScope,
		calendar.CalendarEventsScope,
	}
}
