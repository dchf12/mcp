package gcal

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiter(t *testing.T) {
	limiter := NewRateLimiter()

	t.Run("正常系：制限内のリクエスト", func(t *testing.T) {
		ctx := context.Background()
		err := limiter.Wait(ctx)
		assert.NoError(t, err)
	})

	t.Run("正常系：バースト制限のテスト", func(t *testing.T) {
		ctx := context.Background()
		// QPS 回だけ Wait でトークンを消費（最初のトークンは 0 なので Wait が必要）
		for i := 0; i < QPS; i++ {
			assert.NoError(t, limiter.Wait(ctx))
		}
		// 制限を超えた場合は Allow が false を返す
		assert.False(t, limiter.Allow())
	})

	t.Run("異常系：コンテキストのキャンセル", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // すぐにキャンセル

		err := limiter.Wait(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})

	t.Run("異常系：レート制限超過とリトライ", func(t *testing.T) {
		ctx := context.Background()
		limiter := NewRateLimiter()

		// レート制限を超過させる
		for i := 0; i < QPS*2; i++ {
			start := time.Now()
			err := limiter.Wait(ctx)
			assert.NoError(t, err)

			// リトライによる待機時間を確認
			if i >= QPS {
				elapsed := time.Since(start)
				assert.Greater(t, elapsed.Milliseconds(), int64(90), // 少なくとも90ms以上の待機
					"レート制限による待機が発生していません")
			}
		}
	})
}

func TestRateLimiter_QPS(t *testing.T) {
	limiter := NewRateLimiter()
	ctx := context.Background()

	start := time.Now()
	requestCount := QPS + 10 // QPSより多めにリクエスト

	for i := 0; i < requestCount; i++ {
		err := limiter.Wait(ctx)
		assert.NoError(t, err)
	}

	elapsed := time.Since(start)
	// 先頭 1 リクエスト分はバーストで即時処理されるため 1 秒短い
	expectedMinDuration := time.Duration((requestCount-1)/QPS) * time.Second

	assert.GreaterOrEqual(t, elapsed.Seconds(), expectedMinDuration.Seconds(),
		"レート制限が正しく機能していません")
}

func TestRateLimiter_Reset(t *testing.T) {
	limiter := NewRateLimiter()
	ctx := context.Background()

	// レート制限を超過させる
	for i := 0; i < QPS; i++ {
		limiter.Allow()
	}
	assert.False(t, limiter.Allow(), "レート制限が機能していません")

	// リセット後は再度リクエスト可能
	limiter.Reset()

	err := limiter.Wait(ctx)
	assert.NoError(t, err, "リセット後のリクエストが失敗しました")
}
