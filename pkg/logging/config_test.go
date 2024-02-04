package logging

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name    string
		appName Name
		want    *Config
	}{
		{
			name:    "TestNewConfig",
			appName: "TestNewConfig",
			want: &Config{
				appName: "TestNewConfig",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewConfig(tt.appName)
			require.Equal(t, tt.want, got, "NewConfig() = %v, want %v", got, tt.want)
		})
	}
}
