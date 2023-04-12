package tiktoken

import (
	"fmt"
	"strings"
	"sync"

	"github.com/pandodao/botastic/core"
	"github.com/pkoukk/tiktoken-go"
)

type Handler struct {
	mutex           sync.Mutex
	defaultTiktoken *tiktoken.Tiktoken
	encodingMap     map[string]*tiktoken.Tiktoken
}

func Init() (*Handler, error) {
	// default encoding
	encoding := "cl100k_base"
	t, err := tiktoken.GetEncoding(encoding)
	if err != nil {
		return nil, err
	}

	h := &Handler{
		defaultTiktoken: t,
		encodingMap: map[string]*tiktoken.Tiktoken{
			encoding: t,
		},
	}
	return h, nil
}

func (h *Handler) CalToken(provider, model, text string) (int, error) {
	if provider == core.ModelProviderOpenAI {
		encodingName, ok := tiktoken.MODEL_TO_ENCODING[model]
		if !ok {
			return 0, fmt.Errorf("invalid openai model: %s", model)
		}

		h.mutex.Lock()
		te, ok := h.encodingMap[encodingName]
		h.mutex.Unlock()
		if ok {
			return len(te.Encode(text, nil, nil)), nil
		}

		h.mutex.Lock()
		defer h.mutex.Unlock()

		te, err := tiktoken.GetEncoding(encodingName)
		if err != nil {
			return 0, err
		}
		h.encodingMap[encodingName] = te
		return len(te.Encode(text, nil, nil)), nil
	}

	// use default encoding for other providers
	return len(h.defaultTiktoken.Encode(text, nil, nil)), nil
}

func (h *Handler) CalTokenByName(name, text string) (int, error) {
	ss := strings.Split(name, ":")
	if len(ss) != 2 {
		return 0, fmt.Errorf("invalid name: %s", name)
	}

	return h.CalToken(ss[0], ss[1], text)
}
