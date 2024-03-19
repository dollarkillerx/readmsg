package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gofiber/fiber/v2"
	"github.com/pelletier/go-toml/v2"
	bolt "go.etcd.io/bbolt"

	"log"
	"os"
	"strconv"
	"strings"
)

type SMS struct {
	Body   string `json:"body"`
	Sender string `json:"sender"`
	Time   string `json:"time"`
}

type Config struct {
	TelegramBotToken string `json:"telegram_bot_token"`
	UserToken        string `json:"user_token"`
}

func main() {
	tmlCfg, err := os.ReadFile("server.toml")
	if err != nil {
		panic(err)
	}

	var cfg Config
	err = toml.Unmarshal(tmlCfg, &cfg)
	if err != nil {
		panic(err)
	}

	db, err := bolt.Open("sms_db.sql", 0777, nil)
	if err != nil {
		panic(err)
	}

	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("MyBucket"))
		if err != nil {
			return err
		}
		return nil
	})

	app := fiber.New()

	bot, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	go func() {
		for update := range updates {
			if update.Message != nil { // If we got a message
				log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

				if strings.TrimSpace(update.Message.Text) == "/start" {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Welcome to SMS Service!📱 请输入验证码")
					msg.ReplyToMessageID = update.Message.MessageID

					bot.Send(msg)
					return
				}

				if strings.TrimSpace(update.Message.Text) == cfg.UserToken {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "验证码正确!🔑 绑定成功!🔑")
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)

					db.Update(func(tx *bolt.Tx) error {
						b := tx.Bucket([]byte("MyBucket"))
						err := b.Put([]byte(strconv.Itoa(int(update.Message.Chat.ID))), []byte("0"))
						if err != nil {
							return err
						}
						return nil
					})
					return
				}

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "️🚫验证码错误!🔑请重新输入验证码!🔑")
				msg.ReplyToMessageID = update.Message.MessageID

				bot.Send(msg)
			}
		}
	}()

	var smsChannel = make(chan SMS, 100)

	go func() {
		for {
			select {
			case sms := <-smsChannel:
				db.View(func(tx *bolt.Tx) error {
					b := tx.Bucket([]byte("MyBucket"))
					b.ForEach(func(k, v []byte) error {
						chatID := string(k)
						atoi, err2 := strconv.Atoi(chatID)
						if err2 != nil {
							log.Println(err2)
							return nil
						}

						msg := tgbotapi.NewMessage(int64(atoi), "📱New SMS from: \n"+sms.Sender+"\n🐷Body:\n"+sms.Body+"\n⌚️Time:\n"+sms.Time)
						_, err := bot.Send(msg)
						if err != nil {
							log.Println(err)
							return nil
						}
						return nil
					})
					return nil
				})
			}
		}
	}()

	app.Get("/version", func(c *fiber.Ctx) error {
		return c.SendString("SMS Service v1.0.0")
	})

	app.Post("/send", func(c *fiber.Ctx) error {
		var sms SMS
		err := c.BodyParser(&sms)
		if err != nil {
			return c.Status(400).SendString("Invalid JSON")
		}

		smsChannel <- sms

		return c.SendString("Sending SMS")
	})

	log.Fatal(app.Listen(":7878"))
}
