package myJwt

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"go_url_chortener_api/internal/http-server/customJson"
	"go_url_chortener_api/internal/http-server/middleware"
	"go_url_chortener_api/internal/lib/api/response"
	"go_url_chortener_api/internal/lib/logger/sl"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const jwtTokenLifetime = time.Minute * 15

type jwtClaims struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func permissionDenied(w http.ResponseWriter) {
	customJson.WriteJson(w, http.StatusForbidden, response.Error("authorization failed"))
}

func JwtMiddleware(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr, err := getJWT(r)
			if err != nil {
				log.Error("failed to get JWT", sl.Err(err))
				permissionDenied(w)
				return
			}
			token, err := validateJWT(tokenStr)
			if err != nil {
				log.Error("failed to validate JWT", sl.Err(err))
				permissionDenied(w)
				return
			}
			id, err := getIdCookie(r)
			if err != nil {
				log.Error("failed get Id cookie", sl.Err(err))
				permissionDenied(w)
				return
			}
			tokenClaims := token.Claims.(jwt.MapClaims)
			if int(tokenClaims["id"].(float64)) != id {
				log.Error("id's from claims and from cookie do not match", sl.Err(err))
				permissionDenied(w)
				return
			}
			log.Info("jwt middleware ended successfully")
			next.ServeHTTP(w, r)
		})
	}
}

func getJWT(r *http.Request) (string, error) {
	tokenBearer := r.Header.Get("Authorization")
	if tokenBearer == "" {
		return "", fmt.Errorf("token is empty")
	}
	token := strings.Split(tokenBearer, " ")
	if len(token) != 2 {
		return "", fmt.Errorf("token is invalid: %s", tokenBearer)
	}
	return token[1], nil
}

func getIdCookie(r *http.Request) (int, error) {
	cookie, err := r.Cookie("id")
	if err != nil {
		return -1, fmt.Errorf("failed to get id cookie")
	}
	id, err := strconv.Atoi(cookie.Value)
	if err != nil {
		return -1, fmt.Errorf("invalid id")
	}
	return id, err
}

func CreateJWT(id int, email string) (string, error) {
	claims := jwtClaims{
		Id:    id,
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtTokenLifetime)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(middleware.JwtSecret))
}

func SetJWTHeader(w http.ResponseWriter, token string) {
	w.Header().Set("Authorization", "Bearer "+token)
}

func SetIdCookie(w http.ResponseWriter, id int) {
	idStr := strconv.Itoa(id)
	http.SetCookie(w, &http.Cookie{
		Name:     "id",
		Value:    idStr,
		Path:     "/",
		HttpOnly: true,
	})
}

func validateJWT(tknStr string) (*jwt.Token, error) {
	token, err := jwt.Parse(tknStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(middleware.JwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return token, nil
}
