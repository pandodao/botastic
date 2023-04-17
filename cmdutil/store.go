package cmdutil

import (
	"encoding/json"
	"fmt"

	"github.com/fox-one/mixin-sdk-go"
)

type Keystore struct {
	*mixin.Keystore
	Pin string `json:"pin,omitempty"`
}

func DecodeKeystore(b []byte) (*mixin.Keystore, string, error) {
	var store Keystore
	if err := json.Unmarshal(b, &store); err != nil {
		return nil, "", fmt.Errorf("json decode keystore failed: %w", err)
	}

	return store.Keystore, store.Pin, nil
}
