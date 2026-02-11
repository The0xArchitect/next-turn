// internal/config/env.go
package config

import (
	"fmt"
	"os"
)

var (
	BotToken    = os.Getenv("BOT_TOKEN")
	DatabaseURL = os.Getenv("DB_URL")
)

func Validate() {
	fmt.Println(BotToken)
	fmt.Println(DatabaseURL)
	if BotToken == "" {
		panic("BOT_TOKEN is required")
	}
	if DatabaseURL == "" {
		panic("DB_URL is required")
	}
}
