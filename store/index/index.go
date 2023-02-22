package index

import (
	"context"
	"fmt"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/internal/milvus"
)

func New(ctx context.Context, client *milvus.Client) core.IndexStore {
	return &storeImpl{
		client: client,
	}
}

type storeImpl struct {
	client *milvus.Client
}

func (s *storeImpl) QueryIndex(ctx context.Context, idx *core.Index) error {
	return nil
}

func (s *storeImpl) DeleteByPks(ctx context.Context, items []*core.Index) error {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		id := fmt.Sprintf("%s:%s", item.AppID, item.ObjectID)
		ids = append(ids, id)
	}

	index := core.Index{}
	pks := entity.NewColumnVarChar("id", ids)
	if err := s.client.DeleteByPks(ctx, index.CollectionName(), "", pks); err != nil {
		return err
	}

	return nil
}

func (s *storeImpl) CreateIndices(ctx context.Context, idx []*core.Index) error {
	l := len(idx)
	ids := make([]string, 0, l)
	appIds := make([]string, 0, l)
	datas := make([]string, 0, l)
	vectors := make([][]float32, 0, l)
	objectIds := make([]string, 0, l)
	categories := make([]string, 0, l)
	properties := make([]string, 0, l)
	createdAts := make([]int64, 0, l)
	for _, ix := range idx {
		ix.ID = fmt.Sprintf("%s:%s", ix.AppID, ix.ObjectID)
		ids = append(ids, ix.ID)
		appIds = append(appIds, ix.AppID)
		datas = append(datas, ix.Data)
		vectors = append(vectors, ix.Vectors)
		objectIds = append(objectIds, ix.ObjectID)
		categories = append(categories, ix.Category)
		properties = append(properties, ix.Properties)
		createdAts = append(createdAts, ix.CreatedAt)
	}

	ix := &core.Index{}
	_, err := s.client.Insert(
		ctx,
		ix.CollectionName(),
		"",
		entity.NewColumnVarChar("id", ids),
		entity.NewColumnVarChar("app_id", appIds),
		entity.NewColumnVarChar("data", datas),
		entity.NewColumnFloatVector("vectors", 1536, vectors),
		entity.NewColumnVarChar("object_id", objectIds),
		entity.NewColumnVarChar("category", categories),
		entity.NewColumnVarChar("properties", properties),
		entity.NewColumnInt64("created_at", createdAts),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *storeImpl) Search(ctx context.Context, vectors []float32, n int) ([]*core.Index, error) {
	idx := &core.Index{}
	sp, _ := entity.NewIndexFlatSearchParam()
	searchResult, err := s.client.Search(
		ctx,
		idx.CollectionName(),
		[]string{},
		"",
		[]string{"data"},
		[]entity.Vector{
			entity.FloatVector(vectors),
		},
		"vectors",
		entity.L2,
		n,
		sp,
	)
	if err != nil {
		return nil, err
	}

	indices := []*core.Index{}

	// TODO: parse searchResult to indices
	_ = searchResult

	// index := &core.Index{}
	// sr := searchResult[0]
	// for _, f := range sr.Fields {
	// 	switch f.Name() {
	// 	case "id":
	// 	}
	// }
	//
	return indices, nil
}
