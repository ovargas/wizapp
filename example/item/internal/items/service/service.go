package service

import (
	"context"
	itemsV1 "github.com/ovargas/wizapp/example/api/items/v1"
)

type Item struct {
	itemsV1.UnimplementedItemServiceServer
}

var fn func()

func New() *Item {
	return new(Item)
}

func (*Item) Get(ctx context.Context, r *itemsV1.GetRequest) (*itemsV1.Item, error) {
	if r.Id == "panic" {
		fn()
	}

	return &itemsV1.Item{Name: "The item"}, nil
}

func (*Item) Create(context.Context, *itemsV1.CreateRequest) (*itemsV1.Item, error) {
	return &itemsV1.Item{Name: "The item"}, nil
}
