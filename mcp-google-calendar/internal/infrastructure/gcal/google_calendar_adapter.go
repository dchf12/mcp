package gcal

import (
	"context"
	"fmt"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"github.com/dch/mcp-google-calendar/internal/domain"
)

type CalendarService interface {
	ListCalendars(ctx context.Context) ([]domain.Calendar, error)
	CreateEvent(ctx context.Context, calID string, ev *domain.Event) (*domain.Event, error)
}

type GoogleCalendarAdapter struct {
	service CalendarService
}

func New(ctx context.Context, opts ...option.ClientOption) (*GoogleCalendarAdapter, error) {
	raw, err := calendar.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service: %w", err)
	}
	return &GoogleCalendarAdapter{
		service: &googleCalendarService{raw: raw},
	}, nil
}

func (a *GoogleCalendarAdapter) ListCalendars(ctx context.Context) ([]domain.Calendar, error) {
	return a.service.ListCalendars(ctx)
}

func (a *GoogleCalendarAdapter) CreateEvent(ctx context.Context, calendarID string, event *domain.Event) (*domain.Event, error) {
	return a.service.CreateEvent(ctx, calendarID, event)
}

// googleCalendarService は google カレンダー API を直接呼び出し、
// CalendarService インターフェースを実装する内部用ラッパーです。
type googleCalendarService struct {
	raw *calendar.Service
}

var _ CalendarService = (*googleCalendarService)(nil)

func (g *googleCalendarService) ListCalendars(ctx context.Context) ([]domain.Calendar, error) {
	list, err := g.raw.CalendarList.List().Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list calendars: %w", err)
	}
	cals := make([]domain.Calendar, len(list.Items))
	for i, item := range list.Items {
		cals[i] = domain.Calendar{
			ID:          item.Id,
			Title:       item.Summary,
			Description: item.Description,
			TimeZone:    item.TimeZone,
		}
	}
	return cals, nil
}

func (g *googleCalendarService) CreateEvent(ctx context.Context, calID string, ev *domain.Event) (*domain.Event, error) {
	if ev == nil {
		return nil, fmt.Errorf("event cannot be nil")
	}

	gcalEv := &calendar.Event{
		Summary:     ev.Title,
		Description: ev.Description,
		Start: &calendar.EventDateTime{
			DateTime: ev.Start.DateTime,
			Date:     ev.Start.Date,
			TimeZone: ev.Start.TimeZone,
		},
		End: &calendar.EventDateTime{
			DateTime: ev.End.DateTime,
			Date:     ev.End.Date,
			TimeZone: ev.End.TimeZone,
		},
	}

	if ev.Location != nil {
		gcalEv.Location = *ev.Location
	}

	if len(ev.Attendees) > 0 {
		attendees := make([]*calendar.EventAttendee, len(ev.Attendees))
		for i, email := range ev.Attendees {
			attendees[i] = &calendar.EventAttendee{Email: email}
		}
		gcalEv.Attendees = attendees
	}

	created, err := g.raw.Events.Insert(calID, gcalEv).Do()
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
