package post

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"log/slog"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"github.com/kldd0/goods-service/internal/domain/models"
	http_server "github.com/kldd0/goods-service/internal/http-server"
	"github.com/kldd0/goods-service/internal/logger"
	"github.com/kldd0/goods-service/internal/storage"
)

type Request struct {
	Payload models.Good
}

type goodSaver interface {
	SaveGood(ctx context.Context, good models.Good) (models.Good, error)
}

func New(log *slog.Logger, db goodSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "good.post.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			// body of request is empty
			log.Error("request body is empty")
			render.JSON(w, r, http_server.Error("empty request"))
			return
		}

		if err != nil {
			log.Error("failed to decode request body", logger.Err(err))
			render.JSON(w, r, http_server.Error("failed to decode request"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Error("invalid request", logger.Err(err))
			render.JSON(w, r, http_server.ValidationError(validateErr))
			return
		}

		good, err := db.SaveGood(r.Context(), req.Payload)
		if errors.Is(err, storage.ErrEntryAlreadyExists) {
			log.Info("good already exists")
			RespondWithErr(err, w, r, "good already exists", http.StatusConflict)
			return
		}

		if err != nil {
			log.Error("failed to add good", logger.Err(err))
			RespondWithErr(err, w, r, "failed to add good", http.StatusInternalServerError)
			return
		}

		log.Info("good added", slog.Int64("id", int64(good.ID)))

		render.JSON(w, r, good)
	}
}

func RespondWithErr(err error, w http.ResponseWriter, r *http.Request, msg string, status int) {
	resp := http_server.Error(msg)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}
