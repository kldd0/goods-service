package delete

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"log/slog"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	http_serv "github.com/kldd0/goods-service/internal/http-server"
	"github.com/kldd0/goods-service/internal/logger"
	"github.com/kldd0/goods-service/internal/storage"
)

type Response struct {
	ID        int  `json:"id"`
	ProjectId int  `json:"project_id"`
	Removed   bool `json:"removed"`
}

type goodDeleter interface {
	DeleteGood(ctx context.Context, goodId string, projectId string) error
}

type cacheInvalidator interface {
	Delete(ctx context.Context, goodId string) error
}

func New(log *slog.Logger, db goodDeleter, cache cacheInvalidator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "good.post.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		goodId := r.URL.Query().Get("id")

		// check if string id is number
		goodIdNum, err := strconv.Atoi(goodId)
		if goodId == "" || err != nil {
			log.Info("bad request", slog.Any("goodId", goodId))
			http_serv.RespondWithErr(err, w, r, "bad request", http.StatusBadRequest)
			return
		}

		projectId := r.URL.Query().Get("projectId")

		// check if string id is number
		projectIdNum, err := strconv.Atoi(projectId)
		if projectId == "" || err != nil {
			log.Info("bad request", slog.Any("projectId", projectId))
			http_serv.RespondWithErr(err, w, r, "bad request", http.StatusBadRequest)
			return
		}

		err = db.DeleteGood(r.Context(), goodId, projectId)
		if errors.Is(err, storage.ErrEntryDoesntExist) {
			log.Info("good doesn't exist")
			http_serv.RespondWithErr(err, w, r, "good doesn't exist", http.StatusNotFound)
			return
		}

		if err != nil {
			log.Error("failed to delete good", logger.Err(err))
			http_serv.RespondWithErr(err, w, r, "failed to delete good", http.StatusInternalServerError)
			return
		}

		resp := Response{goodIdNum, projectIdNum, true}

		log.Info("good removed", slog.Int64("id", int64(goodIdNum)))

		// invalidate cache
		err = cache.Delete(r.Context(), fmt.Sprintf("%s$%s", goodId, projectId))
		if err != nil {
			log.Error("failed to delete good from cache", logger.Err(err))
		}

		render.JSON(w, r, resp)
	}
}
