package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"testing"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}
)

func TestValidate(t *testing.T) { //nolint:funlen // табличные тесты могут быть длинными
	tests := []struct {
		name        string
		in          interface{}
		expectedErr error
	}{
		{
			name: "valid user",
			in: User{
				ID:     "123g4567-a15i-12d3-a456-123124254341",
				Name:   "Boris Fet",
				Age:    25,
				Email:  "boris@example.com",
				Role:   "admin",
				Phones: []string{"79991234567", "79991234568"},
			},
			expectedErr: nil,
		},
		{
			name: "invalid user - short ID",
			in: User{
				ID:     "short",
				Age:    25,
				Email:  "boris@example.com",
				Role:   "admin",
				Phones: []string{"79991234567"},
			},
			expectedErr: ErrInvalidLength,
		},
		{
			name: "invalid user - age too young",
			in: User{
				ID:     "123g4567-a15i-12d3-a456-123124254341",
				Age:    15,
				Email:  "boris@example.com",
				Role:   "admin",
				Phones: []string{"79991234567"},
			},
			expectedErr: ErrLessThanMin,
		},
		{
			name: "invalid user - age too old",
			in: User{
				ID:     "123g4567-a15i-12d3-a456-123124254341",
				Age:    60,
				Email:  "boris@example.com",
				Role:   "admin",
				Phones: []string{"79991234567"},
			},
			expectedErr: ErrGreaterThanMax,
		},
		{
			name: "invalid user - bad email",
			in: User{
				ID:     "123g4567-a15i-12d3-a456-123124254341",
				Age:    25,
				Email:  "invalid-email",
				Role:   "admin",
				Phones: []string{"79991234567"},
			},
			expectedErr: ErrRegexpMismatch,
		},
		{
			name: "invalid user - bad role",
			in: User{
				ID:     "123g4567-a15i-12d3-a456-123124254341",
				Age:    25,
				Email:  "boris@example.com",
				Role:   "superuser",
				Phones: []string{"79991234567"},
			},
			expectedErr: ErrNotInSet,
		},
		{
			name: "invalid user - bad phone length",
			in: User{
				ID:     "123g4567-a15i-12d3-a456-123124254341",
				Age:    25,
				Email:  "boris@example.com",
				Role:   "admin",
				Phones: []string{"799912345"},
			},
			expectedErr: ErrInvalidLength,
		},
		{
			name: "valid app",
			in: App{
				Version: "1.0.0",
			},
			expectedErr: nil,
		},
		{
			name: "invalid app - version too long",
			in: App{
				Version: "1.0.0.0",
			},
			expectedErr: ErrInvalidLength,
		},
		{
			name: "valid token - no validation tags",
			in: Token{
				Header:    []byte("header"),
				Payload:   []byte("payload"),
				Signature: []byte("signature"),
			},
			expectedErr: nil,
		},
		{
			name: "valid response",
			in: Response{
				Code: 200,
				Body: "OK",
			},
			expectedErr: nil,
		},
		{
			name: "invalid response - bad code",
			in: Response{
				Code: 403,
				Body: "Forbidden",
			},
			expectedErr: ErrNotInSet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt := tt
			t.Parallel()

			err := Validate(tt.in)

			// не ожидается ошибки и она не получена
			if tt.expectedErr == nil {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
				return
			}

			// ожидали ошибку, но её нет - тест провален
			if err == nil {
				t.Errorf("expected error %v, got nil", tt.expectedErr)
				return
			}

			// Проверка типа ошибки
			var validationErrs ValidationErrors
			if !errors.As(err, &validationErrs) {
				t.Errorf("expected ValidationErrors, got: %v", err)
				return
			}

			// Проверяем, что хотя бы одна ошибка валидации соответствует ожидаемой
			found := false
			for _, vErr := range validationErrs {
				if errors.Is(vErr.Err, tt.expectedErr) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("expected error %v in validation errors, got: %v", tt.expectedErr, validationErrs)
			}
		})
	}
}

func TestValidate_MultipleErrors(t *testing.T) {
	// Тест на накопление нескольких ошибок валидации
	user := User{
		ID:     "short",
		Age:    15,
		Email:  "bad-email",
		Role:   "hacker",
		Phones: []string{"123"},
	}

	err := Validate(user)
	if err == nil {
		t.Fatal("expected validation errors, got nil")
	}

	var validationErrs ValidationErrors
	if !errors.As(err, &validationErrs) {
		t.Fatalf("expected ValidationErrors, got: %v", err)
	}

	// Должно быть несколько ошибок
	if len(validationErrs) < 2 {
		t.Errorf("expected multiple validation errors, got %d: %v", len(validationErrs), validationErrs)
	}

	// Проверяем что есть разные типы ошибок
	hasLengthErr := false
	hasMinErr := false
	for _, vErr := range validationErrs {
		if errors.Is(vErr.Err, ErrInvalidLength) {
			hasLengthErr = true
		}
		if errors.Is(vErr.Err, ErrLessThanMin) {
			hasMinErr = true
		}
	}

	if !hasLengthErr {
		t.Error("expected ErrInvalidLength in validation errors")
	}
	if !hasMinErr {
		t.Error("expected ErrLessThanMin in validation errors")
	}
}

func TestValidate_CombinedValidators(t *testing.T) {
	type TestStruct struct {
		Value int `validate:"min:10|max:20"`
	}

	tests := []struct {
		name        string
		value       int
		expectedErr error
	}{
		{
			name:        "valid value in range",
			value:       15,
			expectedErr: nil,
		},
		{
			name:        "value too small",
			value:       5,
			expectedErr: ErrLessThanMin,
		},
		{
			name:        "value too large",
			value:       25,
			expectedErr: ErrGreaterThanMax,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := TestStruct{Value: tt.value}
			err := Validate(s)

			if tt.expectedErr == nil {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
				return
			}

			if err == nil {
				t.Errorf("expected error %v, got nil", tt.expectedErr)
				return
			}

			var validationErrs ValidationErrors
			if !errors.As(err, &validationErrs) {
				t.Errorf("expected ValidationErrors, got: %v", err)
				return
			}

			found := false
			for _, vErr := range validationErrs {
				if errors.Is(vErr.Err, tt.expectedErr) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("expected error %v, got: %v", tt.expectedErr, validationErrs)
			}
		})
	}
}
