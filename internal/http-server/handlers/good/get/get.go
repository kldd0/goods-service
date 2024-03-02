package get

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"log/slog"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/kldd0/goods-service/internal/domain/models"
	http_serv "github.com/kldd0/goods-service/internal/http-server"
	"github.com/kldd0/goods-service/internal/logger"
	"github.com/kldd0/goods-service/internal/storage"
)

type goodGetter interface {
	GetGood(ctx context.Context, key string) (models.Good, error)
}

type cacheModifier interface {
	GetGood(ctx context.Context, key string) (models.Good, error)
	SetGood(ctx context.Context, key string, value models.Good) error
}

func New(log *slog.Logger, db goodGetter, cache cacheModifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "good.get.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		goodId := chi.URLParam(r, "id")

		// check if string id is number
		_, err := strconv.Atoi(goodId)
		if goodId == "" || err != nil {
			log.Info("bad request")
			http_serv.RespondWithErr(err, w, r, "bad request", http.StatusBadRequest)
			return
		}

		// looking for the good in cache
		// TODO: make correct check for cache response (there are multiple cases)
		// a more advanced check is needed here
		requestedGood, err := cache.GetGood(r.Context(), goodId)
		if (err == nil && requestedGood != models.Good{}) {
			log.Info("the good was taken from the cache", slog.String("id", goodId))
			render.JSON(w, r, requestedGood)
			return
		}

		requestedGood, err = db.GetGood(r.Context(), goodId)
		if errors.Is(err, storage.ErrEntryDoesntExist) {
			log.Info("good not found", slog.String("id", goodId))
			http_serv.RespondWithErr(err, w, r, "not found", http.StatusNotFound)
			return
		}

		if err != nil {
			log.Error("failed to get the good", logger.Err(err))
			http_serv.RespondWithErr(err, w, r, "internal error", http.StatusInternalServerError)
			return
		}

		// add the retrieved good to the cache
		key := fmt.Sprintf("%s$%d", goodId, requestedGood.ProjectId)
		err = cache.SetGood(r.Context(), key, requestedGood)
		if err != nil {
			log.Error("failed to set good in cache", logger.Err(err))
		}

		render.JSON(w, r, requestedGood)
	}
}
