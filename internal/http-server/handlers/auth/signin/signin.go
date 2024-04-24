package signin

import "log/slog"

type Request struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func New(log *slog.Logger) {}
