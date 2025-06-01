package gcal

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v5"
	"golang.org/x/time/rate"

	"github.com/dch/mcp-google-calendar/pkg/errors"
)

// APIクォータに基づくレート制限の定義
// https://developers.google.com/calendar/api/guides/quota
const (
	// 1日あたりのリクエスト制限：1,000,000回
	DailyQuota = 1_000_000

	// 100秒あたりのユーザーごとのリクエスト制限：100回
	// リクエスト制限を1秒あたりに換算すると1回
	QPS = 1
)

// RateLimiter はGoogle Calendar APIのレートリミット制御を行います
type RateLimiter struct {
	limiter *rate.Limiter
	backoff backoff.BackOff
}

// NewRateLimiter は新しいレートリミッターを作成します
func NewRateLimiter() *RateLimiter {
	// 1秒あたり1リクエストのレートリミッターを設定
	// バースト値は10に設定して、短時間の突発的なリクエストに対応
	limiter := rate.NewLimiter(rate.Limit(QPS), 10)

	// 指数バックオフの設定
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = 100 * time.Millisecond
	expBackoff.MaxInterval = 10 * time.Second
	expBackoff.Multiplier = 2.0
	expBackoff.RandomizationFactor = 0.2 // ±20%のジッター

	return &RateLimiter{
		limiter: limiter,
		backoff: expBackoff,
	}
}

// Wait は次のリクエストが許可されるまで待機します
func (r *RateLimiter) Wait(ctx context.Context) error {
	// まずレートリミッターで待機を試みる
	if err := r.limiter.Wait(ctx); err == nil {
		return nil
	}

	if ctx.Err() != nil {
		// コンテキストのキャンセルやタイムアウトの場合はエラーを返す
		return errors.NewAPIError(
			"rate_limit",
			"context canceled",
			429,
			ctx.Err(),
		)
	}

	// レート制限エラーの場合はバックオフを使用
	ticker := backoff.NewTicker(r.backoff)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return errors.NewAPIError(
				"rate_limit",
				"context canceled",
				429,
				ctx.Err(),
			)
		case <-ticker.C:
			if err := r.limiter.Wait(ctx); err == nil {
				return nil
			}
		}
	}
}

// Allow は現在のリクエストが制限を超えていないか確認します
func (r *RateLimiter) Allow() bool {
	return r.limiter.Allow()
}

// Reset はバックオフの状態をリセットします
func (r *RateLimiter) Reset() {
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = 100 * time.Millisecond
	expBackoff.MaxInterval = 10 * time.Second
	expBackoff.Multiplier = 2.0
	expBackoff.RandomizationFactor = 0.2
	r.backoff = expBackoff
}
