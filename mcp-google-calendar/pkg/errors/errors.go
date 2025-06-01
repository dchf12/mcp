// Package errors はアプリケーション固有のエラー型を定義します。
package errors

import (
	"errors"
	"fmt"
)

// 基本的なエラータイプの定義
var (
	// ErrNotFound は要求されたリソースが見つからない場合に使用されます
	ErrNotFound = errors.New("requested resource not found")
	// ErrInvalidInput は入力が無効な場合に使用されます
	ErrInvalidInput = errors.New("invalid input")
	// ErrUnauthorized は認証が必要な場合に使用されます
	ErrUnauthorized = errors.New("unauthorized")
)

// ConfigError は設定関連のエラーを表します
type ConfigError struct {
	Field   string
	Message string
	Err     error
}

func (e *ConfigError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("config error: %s: %s: %v", e.Field, e.Message, e.Err)
	}
	return fmt.Sprintf("config error: %s: %s", e.Field, e.Message)
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}

// NewConfigError は新しい設定エラーを作成します
func NewConfigError(field, message string, err error) error {
	return &ConfigError{
		Field:   field,
		Message: message,
		Err:     err,
	}
}

// OAuthError はOAuth認証関連のエラーを表します
type OAuthError struct {
	Operation string
	Message   string
	Err       error
}

func (e *OAuthError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("oauth error: %s: %s: %v", e.Operation, e.Message, e.Err)
	}
	return fmt.Sprintf("oauth error: %s: %s", e.Operation, e.Message)
}

func (e *OAuthError) Unwrap() error {
	return e.Err
}

// NewOAuthError は新しいOAuth認証エラーを作成します
func NewOAuthError(operation, message string, err error) error {
	return &OAuthError{
		Operation: operation,
		Message:   message,
		Err:       err,
	}
}

// APIError はGoogle Calendar API関連のエラーを表します
type APIError struct {
	Operation string
	Message   string
	Code      int
	Err       error
}

func (e *APIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("api error: %s: %s (code: %d): %v", e.Operation, e.Message, e.Code, e.Err)
	}
	return fmt.Sprintf("api error: %s: %s (code: %d)", e.Operation, e.Message, e.Code)
}

func (e *APIError) Unwrap() error {
	return e.Err
}

// NewAPIError は新しいAPI関連エラーを作成します
func NewAPIError(operation, message string, code int, err error) error {
	return &APIError{
		Operation: operation,
		Message:   message,
		Code:      code,
		Err:       err,
	}
}

// ValidationError は入力検証関連のエラーを表します
type ValidationError struct {
	Field   string
	Message string
	Err     error
}

func (e *ValidationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("validation error: %s: %s: %v", e.Field, e.Message, e.Err)
	}
	return fmt.Sprintf("validation error: %s: %s", e.Field, e.Message)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// NewValidationError は新しい入力検証エラーを作成します
func NewValidationError(field, message string, err error) error {
	return &ValidationError{
		Field:   field,
		Message: message,
		Err:     err,
	}
}
