package refresh

import (
	"github.com/golang-jwt/jwt/v5"
	"go_url_chortener_api/internal/domain"
	"go_url_chortener_api/internal/http-server/customJson"
	"go_url_chortener_api/internal/http-server/middleware/myJwt"
	resp "go_url_chortener_api/internal/lib/api/response"
	"go_url_chortener_api/internal/lib/logger/sl"
	"log/slog"
	"net/http"
)

type Token struct {
	Id     int    `json:"id"`
	Token  string `json:"token"`
	UserId int    `json:"userId"`
}

type Refresher interface {
	GetUserById(int) (*domain.User, error)
	GetRefresh(userId int) (*Token, error)
	DeleteRefresh(token string) error
	UpdateRefresh(id int, token string) error
}

func permissionDenied(w http.ResponseWriter) {
	customJson.WriteJson(w, http.StatusForbidden, resp.Error("permission denied"))
}

func New(log *slog.Logger, refresher Refresher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.refresh.New"
		refToken, err := getRefreshCookie(r)
		if err != nil {
			log.Error("failed to get refresh cookie", sl.Err(err))
			customJson.WriteJson(w, http.StatusForbidden, resp.Error("invalid request"))
			return
		}
		token, err := validateRefresh(refToken)
		if err != nil {
			if err := refresher.DeleteRefresh(refToken); err != nil {
				log.Error("failed to delete refresh token", sl.Err(err))
			}
			log.Error("failed to validate refresh token", sl.Err(err))
			customJson.WriteJson(w, http.StatusForbidden, resp.Error("invalid request"))
			return
		}
		claims := token.Claims.(jwt.MapClaims)

		userId := int(claims["userId"].(float64))

		user, err := refresher.GetUserById(userId)
		if err != nil {
			log.Error("failed to get user", sl.Err(err))
			customJson.WriteJson(w, http.StatusInternalServerError, resp.Error("server error"))
			return
		}
		tokenFromStorage, err := refresher.GetRefresh(user.Id)
		if err != nil {
			log.Error("failed to get refresh token", sl.Err(err))
			customJson.WriteJson(w, http.StatusInternalServerError, resp.Error("server error"))
			return
		}
		if tokenFromStorage.Token != refToken {
			log.Info("tokens do not match",
				slog.String("ref1", refToken),
				slog.String("ref2", tokenFromStorage.Token),
			)
		}

		if tokenFromStorage.UserId != userId {
			permissionDenied(w)
			return
		}

		// Creating new refresh token
		newRefresh, err := CreateRefresh(userId)
		if err != nil {
			log.Error("failed to create new refresh token", sl.Err(err))
			customJson.WriteJson(w, http.StatusInternalServerError, resp.Error("server error"))
			return
		}

		err = refresher.UpdateRefresh(tokenFromStorage.Id, newRefresh)
		if err != nil {
			log.Error("failed to create new refresh token", sl.Err(err))
			customJson.WriteJson(w, http.StatusInternalServerError, resp.Error("server error"))
			return
		}

		// Creating new JWT
		newJWT, err := myJwt.CreateJWT(user.Id, user.Email)
		if err != nil {
			log.Error("failed to create new refresh token", sl.Err(err))
			customJson.WriteJson(w, http.StatusInternalServerError, resp.Error("server error"))
			return
		}

		myJwt.SetJWTHeader(w, newJWT)
		SetRefreshCookie(w, newRefresh)

		log.Info("new tokens were set")
		customJson.WriteJson(w, http.StatusOK, map[string]string{
			"jwt": newJWT,
		})
	}
}
