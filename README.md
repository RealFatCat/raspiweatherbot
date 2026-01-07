# Raspberry Pi Weather Bot

Telegram bot that sends weather sensor data from Raspberry Pi to authorized users.

It was written to use with [RaspiWeather project](https://github.com/RealFatCat/raspiweather).

## Install

```bash
git clone https://github.com/realfatcat/raspiweatherbot.git
cd raspiweatherbot
# To build for current architecture
make build  
```

To build for raspberry pi, instead of `make build`, run make for your raspberry pi architecture, `make arm6` for example.

Or just run `make build-all` to build binaries for every arhitecture.

## Config

Set env vars:
- `TELEGRAM_BOT_TOKEN`
- `TELEGRAM_AUTHORIZED_USERS` (comma-separated IDs of allowed users)

Unautorized requests are logged with user IDs, so to determine your user id, just send `/start` to bot - your ID will be in the logs of app.

## Run

```bash
./raspiweatherbot
```

### Help

```
./raspiweatherbot -h
```

## Usage

Send `/start` in Telegram, click "🌤️ Get Weather Data".