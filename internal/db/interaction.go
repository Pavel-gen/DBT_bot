package db

import (
	"context"
	"fmt"
)

// SaveInteraction сохраняет сообщение + анализ в БД
// Возвращает ID созданной записи interaction (или ошибку)
func SaveInteraction(
	ctx context.Context,
	telegramID int64,
	username, firstName, lastName string,
	userText, botText, rawResponse string,
	trigger, thought, emotionName, action, consequence, goal, ineffectivenessReason, hiddenNeed string,
	emotionIntensity int,
	patterns, alternatives []string,
	physiologyJSON []byte, // JSONB как байты
) (string, error) {
	// 1. Создаём/обновляем юзера (UPSERT)
	var userID int64
	err := Pool.QueryRow(ctx, `
		INSERT INTO users (telegram_id, username, first_name, last_name)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (telegram_id) 
		DO UPDATE SET 
			username = COALESCE(EXCLUDED.username, users.username),
			first_name = COALESCE(EXCLUDED.first_name, users.first_name),
			last_name = COALESCE(EXCLUDED.last_name, users.last_name)
		RETURNING id
	`, telegramID, username, firstName, lastName).Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("upsert user: %w", err)
	}

	// 2. Сохраняем сообщение пользователя
	var msgID string
	err = Pool.QueryRow(ctx, `
		INSERT INTO messages (user_id, content, sender)
		VALUES ($1, $2, 'user')
		RETURNING id
	`, userID, userText).Scan(&msgID)
	if err != nil {
		return "", fmt.Errorf("insert user message: %w", err)
	}

	// 3. Сохраняем сообщение бота (опционально, но полезно для лога)
	_, err = Pool.Exec(ctx, `
		INSERT INTO messages (user_id, content, sender)
		VALUES ($1, $2, 'bot')
	`, userID, botText)
	if err != nil {
		return "", fmt.Errorf("insert bot message: %w", err)
	}

	// 4. Сохраняем interaction
	// PostgreSQL массивы: $10::TEXT[] — явное приведение типа
	_, err = Pool.Exec(ctx, `
		INSERT INTO interactions (
			user_id, message_id,
			trigger, thought, emotion_name, emotion_intensity, action, consequence,
			patterns, goal, ineffectiveness_reason, hidden_need,
			physiology, alternatives, raw_response
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9::TEXT[], $10, $11, $12,
			$13::JSONB, $14::TEXT[], $15
		)
	`,
		userID, msgID,
		trigger, thought, emotionName, emotionIntensity, action, consequence,
		patterns, goal, ineffectivenessReason, hiddenNeed,
		physiologyJSON, alternatives, rawResponse,
	)
	if err != nil {
		return "", fmt.Errorf("insert interaction: %w", err)
	}

	return msgID, nil
}
