package usecase

import (
	"context"

	"github.com/dch/mcp-google-calendar/internal/domain"
)

// EventReadRepository はイベント読み取り操作のリポジトリインターフェースです
type EventReadRepository interface {
	ListEvents(ctx context.Context, params domain.GetEventsParams) ([]domain.Event, error)
}

// GetEventsUseCase はイベント一覧取得のユースケースです
type GetEventsUseCase struct {
	repo EventReadRepository
}

// NewGetEventsUseCase は GetEventsUseCase のコンストラクタです
func NewGetEventsUseCase(repo EventReadRepository) *GetEventsUseCase {
	return &GetEventsUseCase{
		repo: repo,
	}
}

// Execute はイベント一覧取得を実行します
func (uc *GetEventsUseCase) Execute(ctx context.Context, params domain.GetEventsParams) ([]domain.Event, error) {
	// パラメータのバリデーション
	if err := params.Validate(); err != nil {
		return nil, err
	}

	// デフォルト値の設定
	if params.MaxResults <= 0 {
		params.MaxResults = 250
	}

	// リポジトリからイベント一覧を取得
	events, err := uc.repo.ListEvents(ctx, params)
	if err != nil {
		return nil, err
	}

	// 取得したイベントのバリデーション
	for i := range events {
		if err := events[i].Validate(); err != nil {
			return nil, err
		}
	}

	return events, nil
}
