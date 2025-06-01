package errors

import (
	"errors"
	"testing"
)

func TestConfigError(t *testing.T) {
	baseErr := errors.New("base error")
	err := NewConfigError("credentials", "missing file", baseErr)

	// エラーメッセージのテスト
	expected := "config error: credentials: missing file: base error"
	if err.Error() != expected {
		t.Errorf("expected error message %q, got %q", expected, err.Error())
	}

	// Unwrapのテスト
	if !errors.Is(err, baseErr) {
		t.Error("expected errors.Is to find base error")
	}
}

func TestOAuthError(t *testing.T) {
	baseErr := errors.New("base error")
	err := NewOAuthError("token refresh", "token expired", baseErr)

	// エラーメッセージのテスト
	expected := "oauth error: token refresh: token expired: base error"
	if err.Error() != expected {
		t.Errorf("expected error message %q, got %q", expected, err.Error())
	}

	// Unwrapのテスト
	if !errors.Is(err, baseErr) {
		t.Error("expected errors.Is to find base error")
	}
}

func TestAPIError(t *testing.T) {
	baseErr := errors.New("base error")
	err := NewAPIError("get calendars", "request failed", 404, baseErr)

	// エラーメッセージのテスト
	expected := "api error: get calendars: request failed (code: 404): base error"
	if err.Error() != expected {
		t.Errorf("expected error message %q, got %q", expected, err.Error())
	}

	// Unwrapのテスト
	if !errors.Is(err, baseErr) {
		t.Error("expected errors.Is to find base error")
	}
}

func TestValidationError(t *testing.T) {
	baseErr := errors.New("base error")
	err := NewValidationError("event_date", "invalid format", baseErr)

	// エラーメッセージのテスト
	expected := "validation error: event_date: invalid format: base error"
	if err.Error() != expected {
		t.Errorf("expected error message %q, got %q", expected, err.Error())
	}

	// Unwrapのテスト
	if !errors.Is(err, baseErr) {
		t.Error("expected errors.Is to find base error")
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "ErrNotFound",
			err:  ErrNotFound,
			want: "requested resource not found",
		},
		{
			name: "ErrInvalidInput",
			err:  ErrInvalidInput,
			want: "invalid input",
		},
		{
			name: "ErrUnauthorized",
			err:  ErrUnauthorized,
			want: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("expected error message %q, got %q", tt.want, tt.err.Error())
			}
		})
	}
}
