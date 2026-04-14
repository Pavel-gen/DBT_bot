package llm

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type PromptLoader struct {
	prompts map[string]string
	mu      sync.RWMutex
}

func NewPromptLoader(promptsDir string) (*PromptLoader, error) {
	pl := &PromptLoader{
		prompts: make(map[string]string),
	}

	promptFiles := map[string]string{
		"DBT":      "DBTpromt1.txt",
		"BEHAVIOR": "BehaviorAnalysisPrompt.txt",
		"CORE":     "core_prompt.txt",
		"JOURNAL":  "message_to_journal.txt",
		"STATS":    "message_to_daily_stats.txt",
	}

	for key, filename := range promptFiles {
		fullPath := filepath.Join(promptsDir, filename)
		content, err := os.ReadFile(fullPath)

		if err != nil {
			fmt.Printf("Промпт %s не загружен: %v\n", key, err)
			continue
		}

		pl.prompts[key] = string(content)
		fmt.Printf("Промпт %s загружен \n", key)
	}

	return pl, nil
}

func (pl *PromptLoader) Get(key string) string {
	pl.mu.RLock()
	defer pl.mu.RUnlock()
	return pl.prompts[key]
}
