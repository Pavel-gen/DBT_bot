package main

import (
	"context"
	"log"
	"os"

	"mybot/internal/bot"
	"mybot/internal/db"
	"github.com/joho/godotenv"
)

func main() {
	
	if err := godotenv.Load(); err != nil {
		log.Println(" .env not found, using system vars")
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("❌ TELEGRAM_BOT_TOKEN is required")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("❌ DATABASE_URL is required")
	}

	// 2. Подключаем БД
	ctx := context.Background()
	if err := db.Connect(ctx, dsn); err != nil {
		log.Fatalf("🔥 DB connection failed: %v", err)
	}
	defer db.Pool.Close()

	bot.Run(token)
}