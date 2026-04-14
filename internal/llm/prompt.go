package llm 

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type PromptLoader struct {
	prompts map[string]string
	mu 	    sync.RWMutex
}

func newPromptLoader(promptsDir string) (*PromptLoader, error) {
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
}