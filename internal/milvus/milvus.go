package milvus

import (
	"context"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

type Client struct {
	client.Client
}

func Init(ctx context.Context, addr string) (*Client, error) {
	milvusClient, err := client.NewGrpcClient(ctx, addr)
	if err != nil {
		return nil, err
	}

	return &Client{
		Client: milvusClient,
	}, nil
}

func (c *Client) CreateCollectionIfNotExist(ctx context.Context, s *entity.Schema, shardsNum int32, opts ...client.CreateCollectionOption) error {
	collExists, err := c.HasCollection(ctx, s.CollectionName)
	if err != nil {
		return err
	}

	if collExists {
		return nil
	}

	err = c.CreateCollection(ctx, s, shardsNum)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) CreatePartionIfNotExist(ctx context.Context, collName, partionName string) error {
	collExists, err := c.HasPartition(ctx, collName, partionName)
	if err != nil {
		return err
	}

	if collExists {
		return nil
	}

	err = c.CreatePartition(ctx, collName, partionName)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) BuildIndex(ctx context.Context, collName, fieldName string) error {
	idx, err := entity.NewIndexIvfFlat(
		entity.L2,
		1024,
	)
	if err != nil {
		return err
	}

	err = c.CreateIndex(
		ctx,       // ctx
		collName,  // CollectionName
		fieldName, // fieldName
		idx,       // entity.Index
		false,     // async
	)
	if err != nil {
		return err
	}
	return nil
}
