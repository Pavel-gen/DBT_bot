package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"mybot/internal/bot"
	"mybot/internal/db"
	"mybot/internal/llm"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load() // Игнорируем ошибку, если .env нет

	// Проверка токенов
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("❌ TELEGRAM_BOT_TOKEN required")
	}
	if os.Getenv("OPENROUTER_API_KEY") == "" {
		log.Fatal("❌ OPENROUTER_API_KEY required")
	}

	// Опционально: БД (пока не используем, но подключение готово)
	dsn := os.Getenv("DATABASE_URL")
	if dsn != "" {
		ctx := context.Background()
		if err := db.Connect(ctx, dsn); err != nil {
			log.Printf("⚠️ DB connect failed (non-fatal for now): %v", err)
		} else {
			defer db.Pool.Close()
		}
	}

	// Загрузка промптов
	promptsDir := filepath.Join(".", "prompts") // Или укажи абсолютный путь
	loader, err := llm.NewPromptLoader(promptsDir)
	if err != nil {
		log.Fatalf("❌ Failed to load prompts: %v", err)
	}

	// Инициализация и запуск бота
	handler, err := bot.NewHandler(token, loader)
	if err != nil {
		log.Fatalf("❌ Failed to init bot: %v", err)
	}

	log.Println("🚀 Bot starting...")
	handler.Start() // Блокирующий вызов
}
