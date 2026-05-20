package service

import (
	"testing"
)

func TestExtractJSONBlock(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean braces",
			input:    `{"nationalId": "123"}`,
			expected: `{"nationalId": "123"}`,
		},
		{
			name:     "markdown json block",
			input:    "```json\n{\"nationalId\": \"123\"}\n```",
			expected: `{"nationalId": "123"}`,
		},
		{
			name:     "markdown text block",
			input:    "```\n{\"nationalId\": \"123\"}\n```",
			expected: `{"nationalId": "123"}`,
		},
		{
			name:     "surrounding text",
			input:    `Here is the result: {"nationalId": "123"} Hope this helps!`,
			expected: `{"nationalId": "123"}`,
		},
		{
			name:     "no braces",
			input:    `nationalId: 123`,
			expected: `nationalId: 123`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractJSONBlock(tt.input)
			if got != tt.expected {
				t.Errorf("extractJSONBlock(%q) = %q, expected %q", tt.input, got, tt.expected)
			}
		})
	}
}
