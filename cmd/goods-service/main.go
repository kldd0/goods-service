package main

import (
	"context"

	"github.com/kldd0/goods-service/internal/config"
)

func main() {
	_ = config.MustLoad()

	_, cancel := context.WithCancel(context.Background())
	defer cancel()
}
