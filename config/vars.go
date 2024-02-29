package config

import (
	"github.com/joho/godotenv"
	"os"
)

func setupEnvVars() {
	var port string
	if os.Getenv("ENV") == "production" {
		godotenv.Load(".env")
		port = os.Getenv("PORT")
	} else {
		port = "3001"
	}
	os.Setenv("PORT", port)
}
