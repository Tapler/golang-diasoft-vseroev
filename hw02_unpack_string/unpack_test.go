package hw02unpackstring

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnpack(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{input: "a4bc2d5e", expected: "aaaabccddddde"},
		{input: "abccd", expected: "abccd"},
		{input: "", expected: ""},
		{input: "aaa0b", expected: "aab"},
		// uncomment if task with asterisk completed
		// {input: `qwe\4\5`, expected: `qwe45`},
		// {input: `qwe\45`, expected: `qwe44444`},
		// {input: `qwe\\5`, expected: `qwe\\\\\`},
		// {input: `qwe\\\3`, expected: `qwe\3`},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			result, err := Unpack(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestUnpackInvalidString(t *testing.T) {
	invalidStrings := []string{"3abc", "45", "aaa10b"}
	for _, tc := range invalidStrings {
		tc := tc
		t.Run(tc, func(t *testing.T) {
			_, err := Unpack(tc)
			require.Truef(t, errors.Is(err, ErrInvalidString), "actual error %q", err)
		})
	}
}

// Дополнительные тесты для улучшения покрытия.
func TestUnpackAdditional(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "single char", input: "a", expected: "a"},
		{name: "single char with count", input: "a3", expected: "aaa"},
		{name: "multiple zeros", input: "a0b0c", expected: "c"},
		{name: "special chars", input: "!2@3#1", expected: "!!@@@#"},
		{name: "newline", input: "d\n5abc", expected: "d\n\n\n\n\nabc"},
		{name: "tab", input: "t\t2x", expected: "t\t\tx"},
		{name: "unicode", input: "п2р3и1в", expected: "ппрррив"},
		{name: "emoji", input: "😀2🎉3", expected: "😀😀🎉🎉🎉"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result, err := Unpack(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, result)
		})
	}
}

// Дополнительные тесты на некорректные строки.
func TestUnpackInvalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "starts with zero", input: "0abc"},
		{name: "triple digit", input: "a123"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := Unpack(tc.input)
			require.Truef(t, errors.Is(err, ErrInvalidString), "actual error %q", err)
		})
	}
}
