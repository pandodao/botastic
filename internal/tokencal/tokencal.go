package tokencal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Handler struct {
	addr string
}

func New(addr string) *Handler {
	return &Handler{
		addr: strings.TrimRight(addr, "/"),
	}
}

func (h *Handler) GetToken(ctx context.Context, s string) (int, error) {
	postBody := struct {
		S string `json:"s"`
	}{
		S: s,
	}
	data, _ := json.Marshal(postBody)
	req, err := http.NewRequestWithContext(ctx, "POST", h.addr+"/token", bytes.NewBuffer(data))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	var result struct {
		Token int `json:"token"`
	}

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	fmt.Println(string(respData))
	if err := json.Unmarshal(respData, &result); err != nil {
		return 0, err
	}

	return result.Token, nil
}
