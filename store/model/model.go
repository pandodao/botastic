package model

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/pandodao/botastic/core"
	"gopkg.in/yaml.v2"
)

type (
	store struct {
		modelMap map[string]*core.Model
	}
)

//go:embed models.yaml
var modelsConfig string

func New() *store {
	modelMap, err := LoadModels()
	if err != nil {
		panic(err)
	}
	return &store{
		modelMap: modelMap,
	}
}

func LoadModels() (map[string]*core.Model, error) {
	models := []*core.Model{}
	modelMap := make(map[string]*core.Model)

	if err := yaml.Unmarshal([]byte(modelsConfig), &models); err != nil {
		return nil, err
	}

	for _, m := range models {
		key := fmt.Sprintf("%s:%s", m.Provider, m.ProviderModel)
		modelMap[key] = m

		if m.Provider == core.ModelProviderOpenAI {
			switch m.ProviderModel {
			case "gpt-4", "gpt-4-32k", "gpt-4-0314", "gpt-4-32k-0314", "gpt-3.5-turbo", "gpt-3.5-turbo-0301":
				m.Props.IsOpenAIChatModel = true
			case "text-davinci-003":
				m.Props.IsOpenAICompletionModel = true
			case "text-embedding-ada-002":
				m.Props.IsOpenAIEmbeddingModel = true
			}
		}

		// backward compatibility
		if m.Provider == core.ModelProviderOpenAI {
			modelMap[m.ProviderModel] = m
		}
	}

	return modelMap, nil
}

func (s *store) GetModel(ctx context.Context, name string) (*core.Model, error) {
	if model, ok := s.modelMap[name]; ok {
		return model, nil
	}
	return nil, core.ErrInvalidModel
}
