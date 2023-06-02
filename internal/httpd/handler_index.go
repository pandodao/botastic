package httpd

import (
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/pandodao/botastic/api"
	"github.com/pandodao/botastic/internal/utils"
	"github.com/pandodao/botastic/internal/vector"
	"github.com/pandodao/botastic/models"
	llmsapi "github.com/pandodao/botastic/pkg/llms/api"
)

func (h *Handler) UpsertIndexes(c *gin.Context) {
	var req api.UpsertIndexesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respErr(c, http.StatusBadRequest, err)
		return
	}

	m, ok := h.llms.GetEmbeddingModel(req.EmbeddingModel)
	if !ok {
		h.respErr(c, http.StatusBadRequest, errors.New("embedding model does not exist"))
		return
	}

	indexes := make([]*models.Index, 0, len(req.Items))
	embeddingInput := make([]string, 0, len(req.Items))
	embeddingIndexes := make([]*models.Index, 0, len(req.Items))

	for _, item := range req.Items {
		var index *models.Index
		if item.ID != 0 {
			var err error
			index, err = h.sh.GetIndex(c, item.ID)
			if err != nil {
				h.respErr(c, http.StatusInternalServerError, err)
				return
			}
			if index == nil {
				h.respErr(c, http.StatusBadRequest, fmt.Errorf("index not found: %d", item.ID))
				return
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
			if index.Data != item.Data || (h.vectorStorage == nil && len(index.Vector) == 0) {
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

	var vs []*vector.Vector
	if len(embeddingInput) > 0 {
		embeddingRequest := llmsapi.CreateEmbeddingRequest{
			Input: embeddingInput,
		}
		tokens, err := m.CalEmbeddingRequestTokens(embeddingRequest)
		if err != nil {
			h.respErr(c, http.StatusInternalServerError, fmt.Errorf("failed to calculate embedding tokens: %w", err))
			return
		}
		if m.MaxRequestTokens() > 0 && tokens > m.MaxRequestTokens() {
			h.respErr(c, http.StatusBadRequest, fmt.Errorf("too many tokens: %d, max: %d", tokens, m.MaxRequestTokens()))
			return
		}

		embeddingResponse, err := m.CreateEmbedding(c, embeddingRequest)
		if err != nil {
			h.respErr(c, http.StatusInternalServerError, fmt.Errorf("failed to create embedding: %w", err))
			return
		}

		for _, item := range embeddingResponse.Data {
			index := embeddingIndexes[item.Index]
			if h.vectorStorage == nil {
				index.Vector = item.Embedding
			} else {
				vs = append(vs, &vector.Vector{
					IndexID: index.ID,
					Data:    item.Embedding,
				})
			}
		}
	}

	if err := h.sh.UpsertIndexes(c, indexes, func() error {
		if h.vectorStorage != nil && len(vs) > 0 {
			return h.vectorStorage.Upsert(c, req.GroupKey, vs)
		}
		return nil
	}); err != nil {
		h.respErr(c, http.StatusInternalServerError, fmt.Errorf("failed to upsert indexes: %w", err))
		return
	}

	respIndexes := make([]*api.Index, 0, len(indexes))
	for _, index := range indexes {
		respIndexes = append(respIndexes, index.API())
	}

	h.respData(c, api.UpsertIndexesResponse{
		Items: respIndexes,
	})
}

func (h *Handler) SearchIndexes(c *gin.Context) {
	var req api.SearchIndexesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respErr(c, http.StatusBadRequest, err)
		return
	}

	m, ok := h.llms.GetEmbeddingModel(req.EmbeddingModel)
	if !ok {
		h.respErr(c, http.StatusBadRequest, errors.New("embedding model does not exist"))
		return
	}

	embeddingReq := llmsapi.CreateEmbeddingRequest{
		Input: []string{req.Keyword},
	}
	tokens, err := m.CalEmbeddingRequestTokens(embeddingReq)
	if err != nil {
		h.respErr(c, http.StatusInternalServerError, fmt.Errorf("failed to calculate embedding tokens: %w", err))
		return
	}
	if m.MaxRequestTokens() > 0 && tokens > m.MaxRequestTokens() {
		h.respErr(c, http.StatusBadRequest, fmt.Errorf("too many tokens: %d, max: %d", tokens, m.MaxRequestTokens()))
		return
	}

	embeddingResp, err := m.CreateEmbedding(c, embeddingReq)
	if err != nil {
		h.respErr(c, http.StatusInternalServerError, fmt.Errorf("failed to create embedding: %w", err))
		return
	}

	result := &api.SearchIndexesResponse{
		Items: make([]*api.Index, 0, req.Limit),
	}
	if h.vectorStorage == nil {
		indexes, err := h.sh.GetIndexesByGroupKey(c, req.GroupKey)
		if err != nil {
			h.respErr(c, http.StatusInternalServerError, fmt.Errorf("failed to get indexes: %w", err))
			return
		}
		indexesMap := make(map[uint]*models.Index, len(indexes))
		scoreMap := make(map[int64]uint, len(indexes))
		scores := make([]float64, 0, len(indexes))
		for _, index := range indexes {
			score := utils.CosineSimilarity(embeddingResp.Data[0].Embedding, index.Vector)
			scores = append(scores, score)
			scoreMap[int64(index.ID)] = index.ID
			indexesMap[index.ID] = index
		}
		sort.Slice(scores, func(i, j int) bool {
			return scores[i] > scores[j]
		})

		for i, score := range scores[:req.Limit] {
			index := indexesMap[scoreMap[int64(i)]]
			apiIndex := index.API()
			apiIndex.Score = score
			result.Items = append(result.Items, apiIndex)
		}
	} else {
		vs, err := h.vectorStorage.Search(c, req.GroupKey, embeddingResp.Data[0].Embedding, req.Limit)
		if err != nil {
			h.respErr(c, http.StatusInternalServerError, fmt.Errorf("failed to search indexes: %w", err))
			return
		}
		ids := make([]uint, len(vs))
		for i, v := range vs {
			ids[i] = v.IndexID
		}
		indexes, err := h.sh.GetIndexes(c, ids)
		if err != nil {
			h.respErr(c, http.StatusInternalServerError, fmt.Errorf("failed to get indexes: %w", err))
			return
		}
		indexesMap := make(map[uint]*models.Index, len(indexes))
		for _, index := range indexes {
			indexesMap[index.ID] = index
		}
		for _, v := range vs {
			index := indexesMap[v.IndexID]
			apiIndex := index.API()
			apiIndex.Score = v.Score
			result.Items = append(result.Items, apiIndex)
		}
	}

	h.respData(c, result)
}
