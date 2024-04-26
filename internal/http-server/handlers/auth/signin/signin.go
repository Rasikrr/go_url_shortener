package signin

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"go_url_chortener_api/internal/domain"
	"go_url_chortener_api/internal/http-server/customJson"
	"go_url_chortener_api/internal/http-server/handlers/refresh"
	"go_url_chortener_api/internal/http-server/middleware/myJwt"
	resp "go_url_chortener_api/internal/lib/api/response"
	"go_url_chortener_api/internal/lib/hash"
	"go_url_chortener_api/internal/lib/logger/sl"
	"log/slog"
	"net/http"
)

type Request struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type SignInner interface {
	GetUser(email string) (*domain.User, error)
	SaveRefresh(string, int) error
	DeleteRefreshByUserId(int) error
}

func New(log *slog.Logger, signInner SignInner, hasher hash.PasswordHasher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.auth.signin.New"
		log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		req := new(Request)

		if err := customJson.DecodeJson(r, req); err != nil {
			log.Error("failed to decode json", sl.Err(err))
			customJson.WriteJson(w, http.StatusBadRequest, resp.Error("bad request"))
		}

		if err := validator.New().Struct(req); err != nil {
			validationErr := err.(validator.ValidationErrors)
			log.Error("invalid request", sl.Err(err))
			customJson.WriteJson(w, http.StatusBadRequest, resp.ValidationError(validationErr))
			return
		}

		user, err := signInner.GetUser(req.Email)
		if err != nil {
			log.Error("failed to find user", sl.Err(err))
			customJson.WriteJson(w, http.StatusBadRequest, resp.Error("user with this email not found"))
			return
		}
		if err := hasher.CheckPassword(user.EncPassword, req.Password); err != nil {
			log.Error("failed to match password", sl.Err(err))
			customJson.WriteJson(w, http.StatusBadRequest, resp.Error("invalid password"))
			return
		}

		jwtToken, err := myJwt.CreateJWT(user.Id, user.Email)

		if err != nil {
			log.Error("failed to create JWT token", sl.Err(err))
			customJson.WriteJson(w, http.StatusInternalServerError, resp.Error("server error"))
			return
		}

		refreshToken, err := refresh.CreateRefresh(user.Id)
		if err != nil {
			log.Error("failed to create refresh token", sl.Err(err))
			customJson.WriteJson(w, http.StatusInternalServerError, resp.Error("server error"))
			return
		}

		if err := signInner.DeleteRefreshByUserId(user.Id); err != nil {
			log.Error("")
			customJson.WriteJson(w, http.StatusInternalServerError, resp.Error("server error"))
			return
		}
		log.Info("old refresh token was deleted")

		if err := signInner.SaveRefresh(refreshToken, user.Id); err != nil {
			log.Error("failed to create refresh token", sl.Err(err))
			customJson.WriteJson(w, http.StatusInternalServerError, resp.Error("server error"))
			return
		}

		myJwt.SetJWTHeader(w, jwtToken)
		myJwt.SetIdCookie(w, user.Id)

		refresh.SetRefreshCookie(w, refreshToken)

		customJson.WriteJson(w, http.StatusOK, map[string]string{
			"jwt": jwtToken,
		})
	}
}
