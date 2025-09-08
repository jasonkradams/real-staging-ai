package job

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatus_String(t *testing.T) {
	testCases := []struct {
		name     string
		status   Status
		expected string
	}{
		{
			name:     "queued",
			status:   StatusQueued,
			expected: "queued",
		},
		{
			name:     "processing",
			status:   StatusProcessing,
			expected: "processing",
		},
		{
			name:     "completed",
			status:   StatusCompleted,
			expected: "completed",
		},
		{
			name:     "failed",
			status:   StatusFailed,
			expected: "failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.status.String())
		})
	}
}

func TestType_String(t *testing.T) {
	testCases := []struct {
		name     string
		jobType  Type
		expected string
	}{
		{
			name:     "stage image",
			jobType:  TypeStageImage,
			expected: "stage:image",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.jobType.String())
		})
	}
}
