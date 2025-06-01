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
		// 新しいリミッターを作成してクリーンな状態でテスト
		testLimiter := NewRateLimiter()
		
		// バースト制限（10回）を消費
		for i := 0; i < 10; i++ {
			allowed := testLimiter.Allow()
			assert.True(t, allowed, "バースト制限内のリクエストは許可されるべきです")
		}
		// 制限を超えた場合は Allow が false を返す
		assert.False(t, testLimiter.Allow(), "バースト制限を超えたリクエストは拒否されるべきです")
	})

	t.Run("異常系：コンテキストのキャンセル", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // すぐにキャンセル

		err := limiter.Wait(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})

	t.Run("異常系：レート制限超過とリトライ", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		// 新しいリミッターでバースト制限を超過させる
		testLimiter := NewRateLimiter()
		// バースト制限を消費
		for i := 0; i < 10; i++ {
			testLimiter.Allow()
		}

		start := time.Now()
		err := testLimiter.Wait(ctx)
		elapsed := time.Since(start)

		// タイムアウトまたは正常完了を確認
		if err != nil {
			assert.Contains(t, err.Error(), "context canceled")
		}
		assert.Greater(t, elapsed.Milliseconds(), int64(90), "レート制限による待機が発生していません")
	})
}

func TestRateLimiter_QPS(t *testing.T) {
	limiter := NewRateLimiter()
	ctx := context.Background()

	start := time.Now()
	requestCount := 3 // 少ない回数でテスト

	for i := 0; i < requestCount; i++ {
		err := limiter.Wait(ctx)
		assert.NoError(t, err)
	}

	elapsed := time.Since(start)
	// 最初の10回はバーストで即座に処理され、その後は1秒間隔
	// 3回の場合、最初の3回はバーストで処理されるため、時間はほぼ0秒
	assert.Less(t, elapsed.Seconds(), 1.0, "バースト処理が正しく機能していません")
}

func TestRateLimiter_Reset(t *testing.T) {
	limiter := NewRateLimiter()

	// バースト制限を超過させる  
	for i := 0; i < 10; i++ {
		limiter.Allow()
	}
	assert.False(t, limiter.Allow(), "レート制限が機能していません")

	// リセット実行（バックオフのリセットのみ）
	limiter.Reset()

	// レート制限自体はリセットされないため、まだ制限されている
	assert.False(t, limiter.Allow(), "レート制限は継続しているべきです")
}
