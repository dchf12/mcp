package usecase

import (
	"context"
	"fmt"

	"github.com/dch/mcp-google-calendar/internal/domain"
)

type EventRepository interface {
	CreateEvent(ctx context.Context, calendarID string, event *domain.Event) (*domain.Event, error)
}

type CreateEventUseCase struct {
	repo EventRepository
}

func NewCreateEventUseCase(repo EventRepository) *CreateEventUseCase {
	return &CreateEventUseCase{
		repo: repo,
	}
}

func (uc *CreateEventUseCase) Execute(ctx context.Context, calendarID string, event *domain.Event) (*domain.Event, error) {
	if event == nil {
		return nil, fmt.Errorf("event cannot be nil")
	}

	if err := event.Validate(); err != nil {
		return nil, err
	}

	createdEvent, err := uc.repo.CreateEvent(ctx, calendarID, event)
	if err != nil {
		return nil, err
	}

	if err := createdEvent.Validate(); err != nil {
		return nil, err
	}

	return createdEvent, nil
}
