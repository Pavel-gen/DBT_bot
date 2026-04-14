package db

import (
	"context"
	"fmt"
	"time"
)

// InteractionReport — плоская структура для отчёта (удобно для форматирования)
type InteractionReport struct {
	UserMessage           string
	BotResponse           string
	Trigger               string
	Thought               string
	EmotionName           string
	EmotionIntensity      int
	Action                string
	Consequence           string
	Patterns              []string
	Goal                  string
	IneffectivenessReason string
	HiddenNeed            string
	Physiology            string // JSON как строка, распарсим при форматировании если надо
	Alternatives          []string
	CreatedAt             time.Time
}

// GetInteractionsByPeriod возвращает взаимодействия за последние `days` дней
// days=0 означает "все записи"
func GetInteractionsByPeriod(ctx context.Context, telegramID int64, days int) ([]InteractionReport, error) {
	var query string
	var args []interface{}

	if days > 0 {
		query = `
			SELECT 
				m.content as user_message,
				i.raw_response, -- пока не используем, но можно для отладки
				i.trigger, i.thought, i.emotion_name, i.emotion_intensity,
				i.action, i.consequence, i.patterns, i.goal,
				i.ineffectiveness_reason, i.hidden_need,
				i.physiology::text, i.alternatives, i.created_at
			FROM interactions i
			JOIN messages m ON i.message_id = m.id
			JOIN users u ON i.user_id = u.id
			WHERE u.telegram_id = $1 
				AND i.created_at >= NOW() - INTERVAL '1 day' * $2
			ORDER BY i.created_at DESC
		`
		args = []interface{}{telegramID, days}
	} else {
		query = `
			SELECT 
				m.content as user_message,
				i.raw_response,
				i.trigger, i.thought, i.emotion_name, i.emotion_intensity,
				i.action, i.consequence, i.patterns, i.goal,
				i.ineffectiveness_reason, i.hidden_need,
				i.physiology::text, i.alternatives, i.created_at
			FROM interactions i
			JOIN messages m ON i.message_id = m.id
			JOIN users u ON i.user_id = u.id
			WHERE u.telegram_id = $1
			ORDER BY i.created_at DESC
		`
		args = []interface{}{telegramID}
	}

	rows, err := Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query interactions: %w", err)
	}
	defer rows.Close()

	var results []InteractionReport
	for rows.Next() {
		var r InteractionReport
		var physiologyStr string
		err := rows.Scan(
			&r.UserMessage,
			new(string), // raw_response (пока не используем)
			&r.Trigger, &r.Thought, &r.EmotionName, &r.EmotionIntensity,
			&r.Action, &r.Consequence, &r.Patterns, &r.Goal,
			&r.IneffectivenessReason, &r.HiddenNeed,
			&physiologyStr, &r.Alternatives, &r.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		// Physiology пока храним как строку, при форматировании можно распарсить если надо
		r.Physiology = physiologyStr
		results = append(results, r)
	}

	return results, nil
}
