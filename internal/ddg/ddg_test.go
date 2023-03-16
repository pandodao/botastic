package ddg

import (
	"context"
	"os"
	"testing"
)

func TestWebSearch(t *testing.T) {
	if os.Getenv("DDG_TEST") != "1" {
		t.Skip("DDG_TEST not set")
	}

	ctx := context.Background()
	result, err := WebSearch(ctx, "golang", 3)
	if err != nil {
		t.Fatal(err)
	}
	text, err := result.Text()
	if err != nil {
		t.Error(err)
	}
	t.Log(text)
}
