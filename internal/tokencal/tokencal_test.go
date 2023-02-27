package tokencal

import (
	"context"
	"os"
	"testing"
)

func TestTokenCal(t *testing.T) {
	addr := os.Getenv("TOKENCAL_ADDR")
	if addr == "" {
		t.Skip("TOKENCAL_ADDR is not set")
	}

	h := New(addr)
	table := []struct {
		name string
		s    string
		want int
	}{
		{
			name: "test string",
			s:    "test",
			want: 1,
		},
		{
			name: "chineese string",
			s:    "你好",
			want: 4,
		},
	}

	for _, c := range table {
		t.Run(c.name, func(t *testing.T) {
			r, err := h.GetToken(context.Background(), c.s)
			if err != nil {
				t.Errorf("GetToken(%s) error: %v", c.s, err)
			}

			if r != c.want {
				t.Errorf("GetToken(%s) = %d, want %d", c.s, r, c.want)
			}
		})
	}
}
