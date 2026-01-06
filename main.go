package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	clientTimeout = 5 * time.Second
	sensorsURL    = "http://localhost:9111/sensor-data"
)

// SensorData struct to hold the sensor information
type SensorData struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Pressure    float64 `json:"pressure"`
}

// Authorized Telegram user IDs
var authorizedUsers map[int64]struct{}

// Fetch the sensor data from the local endpoint
func getSensorData() (*SensorData, error) {
	client := &http.Client{Timeout: clientTimeout}
	resp, err := client.Get(sensorsURL)
	if err != nil {
		return nil, fmt.Errorf("could not fetch sensor data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response status: %s", resp.Status)
	}

	var data SensorData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("could not decode sensor data: %v", err)
	}
	return &data, nil
}

// Create a new reply keyboard with a button
func createReplyKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🌤️ Get Weather Data"),
		),
	)
}

// Handle the received command
func handleMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	// Check if user is authorized
	if _, authorized := authorizedUsers[update.Message.From.ID]; !authorized {
		log.Printf("Unauthorized start attempt by user ID: %d", update.Message.From.ID)
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "You are not authorized to use this bot."))
		return
	}

	// Handle start command
	if update.Message.Text == "/start" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Welcome! Use the keyboard below to get weather data.")
		msg.ReplyMarkup = createReplyKeyboard()
		bot.Send(msg)
	}

	// Handle keyboard button press
	if update.Message.Text == "🌤️ Get Weather Data" {
		// Fetch the sensor data from localhost
		data, err := getSensorData()
		if err != nil {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Error fetching sensor data: %v", err)))
			return
		}

		// Create the response message
		sensorInfo := fmt.Sprintf("🌡️ Temperature: %.2f°C\n💧 Humidity: %.2f%%\n🌪️ Pressure: %.2fmmHg",
			data.Temperature, data.Humidity, data.Pressure*0.75)

		// Send the data to the user
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, sensorInfo))
	}
}

func main() {
	// Load bot token from environment
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN not set")
	}

	// Load authorized user IDs from environment
	authorizedUsersStr := os.Getenv("TELEGRAM_AUTHORIZED_USERS")
	authorizedUsers = make(map[int64]struct{})
	for idStr := range strings.SplitSeq(authorizedUsersStr, ",") {
		if len(idStr) == 0 {
			continue
		}
		idStr = strings.TrimSpace(idStr)
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Fatalf("Invalid user ID in TELEGRAM_AUTHORIZED_USERS: %s", idStr)
		}
		authorizedUsers[id] = struct{}{}
	}

	// Create a new bot instance with the bot token
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	// Set up the updates channel
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	log.Println("App started successfully")

	// Handle updates (messages and button presses)
	for update := range updates {
		if update.Message != nil {
			handleMessage(bot, update)
		}
	}
}
