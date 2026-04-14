package bot

import (
	"context"
	"encoding/json" // ← добавь это, если ещё нет
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"mybot/internal/db" // ← тоже проверь, что есть
	"mybot/internal/llm"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	bot    *tgbotapi.BotAPI
	llmSvc *llm.Service
	state  *StateStore // ← новое
}

func NewHandler(token string, promptLoader *llm.PromptLoader) (*Handler, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("init bot: %w", err)
	}
	log.Printf("🤖 Bot authorized as @%s", bot.Self.UserName)
	return &Handler{
		bot:    bot,
		llmSvc: llm.NewService(promptLoader),
		state:  NewStateStore(),
	}, nil
}

func (h *Handler) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := h.bot.GetUpdatesChan(u)
	log.Println("🟢 Bot is listening...")

	for update := range updates {
		if update.Message == nil || update.Message.Text == "" {
			continue
		}
		// Запускаем обработку в горутине, чтобы не блокировать получение новых сообщений
		go h.handleMessage(update.Message)
	}
}

func (h *Handler) handleMessage(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	userText := strings.TrimSpace(msg.Text)

	// === ОБРАБОТКА КОМАНДЫ /file ===
	if userText == "/file" {
		h.state.SetAwaitingDays(chatID, true)
		h.sendText(chatID, "За сколько дней выгрузить отчёт? (введите число или 'all' для всех записей)")
		return
	}

	// === ЕСЛИ ЖДЁМ ВВОД ДНЕЙ ===
	if h.state.IsAwaitingDays(chatID) {
		h.state.SetAwaitingDays(chatID, false) // сбрасываем состояние

		var days int
		if strings.ToLower(userText) == "all" {
			days = 0
		} else if num, err := fmt.Sscanf(userText, "%d", &days); num == 0 || err != nil {
			days = 10 // дефолт
		}

		// Запускаем генерацию в горутине (может быть долго)
		go h.handleFileExport(chatID, days)
		return
	}

	log.Printf("📨 [Step 1] Message received from %d: %s", chatID, userText)

	// Создаем контекст с таймаутом 60 секунд (защита от вечного зависания)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// --- ШАГ 2: Вызов LLM ---
	log.Printf("🧠 [Step 2] Calling LLM with prompt 'DBT'...")
	rawResp, err := h.llmSvc.Call(ctx, "DBT", userText, 1200, 0.9)
	if err != nil {
		log.Printf("❌ [Step 2 Error] LLM call failed: %v", err)
		h.sendText(chatID, "⚠️ Ошибка связи с мозгом: "+err.Error())
		return
	}
	log.Printf("✅ [Step 2 Done] LLM response received (%d chars)", len(rawResp))

	// --- ШАГ 3: Парсинг ---
	log.Printf("🔍 [Step 3] Parsing JSON response...")
	aiText, err := llm.GenerateReadableText(rawResp)
	if err != nil {
		log.Printf("⚠️ [Step 3 Warn] Parse failed: %v", err)
		// Если парсинг упал, отправляем сырой ответ, чтобы пользователь не молчал
		aiText = fmt.Sprintf("🔧 Raw mode:\n%s", rawResp)
	} else {
		log.Printf("✅ [Step 3 Done] Text generated")
	}

	// --- ШАГ 4: Отправка ---
	log.Printf("📤 [Step 4] Sending reply to %d...", chatID)
	h.sendText(chatID, aiText)
	log.Printf("🏁 [Step 4 Done] Reply sent.")

	go func() {
		var resp llm.DBTResponse
		if err := json.Unmarshal([]byte(rawResp), &resp); err != nil {
			log.Printf("⚠️ [DB] Failed to unmarshal for saving: %v", err)
			return
		}

		// Подготовка физиологии в JSONB
		var physiologyJSON []byte
		if resp.Physiology != nil {
			physiologyJSON, _ = json.Marshal(resp.Physiology)
		}

		// Вспомогательная функция для безопасного доступа к вложенным полям
		// (чтобы не было паники при nil-полях)
		_, err := db.SaveInteraction(
			context.Background(), // отдельный контекст, не зависит от таймаута запроса
			msg.From.ID,
			msg.From.UserName,
			msg.From.FirstName,
			msg.From.LastName,
			userText, // текст пользователя
			aiText,   // человекочитаемый ответ
			rawResp,  // сырой JSON от LLM
			// Chain fields
			getString(func() *string {
				if resp.Chain != nil {
					return &resp.Chain.Trigger
				}
				return nil
			}()),
			getString(func() *string {
				if resp.Chain != nil {
					return &resp.Chain.Thought
				}
				return nil
			}()),
			getString(func() *string {
				if resp.Chain != nil && resp.Chain.Emotion != nil {
					return &resp.Chain.Emotion.Name
				}
				return nil
			}()),
			getString(func() *string {
				if resp.Chain != nil {
					return &resp.Chain.Action
				}
				return nil
			}()),
			getString(func() *string {
				if resp.Chain != nil {
					return &resp.Chain.Consequence
				}
				return nil
			}()),
			// Analysis fields
			getString(func() *string {
				if resp.Analysis != nil {
					return &resp.Analysis.Goal
				}
				return nil
			}()),
			getString(func() *string {
				if resp.Analysis != nil {
					return &resp.Analysis.IneffectivenessReason
				}
				return nil
			}()),
			getString(func() *string {
				if resp.Analysis != nil {
					return &resp.Analysis.HiddenNeed
				}
				return nil
			}()),
			// Emotion intensity
			getInt(func() *int {
				if resp.Chain != nil && resp.Chain.Emotion != nil {
					return &resp.Chain.Emotion.Intensity
				}
				return nil
			}()),
			// Arrays
			resp.Patterns,
			resp.Alternatives,
			// JSONB
			physiologyJSON,
		)

		if err != nil {
			log.Printf("❌ [DB] SaveInteraction failed: %v", err)
		} else {
			log.Printf("✅ [DB] Interaction saved for user %d", msg.From.ID)
		}
	}()
}

func (h *Handler) handleFileExport(chatID int64, days int) {
	// Индикатор "печатает..."
	label := "все записи"
	if days > 0 {
		label = fmt.Sprintf("%d дн.", days)
	}
	h.sendText(chatID, fmt.Sprintf("🔄 Генерирую отчёт за %s...", label))

	ctx := context.Background()

	// 1. Получаем данные из БД
	// Telegram chatID == telegram_id в нашей БД
	interactions, err := db.GetInteractionsByPeriod(ctx, chatID, days)
	if err != nil {
		log.Printf("❌ [Report] DB query failed: %v", err)
		h.sendText(chatID, "❌ Ошибка при чтении данных из БД")
		return
	}

	if len(interactions) == 0 {
		h.sendText(chatID, fmt.Sprintf("📭 Нет записей за последние %d дн.", days))
		return
	}

	// 2. Форматируем отчёт
	reportText := FormatExportReport(interactions, days)

	// 3. Создаём временный файл (путь для os.Remove позже)
	tmpPath, err := GenerateReportFile(reportText)
	if err != nil {
		log.Printf("❌ [Report] File generation failed: %v", err)
		h.sendText(chatID, "❌ Ошибка при создании файла")
		return
	}
	defer os.Remove(tmpPath) // удаляем после отправки

	// 4. Читаем файл в память для отправки с кастомным именем
	fileContent, err := os.ReadFile(tmpPath)
	if err != nil {
		log.Printf("❌ [Report] Read file failed: %v", err)
		h.sendText(chatID, "❌ Ошибка при чтении файла")
		return
	}

	// 5. Формируем имя файла: отчёт_7дн_2026-04-14.txt
	safeLabel := strings.ReplaceAll(label, " ", "_") // убираем пробелы из имени файла
	filename := fmt.Sprintf("otchet_%s_%s.txt",
		safeLabel,
		time.Now().Format("2006-01-02"),
	)

	// 6. Отправляем файл через FileBytes (единственный способ задать имя)
	doc := tgbotapi.NewDocument(chatID, tgbotapi.FileBytes{
		Name:  filename,
		Bytes: fileContent,
	})

	if _, err := h.bot.Send(doc); err != nil {
		log.Printf("❌ [Report] Send file failed: %v", err)
		h.sendText(chatID, "❌ Ошибка при отправке файла")
		return
	}

	log.Printf("✅ [Report] Sent %d interactions to chat %d", len(interactions), chatID)
	h.sendText(chatID, fmt.Sprintf("✅ Готово. Записей в отчёте: %d", len(interactions)))
}

func getString(val *string) string {
	if val != nil {
		return *val
	}
	return ""
}

func getInt(val *int) int {
	if val != nil {
		return *val
	}
	return 0
}

func (h *Handler) sendText(chatID int64, text string) {
	// Экранируем спецсимволы, если используем Markdown, или ставим пустую строку для простого текста
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "" // Отключаем форматирование, чтобы не ломать текст скобками

	_, err := h.bot.Send(msg)
	if err != nil {
		log.Printf("❌ [Send Error] Failed to send message: %v", err)
	}
}
