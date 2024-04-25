package refresh

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"go_url_chortener_api/internal/http-server/middleware"
	"net/http"
	"time"
)

type tokenClaims struct {
	UserId int `json:"userId"`
	*jwt.RegisteredClaims
}

func CreateRefresh(userId int) (string, error) {
	claims := &tokenClaims{
		UserId: userId,
		RegisteredClaims: &jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 30)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(middleware.RefreshSecret))
}

func validateRefresh(tknStr string) (*jwt.Token, error) {
	token, err := jwt.Parse(tknStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(middleware.RefreshSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return token, nil
}

func SetRefreshCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		HttpOnly: true,
		Name:     "refresh-token",
		Value:    token,
		Path:     "/",
	})
}

func getRefreshCookie(r *http.Request) (string, error) {
	const fn = "handlers.refresh.getRefreshCookie"
	cookie, err := r.Cookie("refresh-token")

	if err != nil {
		return "", fmt.Errorf("%s : %w", fn, err)
	}
	return cookie.Value, nil

}
