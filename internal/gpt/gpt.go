package gpt

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/fox-one/pkg/logger"
	gogpt "github.com/sashabaranov/go-gpt3"
)

var ErrTooManyRequests = errors.New("too many requests")

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

func (h *Handler) CreateCompletion(ctx context.Context, request gogpt.CompletionRequest) (gogpt.CompletionResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, h.cfg.Timeout)
	defer cancel()
	h.Lock()
	client := h.clients[h.index]
	h.index = (h.index + 1) % len(h.clients)
	h.Unlock()

	resp, err := client.CreateCompletion(ctx, request)
	if err != nil {
		var perr *gogpt.APIError
		if errors.As(err, &perr) {
			if perr.StatusCode == 429 {
				return resp, ErrTooManyRequests
			}
		}

		var cerr *gogpt.RequestError
		if errors.As(err, &cerr) {
			if cerr.StatusCode == 429 {
				return resp, ErrTooManyRequests
			}
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return resp, ErrTooManyRequests
		}

		logger.FromContext(ctx).WithError(err).Errorln("gpt: create completion failed")
	}
	return resp, err
}

func (h *Handler) CreateChatCompletion(ctx context.Context, request gogpt.ChatCompletionRequest) (gogpt.ChatCompletionResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, h.cfg.Timeout)
	defer cancel()
	h.Lock()
	client := h.clients[h.index]
	h.index = (h.index + 1) % len(h.clients)
	h.Unlock()

	resp, err := client.CreateChatCompletion(ctx, request)
	if err != nil {
		var perr *gogpt.APIError
		if errors.As(err, &perr) {
			if perr.StatusCode == 429 {
				return resp, ErrTooManyRequests
			}
		}

		var cerr *gogpt.RequestError
		if errors.As(err, &cerr) {
			if cerr.StatusCode == 429 {
				return resp, ErrTooManyRequests
			}
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return resp, ErrTooManyRequests
		}

		logger.FromContext(ctx).WithError(err).Errorln("gpt: create chat completion failed")
	}
	return resp, err
}
