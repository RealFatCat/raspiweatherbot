package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
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
func createReplyKeyboard() *models.ReplyKeyboardMarkup {
	return &models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{
				{Text: "🌤️ Get Weather Data"},
			},
		},
		ResizeKeyboard: true,
	}
}

// Handle the received command
func handleMessage(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	// Check if user is authorized
	if _, authorized := authorizedUsers[update.Message.From.ID]; !authorized {
		log.Printf("Unauthorized start attempt by user ID: %d", update.Message.From.ID)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "You are not authorized to use this bot.",
		})
		return
	}

	// Handle start command
	if update.Message.Text == "/start" {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "Welcome! Use the keyboard below to get weather data.",
			ReplyMarkup: createReplyKeyboard(),
		})
	}

	// Handle keyboard button press
	if update.Message.Text == "🌤️ Get Weather Data" {
		// Fetch the sensor data from localhost
		data, err := getSensorData()
		if err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   fmt.Sprintf("Error fetching sensor data: %v", err),
			})
			return
		}

		// Create the response message
		sensorInfo := fmt.Sprintf("🌡️ Temperature: %.2f°C\n💧 Humidity: %.2f%%\n🌪️ Pressure: %.2fmmHg",
			data.Temperature, data.Humidity, data.Pressure*0.75)

		// Send the data to the user
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   sensorInfo,
		})
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

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
	opts := []bot.Option{
		bot.WithDefaultHandler(handleMessage),
	}

	b, err := bot.New(botToken, opts...)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("App started successfully")

	// Start the bot
	b.Start(ctx)
}
