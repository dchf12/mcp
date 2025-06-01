package usecase

import (
	"context"

	"github.com/dch/mcp-google-calendar/internal/domain"
)

type CalendarRepository interface {
	ListCalendars(ctx context.Context) ([]domain.Calendar, error)
}

type GetCalendarsUseCase struct {
	repo CalendarRepository
}

func NewGetCalendarsUseCase(repo CalendarRepository) *GetCalendarsUseCase {
	return &GetCalendarsUseCase{
		repo: repo,
	}
}

func (uc *GetCalendarsUseCase) Execute(ctx context.Context) ([]domain.Calendar, error) {
	calendars, err := uc.repo.ListCalendars(ctx)
	if err != nil {
		return nil, err
	}

	for i := range calendars {
		if err := calendars[i].Validate(); err != nil {
			return nil, err
		}
	}

	return calendars, nil
}
