package gpt

import (
	"context"
	"sync"
	"time"

	"github.com/fox-one/pkg/logger"
	gogpt "github.com/sashabaranov/go-gpt3"
)

type Config struct {
	Keys    []string
	Timeout time.Duration
}

type Client struct {
	id           int
	failedCount  int
	successCount int
	inUse        bool
	*gogpt.Client
}

type Handler struct {
	sync.Mutex
	cfg     Config
	index   int
	clients []*Client
}

func New(cfg Config) *Handler {
	h := &Handler{
		cfg:     cfg,
		clients: make([]*Client, len(cfg.Keys)),
	}
	for i, key := range cfg.Keys {
		c := &Client{
			id:     i,
			Client: gogpt.NewClient(key),
		}
		h.clients[i] = c
	}
	return h
}

func (h *Handler) CreateEmbeddings(ctx context.Context, request gogpt.EmbeddingRequest) (gogpt.EmbeddingResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, h.cfg.Timeout)
	defer cancel()

	h.Lock()
	client := h.clients[h.index]
	h.index = (h.index + 1) % len(h.clients)
	h.Unlock()

	request.Model = gogpt.AdaEmbeddingV2
	resp, err := client.CreateEmbeddings(ctx, request)
	if err != nil {
		logger.FromContext(ctx).WithError(err).Errorln("gpt: create embeddings failed")
	}
	return resp, err
}
