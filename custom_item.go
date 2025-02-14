package main

import (
	"image"

	"github.com/df-mc/dragonfly/server/item/category"
)

// todo: remove after debugging

type testItem struct {
}

func (i testItem) EncodeItem() (name string, meta int16) {
	return "multiversion:test", 0
}

func (i testItem) Name() string {
	return "testme!"
}

func (i testItem) Texture() image.Image {
	return image.NewRGBA(image.Rect(0, 0, 32, 32))
}

func (i testItem) Category() category.Category {
	return category.Items()
}
