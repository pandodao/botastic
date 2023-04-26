package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/pandodao/botastic/core"
)

type intentRecognition struct{}

type intentRecognitionOptions struct {
	Intents []string
}

func InitIntentRecognition() *intentRecognition {
	return &intentRecognition{}
}

func (m *intentRecognition) Name() string {
	return core.MiddlewareIntentRecognition
}

func (m *intentRecognition) ValidateOptions(opts map[string]any) (any, error) {
	options := &intentRecognitionOptions{}

	val, ok := opts["intents"]
	if ok {
		_val, ok := val.([]interface{})
		if !ok {
			return nil, fmt.Errorf("intents should be an array: %v", val)
		}
		for _, v := range _val {
			intent, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("intents should be an array of string: %v", v)
			}
			options.Intents = append(options.Intents, intent)
		}
	}

	return options, nil
}

func (m *intentRecognition) Process(ctx context.Context, opts any, turn *core.ConvTurn) (string, error) {
	options := opts.(*intentRecognitionOptions)

	prompt := `You will analyze the intent of the request.
You will output the analyze result at the beginning of your response as json format: {"intent": Here is your intent analyze result}
The possible intents should be one of following. If you have no confident about the intent, please use "unknown intent":`

	if len(options.Intents) == 0 {
		return "", nil
	}

	return fmt.Sprintf("%s\n%s", prompt, strings.Join(options.Intents, "\n")), nil
}
