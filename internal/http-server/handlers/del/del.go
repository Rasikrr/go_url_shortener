package del

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go_url_chortener_api/internal/http-server/customJson"
	resp "go_url_chortener_api/internal/lib/api/response"
	"go_url_chortener_api/internal/lib/logger/sl"
	"log/slog"
	"net/http"
)

type URLDeleter interface {
	DeleteURL(string) error
}

func New(log *slog.Logger, deleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.del.New"
		log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := getAlias(r)

		err := deleter.DeleteURL(alias)
		if err != nil {
			log.Error("failed to delete url", sl.Err(err))
			customJson.WriteJson(w, http.StatusInternalServerError, resp.Error("failed to delete url"))
			return
		}
		customJson.WriteJson(w, http.StatusOK, resp.OK())
	}
}

func getAlias(r *http.Request) string {
	alias := chi.URLParam(r, "alias")
	return alias
}
