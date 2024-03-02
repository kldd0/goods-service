package storage

import (
	"context"
	"fmt"

	"github.com/kldd0/goods-service/internal/domain/models"
)

type Storage interface {
	GetGood(ctx context.Context, goodId string) (models.Good, error)
	SaveGood(ctx context.Context, good models.Good) (models.Good, error)
	PatchGood(ctx context.Context, patchedGood models.Good) (models.Good, error)
}

var (
	ErrEntryAlreadyExists  = fmt.Errorf("entry already exists")
	ErrEntryDoesntExist    = fmt.Errorf("entry doesn't exist")
	ErrGettingInsertedRows = fmt.Errorf("failed getting inserted row")
)
