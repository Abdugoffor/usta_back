package telegram_auth_service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type tgUpdate struct {
	UpdateID int64 `json:"update_id"`
	Message  *struct {
		Text string `json:"text"`
		From struct {
			ID        int64  `json:"id"`
			Username  string `json:"username"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		} `json:"from"`
	} `json:"message"`
}

type tgResponse struct {
	OK     bool       `json:"ok"`
	Result []tgUpdate `json:"result"`
}

func StartPoller(ctx context.Context, botToken string, s Service) {
	if botToken == "" {
		log.Println("⚠️  TELEGRAM_BOT_TOKEN bo'sh — Telegram login o'chirilgan")
		return
	}

	go func() {
		var offset int64
		client := &http.Client{Timeout: 35 * time.Second}

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?timeout=30&offset=%d",
				botToken, offset)

			resp, err := client.Get(url)
			if err != nil {
				log.Printf("tg poll error: %v", err)
				time.Sleep(3 * time.Second)
				continue
			}
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			var r tgResponse
			if err := json.Unmarshal(body, &r); err != nil || !r.OK {
				log.Printf("tg parse error: %v body=%s", err, string(body))
				time.Sleep(3 * time.Second)
				continue
			}

			for _, u := range r.Result {
				offset = u.UpdateID + 1
				if u.Message == nil {
					continue
				}
				text := strings.TrimSpace(u.Message.Text)
				if !strings.HasPrefix(text, "/start") {
					continue
				}
				parts := strings.Fields(text)
				if len(parts) < 2 {
					continue
				}
				loginToken := parts[1]
				from := u.Message.From

				err := s.HandleTelegramStart(ctx, loginToken, from.ID,
					from.Username, from.FirstName, from.LastName)
				if err != nil {
					log.Printf("tg login error (token=%s): %v", loginToken, err)
				} else {
					log.Printf("✓ tg login: tg_id=%d token=%s", from.ID, loginToken)
				}
			}
		}
	}()

	log.Println("✓ Telegram poller ishga tushdi")
}
