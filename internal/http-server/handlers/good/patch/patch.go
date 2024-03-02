package patch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"log/slog"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"github.com/kldd0/goods-service/internal/domain/models"
	http_serv "github.com/kldd0/goods-service/internal/http-server"
	"github.com/kldd0/goods-service/internal/logger"
	"github.com/kldd0/goods-service/internal/storage"
)

type GoodPatched struct {
	ID          int       `json:"id"`
	ProjectId   int       `json:"project_id"`
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description" validate:"required"`
	Priority    *int      `json:"priority"`
	Removed     bool      `json:"removed" validate:"required"`
	CreatedAt   time.Time `json:"created_at"`
}

type Request struct {
	Payload GoodPatched `json:"Payload"`
}

type goodPatcher interface {
	PatchGood(ctx context.Context, patchedGood models.Good) (models.Good, error)
}

type cacheInvalidator interface {
	Delete(ctx context.Context, goodId string) error
}

func New(log *slog.Logger, db goodPatcher, cache cacheInvalidator) http.HandlerFunc {
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
			log.Info("bad request")
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

		var req Request

		err = render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			// body of request is empty
			log.Error("request body is empty")
			render.JSON(w, r, http_serv.Error("empty request"))
			return
		}

		if err != nil {
			log.Error("failed to decode request body", logger.Err(err))
			render.JSON(w, r, http_serv.Error("failed to decode request"))
			return
		}

		req.Payload.ID = goodIdNum
		req.Payload.ProjectId = projectIdNum
		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req.Payload); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Error("invalid request", logger.Err(err))
			render.JSON(w, r, http_serv.ValidationError(validateErr))
			return
		}

		// potentially (possible) UB type conversion
		// but it is necessary to check other group of required fields
		// TODO: redo
		good, err := db.PatchGood(r.Context(), models.Good(req.Payload))
		if errors.Is(err, storage.ErrEntryDoesntExist) {
			log.Info("good doesn't exist")
			http_serv.RespondWithErr(err, w, r, "good doesn't exist", http.StatusNotFound)
			return
		}

		if errors.Is(err, storage.ErrGettingInsertedRows) {
			log.Info("failed getting inserted row", logger.Err(err))
		}

		if err != nil && !errors.Is(err, storage.ErrGettingInsertedRows) {
			log.Error("failed to patch good", logger.Err(err))
			http_serv.RespondWithErr(err, w, r, "failed to add good", http.StatusInternalServerError)
			return
		}

		log.Info("good added", slog.Int64("id", int64(good.ID)))

		// invalidate cache
		err = cache.Delete(r.Context(), fmt.Sprintf("%s$%s", goodId, projectId))
		if err != nil {
			log.Error("failed to delete good from cache", logger.Err(err))
		}

		render.JSON(w, r, good)
	}
}
