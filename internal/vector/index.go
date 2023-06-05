package vector

import (
	"context"
	"fmt"
	"sort"

	"github.com/pandodao/botastic/api"
	"github.com/pandodao/botastic/internal/utils"
	"github.com/pandodao/botastic/models"
	"github.com/pandodao/botastic/pkg/llms"
	llmsapi "github.com/pandodao/botastic/pkg/llms/api"
	"github.com/pandodao/botastic/storage"
	"go.uber.org/zap"
)

type IndexHandler struct {
	sh     *storage.Handler
	vs     Storage
	llmsh  *llms.Handler
	logger *zap.Logger
}

func NewIndexHandler(vs Storage, sh *storage.Handler, llmsh *llms.Handler, logger *zap.Logger) *IndexHandler {
	return &IndexHandler{
		sh:     sh,
		vs:     vs,
		llmsh:  llmsh,
		logger: logger.Named("index_handler"),
	}
}

type IndexNotFoundError struct {
	ID uint
}

func (e *IndexNotFoundError) Error() string {
	return fmt.Sprintf("index with id %d not found", e.ID)
}

func (h *IndexHandler) SearchIndexes(ctx context.Context, embeddingModel string, keyword string, groupKey string, limit int) ([]*api.Index, error) {
	m, err := h.llmsh.GetEmbeddingModel(embeddingModel)
	if err != nil {
		return nil, fmt.Errorf("failed to get embedding model %s: %w", embeddingModel, err)
	}

	embeddingReq := llmsapi.CreateEmbeddingRequest{
		Input: []string{keyword},
	}
	embeddingResp, err := m.CreateEmbedding(ctx, embeddingReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding: %w", err)
	}
	embeddingData := embeddingResp.Data[0].Embedding

	result := make([]*api.Index, 0, limit)
	if h.vs == nil {
		indexes, err := h.sh.GetIndexesByGroupKey(ctx, groupKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get indexes by group key: %w", err)
		}
		scoreMap := make(map[float64]*models.Index, len(indexes))
		scores := make([]float64, 0, len(indexes))
		for _, index := range indexes {
			score := utils.CosineSimilarity(embeddingData, index.Vector)
			scores = append(scores, score)
			index.Score = score
			scoreMap[score] = index
		}
		sort.Slice(scores, func(i, j int) bool {
			return scores[i] > scores[j]
		})

		var result []*api.Index
		if limit > len(scores) {
			limit = len(scores)
		}
		for _, score := range scores[:limit] {
			index := scoreMap[score]
			result = append(result, index.API())
		}
	} else {
		vs, err := h.vs.Search(ctx, groupKey, embeddingData, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to search vector: %w", err)
		}
		ids := make([]uint, len(vs))
		for i, v := range vs {
			ids[i] = v.IndexID
		}
		indexes, err := h.sh.GetIndexes(ctx, ids)
		if err != nil {
			return nil, fmt.Errorf("failed to get indexes: %w", err)
		}
		indexesMap := make(map[uint]*models.Index, len(indexes))
		for _, index := range indexes {
			indexesMap[index.ID] = index
		}
		for _, v := range vs {
			index := indexesMap[v.IndexID]
			index.Score = v.Score
			result = append(result, index.API())
		}
	}

	return result, nil
}

func (h *IndexHandler) UpsertIndexes(ctx context.Context, req api.UpsertIndexesRequest) ([]*api.Index, error) {
	m, err := h.llmsh.GetEmbeddingModel(req.EmbeddingModel)
	if err != nil {
		return nil, err
	}

	indexes := make([]*models.Index, 0, len(req.Items))
	embeddingInput := make([]string, 0, len(req.Items))
	embeddingIndexes := make([]*models.Index, 0, len(req.Items))
	rollbackIndexes := make([]*models.Index, 0, len(req.Items))

	for _, item := range req.Items {
		var index *models.Index
		if item.ID != 0 {
			var err error
			index, err = h.sh.GetIndex(ctx, item.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to get index %d: %w", item.ID, err)
			}
			if index == nil {
				return nil, &IndexNotFoundError{ID: item.ID}
			}
		}

		createEmbedding := req.ForceRebuild
		if index == nil {
			createEmbedding = true
			index = &models.Index{
				GroupKey:   req.GroupKey,
				Data:       item.Data,
				Properties: models.IndexProperties(item.Properties),
			}
		} else {
			rollbackIndex := *index
			rollbackIndexes = append(rollbackIndexes, &rollbackIndex)
			if index.Data != item.Data || (h.vs == nil && len(index.Vector) == 0) {
				createEmbedding = true
			}
			index.Properties = models.IndexProperties(item.Properties)
		}
		indexes = append(indexes, index)
		if createEmbedding {
			embeddingInput = append(embeddingInput, item.Data)
			embeddingIndexes = append(embeddingIndexes, index)
		}
	}

	var vs []*Vector
	if len(embeddingInput) > 0 {
		embeddingRequest := llmsapi.CreateEmbeddingRequest{
			Input: embeddingInput,
		}
		embeddingResponse, err := m.CreateEmbedding(ctx, embeddingRequest)
		if err != nil {
			return nil, fmt.Errorf("failed to create embedding: %w", err)
		}

		for _, item := range embeddingResponse.Data {
			index := embeddingIndexes[item.Index]
			if h.vs == nil {
				index.Vector = item.Embedding
			} else {
				vs = append(vs, &Vector{
					IndexID: index.ID,
					Data:    item.Embedding,
				})
			}
		}
	}

	newIndexesIDs, err := h.sh.UpsertIndexes(ctx, indexes)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert indexes: %w", err)
	}

	if h.vs != nil && len(vs) > 0 {
		if err := h.vs.Upsert(ctx, req.GroupKey, vs); err != nil {
			if len(newIndexesIDs) > 0 {
				if err := h.sh.DeleteIndexes(ctx, newIndexesIDs); err != nil {
					h.logger.With(zap.Error(err), zap.Any("ids", newIndexesIDs)).Error("rollback error: failed to delete indexes")
				}
			}

			if len(rollbackIndexes) > 0 {
				if _, err := h.sh.UpsertIndexes(ctx, rollbackIndexes); err != nil {
					h.logger.With(zap.Error(err)).Error("rollback error: failed to upsert indexes")
				}
			}
		}
	}

	respIndexes := make([]*api.Index, 0, len(indexes))
	for _, index := range indexes {
		respIndexes = append(respIndexes, index.API())
	}

	return respIndexes, nil
}
