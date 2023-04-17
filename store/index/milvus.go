package index

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/pandodao/botastic/config"
	"github.com/pandodao/botastic/core"
)

type milvusStore struct {
	cfg    *config.IndexStoreMilvusConfig
	client client.Client
}

func initMilvusStore(ctx context.Context, cfg *config.IndexStoreMilvusConfig) (*milvusStore, error) {
	milvusClient, err := client.NewGrpcClient(ctx, cfg.Address)
	if err != nil {
		return nil, err
	}

	return &milvusStore{
		cfg:    cfg,
		client: milvusClient,
	}, nil
}

func (s *milvusStore) Init(ctx context.Context) error {
	collExists, err := s.client.HasCollection(ctx, s.cfg.Collection)
	if err != nil {
		return err
	}

	if !collExists {
		err = s.createCollectionIfNotExist(ctx)
		if err != nil {
			return err
		}

		idx, err := entity.NewIndexIvfFlat(entity.L2, 1024)
		if err != nil {
			return err
		}

		if err := s.client.CreateIndex(ctx, s.cfg.Collection, "vectors", idx, false); err != nil {
			return err
		}
	}

	// load collection
	if err = s.client.LoadCollection(ctx, s.cfg.Collection, false); err != nil {
		return err
	}

	return nil
}

func (s *milvusStore) Delete(ctx context.Context, appID string, items []*core.Index) error {
	if len(items) == 0 {
		return nil
	}

	partitionName := s.getPartitionName(appID)
	ids := make([]string, 0, len(items))
	for _, item := range items {
		if item.AppID != appID {
			return fmt.Errorf("app id not match: %s != %s", item.AppID, appID)
		}

		ids = append(ids, s.getIndexID(item))
	}

	pks := entity.NewColumnVarChar("id", ids)
	if err := s.client.DeleteByPks(ctx, s.cfg.Collection, partitionName, pks); err != nil {
		return err
	}

	return nil
}

func (s *milvusStore) Upsert(ctx context.Context, appId string, idx []*core.Index) error {
	if len(idx) == 0 {
		return nil
	}

	l := len(idx)
	ids := make([]string, 0, l)
	appIds := make([]string, 0, l)
	datas := make([]string, 0, l)
	dataTokens := make([]int64, 0, l)
	vectors := make([][]float32, 0, l)
	objectIds := make([]string, 0, l)
	categories := make([]string, 0, l)
	properties := make([]string, 0, l)
	createdAts := make([]int64, 0, l)
	createdAt := time.Now().Unix()

	partitionName := s.getPartitionName(appId)

	if err := s.createPartionIfNotExist(ctx, partitionName); err != nil {
		return err
	}

	for _, ix := range idx {
		if ix.AppID != appId {
			return fmt.Errorf("app id not match: %s != %s", ix.AppID, appId)
		}

		ix.CreatedAt = createdAt
		ids = append(ids, s.getIndexID(ix))
		appIds = append(appIds, ix.AppID)
		datas = append(datas, ix.Data)
		dataTokens = append(dataTokens, ix.DataToken)
		vectors = append(vectors, ix.Vectors)
		objectIds = append(objectIds, ix.ObjectID)
		categories = append(categories, ix.Category)
		properties = append(properties, ix.Properties)
		createdAts = append(createdAts, ix.CreatedAt)
	}

	if err := s.Delete(ctx, appId, idx); err != nil {
		return err
	}

	_, err := s.client.Insert(
		ctx,
		s.cfg.Collection,
		partitionName,
		entity.NewColumnVarChar("id", ids),
		entity.NewColumnVarChar("app_id", appIds),
		entity.NewColumnVarChar("data", datas),
		entity.NewColumnInt64("data_token", dataTokens),
		entity.NewColumnFloatVector("vectors", 1536, vectors),
		entity.NewColumnVarChar("object_id", objectIds),
		entity.NewColumnVarChar("category", categories),
		entity.NewColumnVarChar("properties", properties),
		entity.NewColumnInt64("created_at", createdAts),
	)
	return err
}

func (s *milvusStore) Reset(ctx context.Context, appID string) error {
	partitionName := s.getPartitionName(appID)
	exists, err := s.client.HasPartition(ctx, s.cfg.Collection, partitionName)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	// query data by app_id
	es, err := s.client.Query(ctx, s.cfg.Collection, []string{partitionName}, fmt.Sprintf(`app_id == "%s"`, appID), []string{"id"})
	if err != nil {
		return err
	}

	e := es[0]
	data := e.FieldData().GetScalars().GetStringData().Data
	if len(data) == 0 {
		return nil
	}

	return s.client.DeleteByPks(ctx, s.cfg.Collection, partitionName, e)
}

func (s *milvusStore) Search(ctx context.Context, appID string, vectors []float32, n int) ([]*core.Index, error) {
	partitionName := s.getPartitionName(appID)
	partitionExist, err := s.client.HasPartition(ctx, s.cfg.Collection, partitionName)
	if err != nil {
		return nil, err
	}
	if !partitionExist {
		return []*core.Index{}, nil
	}

	sp, _ := entity.NewIndexIvfFlatSearchParam(10)
	searchResult, err := s.client.Search(
		ctx,
		s.cfg.Collection,
		[]string{partitionName},
		"",
		[]string{"data", "object_id", "properties", "category", "created_at"},
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
	if len(searchResult) == 0 {
		return []*core.Index{}, nil
	}

	indexes := []*core.Index{}
	sr := searchResult[0]

	var (
		dataCol       *entity.ColumnVarChar
		objectIdCol   *entity.ColumnVarChar
		propertiesCol *entity.ColumnVarChar
		categoryCol   *entity.ColumnVarChar
		createdAtCol  *entity.ColumnInt64
	)
	for _, f := range sr.Fields {
		switch f.Name() {
		case "data":
			dataCol = f.(*entity.ColumnVarChar)
		case "object_id":
			objectIdCol = f.(*entity.ColumnVarChar)
		case "properties":
			propertiesCol = f.(*entity.ColumnVarChar)
		case "category":
			categoryCol = f.(*entity.ColumnVarChar)
		case "created_at":
			createdAtCol = f.(*entity.ColumnInt64)
		}
	}

	for i := 0; i < sr.ResultCount; i++ {
		index := &core.Index{}
		index.Data, _ = dataCol.ValueByIdx(i)
		index.ObjectID, _ = objectIdCol.ValueByIdx(i)
		index.Properties, _ = propertiesCol.ValueByIdx(i)
		index.Category, _ = categoryCol.ValueByIdx(i)
		index.CreatedAt, _ = createdAtCol.ValueByIdx(i)
		index.Score = float64(sr.Scores[i])
		indexes = append(indexes, index)
	}

	return indexes, nil
}

func (s *milvusStore) getSchema() *entity.Schema {
	return &entity.Schema{
		CollectionName: s.cfg.Collection,
		AutoID:         true,
		Fields: []*entity.Field{
			{
				Name:       "id",
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: true,
				TypeParams: map[string]string{"max_length": "64"},
			},
			{
				Name:       "app_id",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "32"},
			},
			{
				Name:       "object_id",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "32"},
			},
			{
				Name:       "data",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "2048"},
			},
			{
				Name:     "data_token",
				DataType: entity.FieldTypeInt64,
			},
			{
				Name:       "vectors",
				DataType:   entity.FieldTypeFloatVector,
				TypeParams: map[string]string{"dim": "1536"},
			},
			{
				Name:       "category",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "32"},
			},
			{
				Name:       "properties",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "1024"},
			},
			{
				Name:     "created_at",
				DataType: entity.FieldTypeInt64,
			},
		},
	}
}

func (s *milvusStore) getPartitionName(appID string) string {
	return strings.ReplaceAll(appID, "-", "_")
}

func (s *milvusStore) createCollectionIfNotExist(ctx context.Context) error {
	collExists, err := s.client.HasCollection(ctx, s.cfg.Collection)
	if err != nil {
		return err
	}

	if collExists {
		return nil
	}

	return s.client.CreateCollection(ctx, s.getSchema(), s.cfg.CollectionShardsNum)
}

func (s *milvusStore) createPartionIfNotExist(ctx context.Context, partionName string) error {
	exists, err := s.client.HasPartition(ctx, s.cfg.Collection, partionName)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	return s.client.CreatePartition(ctx, s.cfg.Collection, partionName)
}

func (s *milvusStore) getIndexID(ix *core.Index) string {
	return fmt.Sprintf("%s:%s", ix.AppID, ix.ObjectID)
}
