package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLuhn(t *testing.T) {
	tests := []struct {
		name   string
		number int64
		valid  bool
	}{
		{
			name:   "ok",
			number: 1234567897,
			valid:  true,
		},
		{
			name:   "not valid",
			number: 1,
			valid:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, Valid(test.number), test.valid)
		})
	}
}
