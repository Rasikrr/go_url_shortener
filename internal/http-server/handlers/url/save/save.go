package save

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"go_url_chortener_api/internal/http-server/customJson"
	resp "go_url_chortener_api/internal/lib/api/response"
	"go_url_chortener_api/internal/lib/logger/sl"
	"go_url_chortener_api/internal/lib/random"
	"go_url_chortener_api/internal/storage"
	"log/slog"
	"net/http"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

// TODO: move to config

const aliasLength = 6

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

type URLSaver interface {
	SaveURL(urlToSave string, alias string) error
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.url.save.New"
		log = log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := customJson.DecodeJson(r, &req)

		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			customJson.WriteJson(w, http.StatusBadRequest, resp.Error("failed to decode request"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validationErr := err.(validator.ValidationErrors)
			log.Error("invalid request", sl.Err(err))
			customJson.WriteJson(w, http.StatusBadRequest, resp.ValidationError(validationErr))
			return
		}
		alias := req.Alias
		// TODO handle alias exists error
		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}
		err = urlSaver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Error("url already exists", slog.String("url", req.URL))
			customJson.WriteJson(w, http.StatusBadRequest, resp.Error("url already exists"))
			return
		}
		if err != nil {
			log.Error("failed to add url", sl.Err(err))
			customJson.WriteJson(w, http.StatusInternalServerError, resp.Error("failed to add error"))
			return
		}
		log.Info("url added")

		customJson.WriteJson(w, http.StatusOK, resp.OK())
	}
}
