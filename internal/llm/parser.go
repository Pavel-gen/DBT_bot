package llm

import (
	"encoding/json"
	"fmt"
	"strings"
)

// DBTResponse — точная структура под твой JSON от LLM
type DBTResponse struct {
	Chain *struct {
		Trigger string `json:"trigger"`
		Thought string `json:"thought"`
		Emotion *struct {
			Name      string `json:"name"`
			Intensity int    `json:"intensity"`
		} `json:"emotion"`
		Action      string `json:"action"`
		Consequence string `json:"consequence"`
	} `json:"chain"`
	Patterns []string `json:"patterns"`
	Analysis *struct {
		Goal                  string `json:"goal"`
		IneffectivenessReason string `json:"ineffectiveness_reason"`
		HiddenNeed            string `json:"hidden_need"`
	} `json:"analysis"`
	Physiology *struct {
		AmygdalaMechanism   string `json:"amygdala_mechanism"`
		BinaryProtocol      string `json:"binary_protocol"`
		PhysicalMarkers     string `json:"physical_markers"`
		PfkOverrideStrategy string `json:"pfk_override_strategy"`
	} `json:"physiology"`
	Alternatives []string `json:"alternatives"`
}

// GenerateReadableText — 1:1 копия твоего JS-метода, только на Go
func GenerateReadableText(rawJSON string) (string, error) {
	var resp DBTResponse
	if err := json.Unmarshal([]byte(rawJSON), &resp); err != nil {
		return "", fmt.Errorf("parse JSON: %w", err)
	}

	var lines []string

	// 1. ЦЕПЬ
	lines = append(lines, "1. ЦЕПЬ:")
	if resp.Chain != nil {
		lines = append(lines, fmt.Sprintf("Триггер — %s", orDash(resp.Chain.Trigger)))
		lines = append(lines, fmt.Sprintf("Мысль — \"%s\"", orDash(resp.Chain.Thought)))

		emotionName := ""
		emotionIntensity := 0
		if resp.Chain.Emotion != nil {
			emotionName = resp.Chain.Emotion.Name
			emotionIntensity = resp.Chain.Emotion.Intensity
		}
		lines = append(lines, fmt.Sprintf("Эмоция — %s (%d/10)", orDash(emotionName), emotionIntensity))

		lines = append(lines, fmt.Sprintf("Действие — %s", orDash(resp.Chain.Action)))
		lines = append(lines, fmt.Sprintf("Последствие — %s", orDash(resp.Chain.Consequence)))
	} else {
		lines = append(lines, "Триггер — -")
		lines = append(lines, "Мысль — \"-\"")
		lines = append(lines, "Эмоция — - (0/10)")
		lines = append(lines, "Действие — -")
		lines = append(lines, "Последствие — -")
	}

	// 2. ПАТТЕРНЫ
	patternsStr := "-"
	if len(resp.Patterns) > 0 {
		patternsStr = strings.Join(resp.Patterns, ", ")
	}
	lines = append(lines, fmt.Sprintf("2. ПАТТЕРНЫ: %s", patternsStr))

	// 3. АНАЛИЗ
	lines = append(lines, "3. АНАЛИЗ:")
	if resp.Analysis != nil {
		lines = append(lines, fmt.Sprintf("Цель — %s", orDash(resp.Analysis.Goal)))
		lines = append(lines, fmt.Sprintf("Не сработало — %s", orDash(resp.Analysis.IneffectivenessReason)))
		lines = append(lines, fmt.Sprintf("Скрытая потребность — %s", orDash(resp.Analysis.HiddenNeed)))
	} else {
		lines = append(lines, "Цель — -")
		lines = append(lines, "Не сработало — -")
		lines = append(lines, "Скрытая потребность — -")
	}

	// 4. ФИЗИОЛОГИЯ (только если есть данные)
	if resp.Physiology != nil {
		lines = append(lines, "4. ФИЗИОЛОГИЯ:")
		lines = append(lines, fmt.Sprintf("Амигдала: %s", orDash(resp.Physiology.AmygdalaMechanism)))
		lines = append(lines, fmt.Sprintf("Протокол: %s", orDash(resp.Physiology.BinaryProtocol)))
		lines = append(lines, fmt.Sprintf("Тело: %s", orDash(resp.Physiology.PhysicalMarkers)))
		lines = append(lines, fmt.Sprintf("ПФК: %s", orDash(resp.Physiology.PfkOverrideStrategy)))
	}

	// 5. АЛЬТЕРНАТИВЫ
	lines = append(lines, "5. АЛЬТЕРНАТИВЫ:")
	if len(resp.Alternatives) > 0 {
		for i, alt := range resp.Alternatives {
			lines = append(lines, fmt.Sprintf("%d) %s", i+1, alt))
		}
	} else {
		lines = append(lines, "-")
	}

	return strings.Join(lines, "\n"), nil
}

// orDash — вспомогательная функция: если строка пустая, возвращает "-"
func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}
