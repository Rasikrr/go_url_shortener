package middleware

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

var (
	JwtSecret     string
	RefreshSecret string
)

func Init() {
	err := godotenv.Load("./.env")
	if err != nil {
		log.Fatal(err)
	}
	JwtSecret = os.Getenv("JWT_SECRET")
	RefreshSecret = os.Getenv("REFRESH_SECRET")

}
