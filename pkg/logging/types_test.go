package logging

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestName_String(t *testing.T) {
	tests := []struct {
		name string
		n    Name
		want string
	}{
		{
			name: "TestName_String",
			n:    "TestName_String",
			want: "TestName_String",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.n.String()
			require.Equal(t, tt.want, got, "Name.String() = %v, want %v", got, tt.want)
		})
	}
}
