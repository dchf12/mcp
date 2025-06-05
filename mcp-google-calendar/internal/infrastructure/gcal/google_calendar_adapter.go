package gcal

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"github.com/dch/mcp-google-calendar/internal/domain"
	"github.com/dch/mcp-google-calendar/pkg/errors"
)

type CalendarService interface {
	ListCalendars(ctx context.Context) ([]domain.Calendar, error)
	CreateEvent(ctx context.Context, calID string, ev *domain.Event) (*domain.Event, error)
}

type GoogleCalendarAdapter struct {
	service CalendarService
	limiter *RateLimiter
}

// NewWithService はテスト用のコンストラクタ。
// 実際の Google API を呼び出す代わりに任意の CalendarService を注入できる。
func NewWithService(svc CalendarService) *GoogleCalendarAdapter {
	return &GoogleCalendarAdapter{
		service: svc,
		limiter: NewRateLimiter(),
	}
}

func New(ctx context.Context, opts ...option.ClientOption) (*GoogleCalendarAdapter, error) {
	raw, err := calendar.NewService(ctx, opts...)
	if err != nil {
		return nil, errors.NewAPIError("new_service", "failed to create calendar service", 500, err)
	}
	return &GoogleCalendarAdapter{
		service: &googleCalendarService{raw: raw},
		limiter: NewRateLimiter(),
	}, nil
}

func (a *GoogleCalendarAdapter) ListCalendars(ctx context.Context) ([]domain.Calendar, error) {
	start := time.Now()
	operation := "list_calendars"
	recordAPIRequest(operation)

	// 速い判定のため Allow() で即座に残トークンを確認
	if !a.limiter.Allow() {
		recordRateLimitHit()
		recordAPIError(operation, "rate_limit")
		return nil, RateLimitExceededError
	}

	cals, err := a.service.ListCalendars(ctx)
	recordAPIResponseDuration(operation, time.Since(start).Seconds())

	if err != nil {
		recordAPIError(operation, "api_error")
		return nil, err
	}

	return cals, nil
}

func (a *GoogleCalendarAdapter) CreateEvent(ctx context.Context, calendarID string, event *domain.Event) (*domain.Event, error) {
	start := time.Now()
	operation := "create_event"
	recordAPIRequest(operation)

	if !a.limiter.Allow() {
		recordRateLimitHit()
		recordAPIError(operation, "rate_limit")
		return nil, RateLimitExceededError
	}

	ev, err := a.service.CreateEvent(ctx, calendarID, event)
	recordAPIResponseDuration(operation, time.Since(start).Seconds())

	if err != nil {
		recordAPIError(operation, "api_error")
		return nil, err
	}

	return ev, nil
}

// googleCalendarService は google カレンダー API を直接呼び出し、
// CalendarService インターフェースを実装する内部用ラッパーです。
type googleCalendarService struct {
	raw *calendar.Service
}

var _ CalendarService = (*googleCalendarService)(nil)

// RateLimitExceededError is returned when the adapter hits its internal rate limit.
var RateLimitExceededError = errors.NewAPIError("rate_limit", "rate limit exceeded", 429, nil)

func (g *googleCalendarService) ListCalendars(ctx context.Context) ([]domain.Calendar, error) {
	list, err := g.raw.CalendarList.List().Context(ctx).Do()
	if err != nil {
		return nil, errors.NewAPIError("list_calendars", "failed to list calendars", 500, err)
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
		return nil, errors.NewValidationError("event", "event cannot be nil", nil)
	}

	if calID == "" {
		return nil, errors.NewValidationError("calendar_id", "calendar ID cannot be empty", nil)
	}

	// カレンダーIDの存在を事前チェック
	if err := g.validateCalendarAccess(ctx, calID); err != nil {
		return nil, err
	}

	if err := ev.Validate(); err != nil {
		return nil, errors.NewValidationError("event", "invalid event data", err)
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

	created, err := g.raw.Events.Insert(calID, gcalEv).Context(ctx).Do()
	if err != nil {
		// より詳細なエラー情報を提供
		errorMsg := fmt.Sprintf("failed to create event in calendar '%s': %v", calID, err)

		// Google API の 404 エラーの場合、特別なメッセージを追加
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "notFound") {
			errorMsg = fmt.Sprintf("calendar '%s' not found or access denied. Please check the calendar ID and permissions", calID)
		}

		return nil, errors.NewAPIError("create_event", errorMsg, 500, err)
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
		Location:  &created.Location,
		Attendees: getEventAttendees(created.Attendees),
	}, nil
}

func getEventAttendees(attendees []*calendar.EventAttendee) []string {
	emails := make([]string, len(attendees))
	for i, a := range attendees {
		emails[i] = a.Email
	}
	return emails
}

// validateCalendarAccess はカレンダーへのアクセス権限を事前チェックします
func (g *googleCalendarService) validateCalendarAccess(ctx context.Context, calID string) error {
	_, err := g.raw.Calendars.Get(calID).Context(ctx).Do()
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "notFound") {
			return errors.NewValidationError("calendar_id",
				fmt.Sprintf("calendar '%s' not found or access denied. Please verify the calendar ID and ensure you have access to it", calID), err)
		}
		return errors.NewAPIError("validate_calendar", fmt.Sprintf("failed to validate calendar access: %v", err), 500, err)
	}
	return nil
}
