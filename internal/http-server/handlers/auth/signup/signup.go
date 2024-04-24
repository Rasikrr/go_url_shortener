package signup

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"go_url_chortener_api/internal/domain"
	"go_url_chortener_api/internal/http-server/customJson"
	resp "go_url_chortener_api/internal/lib/api/response"
	"go_url_chortener_api/internal/lib/hash"
	"go_url_chortener_api/internal/lib/logger/sl"
	"log/slog"
	"net/http"
	"unicode"
)

const passwordLength = 8
const salt = 5

var (
	passwordDoNotMatchErr = errors.New("password do not match")
	uppercaseErr          = errors.New("password must contain at least one uppercase character")
	lowercaseErr          = errors.New("password must contain at least one lowercase character")
	specialSymErr         = errors.New("password must contain al least one special symbol")
	lengthErr             = errors.New(fmt.Sprintf("password must contain at least %d characters", passwordLength))
	numberErr             = errors.New("password must contain at least one number")
)

type Request struct {
	Email     string `json:"email" validate:"required,email"`
	Password1 string `json:"password1" validate:""`
	Password2 string `json:"password2"`
}

type UserSaver interface {
	SaveUser(user *domain.User) error
}

func New(log *slog.Logger, userSaver UserSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.auth.signup.New"
		log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		req := new(Request)

		err := customJson.DecodeJson(r, req)
		if err != nil {
			log.Error("failed to decode json", sl.Err(err))
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

		if err := passwordsEqual(req.Password1, req.Password2); err != nil {
			log.Error(
				"invalid request",
				sl.Err(err),
				slog.String("password_1", req.Password1),
				slog.String("password_2", req.Password2),
			)
			customJson.WriteJson(w, http.StatusBadRequest, resp.Error("passwords do not match"))
			return
		}
		if err := isValidPassword(req.Password1); err != nil {
			log.Error("failed to validate password", sl.Err(err))
			customJson.WriteJson(w, http.StatusBadRequest, resp.Error(err.Error()))
			return
		}
		encPassword, err := hash.NewSHA1Hasher(salt).Hash(req.Password1)
		if err != nil {
			log.Error("failed to hash password", sl.Err(err))
			customJson.WriteJson(w, http.StatusInternalServerError, resp.Error("server error"))
			return
		}

		user := &domain.User{
			Email:       req.Email,
			EncPassword: encPassword,
		}

		if err := userSaver.SaveUser(user); err != nil {
			log.Error("failed to save user", sl.Err(err))
			customJson.WriteJson(w, http.StatusInternalServerError, resp.Error("server error"))
			return
		}
		customJson.WriteJson(w, http.StatusOK, resp.OK())
	}
}

func passwordsEqual(p1, p2 string) error {
	if p1 != p2 {
		return passwordDoNotMatchErr
	}
	return nil
}

func isValidPassword(password string) error {
	const fn = "handlers.auth.signup.isValidPassword"

	var (
		upper      bool
		specialSym bool
		number     bool
		lower      bool
	)
	if len(password) < passwordLength {
		return fmt.Errorf("%s : %w", fn, lengthErr)
	}
	for _, el := range password {
		switch {
		case unicode.IsUpper(el):
			upper = true
		case unicode.IsLower(el):
			lower = true
		case unicode.IsNumber(el):
			number = true
		case unicode.IsSymbol(el) || unicode.IsPunct(el):
			specialSym = true
		}
	}
	if !upper {
		return fmt.Errorf("%s : %w", fn, uppercaseErr)
	}
	if !lower {
		return fmt.Errorf("%s : %w", fn, lowercaseErr)
	}
	if !specialSym {
		return fmt.Errorf("%s : %w", fn, specialSymErr)
	}
	if !number {
		return fmt.Errorf("%s : %w", fn, numberErr)
	}
	return nil
}
