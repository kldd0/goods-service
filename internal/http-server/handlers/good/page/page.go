package page

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"log/slog"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/kldd0/goods-service/internal/domain/models"
	http_serv "github.com/kldd0/goods-service/internal/http-server"
	"github.com/kldd0/goods-service/internal/logger"
	"github.com/kldd0/goods-service/internal/storage"
)

type Meta struct {
	Total   int `json:"total"`
	Removed int `json:"removed"`
	Limit   int `json:"limit"`
	Offset  int `json:"offset"`
}

type Response struct {
	Meta `json:"meta"`

	Goods []models.Good `json:"goods"`
}

type goodsGetter interface {
	ListGoodsWithPagination(ctx context.Context, offset, limit string) ([]models.Good, error)
}

type cacheModifier interface {
	SetGood(ctx context.Context, key string, value models.Good) error
}

func New(log *slog.Logger, db goodsGetter, cache cacheModifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "good.get.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		limit := r.URL.Query().Get("limit")

		// check if string id is number
		limitNum, err := strconv.Atoi(limit)
		if limit == "" || err != nil {
			log.Info("bad request", slog.Any("limit", limit))
			http_serv.RespondWithErr(err, w, r, "bad request", http.StatusBadRequest)
			return
		}

		offset := r.URL.Query().Get("offset")

		// check if string id is number
		offsetNum, err := strconv.Atoi(offset)
		if offset == "" || err != nil {
			log.Info("bad request", slog.Any("offset", offset))
			http_serv.RespondWithErr(err, w, r, "bad request", http.StatusBadRequest)
			return
		}

		requestedGoods, err := db.ListGoodsWithPagination(r.Context(), limit, offset)
		if errors.Is(err, storage.ErrEntryDoesntExist) {
			log.Info("goods not found", slog.Any("limit", limit), slog.Any("offset", offset))
			render.JSON(w, r, requestedGoods)
			return
		}

		if err != nil {
			log.Error("failed to get the goods", logger.Err(err))
			http_serv.RespondWithErr(err, w, r, "internal error", http.StatusInternalServerError)
			return
		}

		log.Info("req", slog.Any("goods", requestedGoods))

		removedCount := 0
		// add the retrieved goods to the cache
		for _, good := range requestedGoods {
			key := fmt.Sprintf("%d$%d", good.ID, good.ProjectId)
			err = cache.SetGood(r.Context(), key, good)
			if err != nil {
				log.Error("failed to set good in cache", logger.Err(err))
			}

			if good.Removed {
				removedCount++
			}
		}

		// creating response
		resp := Response{
			Meta{len(requestedGoods), removedCount, limitNum, offsetNum},
			requestedGoods,
		}

		render.JSON(w, r, resp)
	}
}
