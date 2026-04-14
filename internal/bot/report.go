package bot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mybot/internal/db"
)

// wrapText — перенос длинных строк (как в твоём JS)
func wrapText(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}

	var result strings.Builder
	var line strings.Builder
	words := strings.Fields(text)

	for _, word := range words {
		if line.Len()+len(word)+1 > maxWidth && line.Len() > 0 {
			result.WriteString(line.String() + "\n")
			line.Reset()
		}
		if line.Len() > 0 {
			line.WriteString(" ")
		}
		line.WriteString(word)
	}
	if line.Len() > 0 {
		result.WriteString(line.String())
	}
	return result.String()
}

// wrapAndIndent — wrap + отступ для вложенных строк
func wrapAndIndent(text string, maxWidth, indent int) string {
	wrapped := wrapText(text, maxWidth-indent)
	lines := strings.Split(wrapped, "\n")
	for i := 1; i < len(lines); i++ {
		lines[i] = strings.Repeat(" ", indent) + lines[i]
	}
	return strings.Join(lines, "\n")
}

// FormatExportReport — чистый формат: пользователь с отступом, бот без, физиология распарсена
func FormatExportReport(interactions []db.InteractionReport, days int) string {
	var since time.Time
	if days > 0 {
		since = time.Now().AddDate(0, 0, -days)
	}

	now := time.Now()
	dateRange := "всё время"
	if days > 0 {
		dateRange = fmt.Sprintf("%s – %s",
			since.Format("02.01.2006"),
			now.Format("02.01.2006"),
		)
	}

	var lines []string
	lines = append(lines,
		fmt.Sprintf("Отчёт за %d дн. (%s)", days, dateRange),
		fmt.Sprintf("Всего записей: %d", len(interactions)),
		"",
	)

	for _, it := range interactions {
		// === ПОЛЬЗОВАТЕЛЬ: с переносом и отступом ===
		if it.UserMessage != "" {
			wrapped := wrapText(strings.TrimSpace(it.UserMessage), 76)
			indented := strings.ReplaceAll(wrapped, "\n", "\n    ")
			lines = append(lines, fmt.Sprintf("Пользователь:\n    %s", indented))
			lines = append(lines, "") // пустая строка для разделения
		}

		// === БОТ: чистый формат, без отступов на второй строке ===
		lines = append(lines, "Бот: Анализ ситуации")
		lines = append(lines, fmt.Sprintf("Триггер: %s", it.Trigger))
		lines = append(lines, fmt.Sprintf("Мысль: %s", it.Thought))
		lines = append(lines, fmt.Sprintf("Эмоция: %s (%d/10)", it.EmotionName, it.EmotionIntensity))
		lines = append(lines, fmt.Sprintf("Действие: %s", it.Action))
		lines = append(lines, fmt.Sprintf("Последствие: %s", it.Consequence))

		// Паттерны
		if len(it.Patterns) > 0 {
			lines = append(lines, fmt.Sprintf("Паттерны: %s", strings.Join(it.Patterns, ", ")))
		}

		lines = append(lines, fmt.Sprintf("Цель: %s", it.Goal))
		lines = append(lines, fmt.Sprintf("Причина неэффективности: %s", it.IneffectivenessReason))
		lines = append(lines, fmt.Sprintf("Скрытая потребность: %s", it.HiddenNeed))

		// Физиология: парсим JSON и выводим читаемо
		if it.Physiology != "" && it.Physiology != "{}" {
			physLines := formatPhysiology(it.Physiology)
			if len(physLines) > 0 {
				lines = append(lines, "Физиология:")
				lines = append(lines, physLines...)
			}
		}

		// Альтернативы
		if len(it.Alternatives) > 0 {
			lines = append(lines, "Альтернативы:")
			for i, alt := range it.Alternatives {
				lines = append(lines, fmt.Sprintf("    %d. %s", i+1, alt))
			}
		}

		lines = append(lines, fmt.Sprintf("Дата анализа: %s", it.CreatedAt.Format("02.01.2006 15:04")))
		lines = append(lines, strings.Repeat("─", 80))
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// formatPhysiology — парсит JSON и возвращает читаемые строки
func formatPhysiology(jsonStr string) []string {
	var phys struct {
		AmygdalaMechanism   string `json:"amygdala_mechanism"`
		BinaryProtocol      string `json:"binary_protocol"`
		PhysicalMarkers     string `json:"physical_markers"`
		PfkOverrideStrategy string `json:"pfk_override_strategy"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &phys); err != nil {
		// Если не распарсилось — возвращаем одну строку с сырым текстом (фолбэк)
		return []string{fmt.Sprintf("    %s", jsonStr)}
	}

	var lines []string
	if phys.AmygdalaMechanism != "" {
		lines = append(lines, fmt.Sprintf("    Амигдала: %s", phys.AmygdalaMechanism))
	}
	if phys.BinaryProtocol != "" {
		lines = append(lines, fmt.Sprintf("    Протокол: %s", phys.BinaryProtocol))
	}
	if phys.PhysicalMarkers != "" {
		lines = append(lines, fmt.Sprintf("    Тело: %s", phys.PhysicalMarkers))
	}
	if phys.PfkOverrideStrategy != "" {
		lines = append(lines, fmt.Sprintf("    ПФК: %s", phys.PfkOverrideStrategy))
	}
	return lines
}

// GenerateReportFile — создаёт временный файл с отчётом, возвращает путь
func GenerateReportFile(reportText string) (string, error) {
	filename := fmt.Sprintf("report_%d.txt", time.Now().Unix())
	filepath := filepath.Join(os.TempDir(), filename) // используем системную temp-папку

	err := os.WriteFile(filepath, []byte(reportText), 0644)
	if err != nil {
		return "", fmt.Errorf("write report file: %w", err)
	}
	return filepath, nil
}
