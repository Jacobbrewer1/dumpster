package logging

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommonLogger(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr error
	}{
		{
			name: "TestCommonLogger",
			cfg: &Config{
				appName: "TestCommonLogger",
			},
			wantErr: nil,
		},
		{
			name:    "TestCommonLogger",
			cfg:     nil,
			wantErr: errors.New("logging config is nil"),
		},
		{
			name: "TestCommonLogger",
			cfg: &Config{
				appName: "",
			},
			wantErr: errors.New("app name is empty"),
		},
		{
			name: "TestCommonLogger",
			cfg: &Config{
				appName: "TestCommonLogger",
			},
			wantErr: nil,
		},
		{
			name: "TestCommonLogger",
			cfg: &Config{
				appName: "TestCommonLogger",
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CommonLogger(tt.cfg)
			require.Equal(t, tt.wantErr, err, "CommonLogger() = %v, want %v", err, tt.wantErr)
		})
	}
}
