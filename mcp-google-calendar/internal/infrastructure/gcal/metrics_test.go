package gcal

import (
	"context"
	"testing"
	"time"

	"github.com/dch/mcp-google-calendar/internal/domain"
	"github.com/dch/mcp-google-calendar/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

type dummyService struct{}

func (d *dummyService) ListCalendars(ctx context.Context) ([]domain.Calendar, error) {
	// 必要に応じてダミーデータを返す
	return []domain.Calendar{{ID: "dummy", Title: "Dummy"}}, nil
}

func (d *dummyService) CreateEvent(ctx context.Context, calID string, ev *domain.Event) (*domain.Event, error) {
	// エラー検証用に適宜切り替える
	return nil, errors.NewAPIError("create_event", "api_error_for_test", 400, nil)
}

func TestMetrics(t *testing.T) {
	ctx := context.Background()
	adapter := NewWithService(&dummyService{})

	t.Run("APIリクエストのカウント", func(t *testing.T) {
		// リクエスト前のカウント
		initialCount := testutil.ToFloat64(apiRequestsTotal.WithLabelValues("list_calendars"))

		// APIリクエストの実行
		_, _ = adapter.ListCalendars(ctx)

		// リクエスト後のカウント
		finalCount := testutil.ToFloat64(apiRequestsTotal.WithLabelValues("list_calendars"))
		assert.Equal(t, initialCount+1, finalCount, "APIリクエストがカウントされていません")
	})

	t.Run("APIレスポンス時間の記録", func(t *testing.T) {
		// APIリクエストを実行（Histogram に Observe される）
		_, _ = adapter.ListCalendars(ctx)

		// HistogramVec は Gauge/Counter のように直接値を取得できないため、
		// testutil.CollectAndCount でサンプル数を確認する。
		count := testutil.CollectAndCount(apiResponseDuration)
		assert.Greater(t, count, 0, "レスポンス時間メトリクスが収集されていません")
	})

	t.Run("エラーのカウント", func(t *testing.T) {
		// 独立したアダプタでリミッター消費の影響を受けないようにする
		adapterErr := NewWithService(&dummyService{})

		initialCount := testutil.ToFloat64(apiErrorsTotal.WithLabelValues("create_event", "api_error"))

		// 不正なイベントでエラーを発生させる
		_, err := adapterErr.CreateEvent(ctx, "invalid-calendar-id", &domain.Event{})
		assert.Error(t, err)

		finalCount := testutil.ToFloat64(apiErrorsTotal.WithLabelValues("create_event", "api_error"))
		assert.Equal(t, initialCount+1, finalCount, "APIエラーがカウントされていません")
	})

	t.Run("レート制限のカウント", func(t *testing.T) {
		// バースト制限を超えるリクエストを送信
		initialCount := testutil.ToFloat64(rateLimitHitsTotal)

		// レート制限を超えるまでリクエスト
		for i := 0; i < 20; i++ {
			adapter.ListCalendars(ctx)
		}

		// レート制限カウントの確認（少し待機）
		time.Sleep(100 * time.Millisecond)
		finalCount := testutil.ToFloat64(rateLimitHitsTotal)
		assert.Greater(t, finalCount, initialCount, "レート制限のヒットがカウントされていません")
	})
}
