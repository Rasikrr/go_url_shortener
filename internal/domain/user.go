package domain

type User struct {
	Id          int    `json:"id,omitempty"`
	Email       string `json:"email"`
	EncPassword string `json:"encPassword"`
}
