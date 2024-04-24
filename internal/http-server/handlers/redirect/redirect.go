package redirect

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go_url_chortener_api/internal/http-server/customJson"
	resp "go_url_chortener_api/internal/lib/api/response"
	"log/slog"
	"net/http"
)

type URLGetter interface {
	GetURL(string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.redirect.New"
		log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := getAlias(r)

		urlToResp, err := urlGetter.GetURL(alias)
		if err != nil {
			log.Error("failed getting url", slog.String("alias", alias))
			customJson.WriteJson(w, http.StatusInternalServerError, resp.Error("failed getting url"))
			return
		}
		log.Info("url sent")
		log.Info("redirecting...")
		http.Redirect(w, r, urlToResp, http.StatusFound)
	}
}

func getAlias(r *http.Request) string {
	alias := chi.URLParam(r, "alias")
	return alias
}
