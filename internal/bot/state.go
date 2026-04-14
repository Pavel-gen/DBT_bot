package bot

import "sync"

// Простейшее хранение состояния "ждём ввод дней для отчёта"
// В продакшене лучше использовать Redis, но для минималки хватит map + mutex
type StateStore struct {
	mu           sync.Mutex
	awaitingDays map[int64]bool // chatID -> true
}

func NewStateStore() *StateStore {
	return &StateStore{
		awaitingDays: make(map[int64]bool),
	}
}

func (s *StateStore) SetAwaitingDays(chatID int64, val bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.awaitingDays[chatID] = val
}

func (s *StateStore) IsAwaitingDays(chatID int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.awaitingDays[chatID]
}
