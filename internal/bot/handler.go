package bot

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Run(token string) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Invalid bot token: %v", err)
	}
	log.Printf("Bot authorized as @%s", bot.Self.UserName);

	u := tgbotapi.NewUpdate(0);
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	log.Println("Bot is listening for updates")

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		userText := update.Message.Text 
		log.Printf("[%d] %s", chatID, userText)

		reply := fmt.Sprintf("Echo: %s (LLM and DB pending)", userText)

		msg := tgbotapi.NewMessage(chatID, reply)

		if _, err := bot.Send(msg); err != nil {
			log.Printf("Send error : %v", err)
		}

	}
}