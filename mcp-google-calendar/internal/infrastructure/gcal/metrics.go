package gcal

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// APIリクエスト数のカウンター
	apiRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "google_calendar_api_requests_total",
		Help: "Total number of requests made to Google Calendar API",
	}, []string{"operation"})

	// APIレスポンス時間のヒストグラム
	apiResponseDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "google_calendar_api_response_duration_seconds",
		Help:    "Response time of Google Calendar API requests",
		Buckets: prometheus.DefBuckets,
	}, []string{"operation"})

	// APIエラー数のカウンター
	apiErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "google_calendar_api_errors_total",
		Help: "Total number of API errors",
	}, []string{"operation", "error_type"})

	// レート制限ヒット数のカウンター
	rateLimitHitsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "google_calendar_rate_limit_hits_total",
		Help: "Total number of rate limit hits",
	})
)

// recordAPIRequest は新しいAPIリクエストを記録します
func recordAPIRequest(operation string) {
	apiRequestsTotal.WithLabelValues(operation).Inc()
}

// recordAPIResponseDuration はAPIリクエストの応答時間を記録します
func recordAPIResponseDuration(operation string, duration float64) {
	apiResponseDuration.WithLabelValues(operation).Observe(duration)
}

// recordAPIError はAPIエラーを記録します
func recordAPIError(operation, errorType string) {
	apiErrorsTotal.WithLabelValues(operation, errorType).Inc()
}

// recordRateLimitHit はレート制限のヒットを記録します
func recordRateLimitHit() {
	rateLimitHitsTotal.Inc()
}
