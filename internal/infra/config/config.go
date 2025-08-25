package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env            string
	HTTPPort       string
	LogLevel       string
	RateLimitRPS   float64
	RateLimitBurst int
	RequestTimeout time.Duration

	StripeSecretKey     string
	StripeWebhookSecret string
	StripeEnableTestPM  bool
	StripeTestPaymentPM string // "pm_card_visa"

	CBMaxRequests uint32
	CBInterval    time.Duration
	CBTimeout     time.Duration
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		Env:            getEnv("APP_ENV", "dev"),
		HTTPPort:       getEnv("HTTP_PORT", "8080"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		RateLimitRPS:   getEnvFloat("RATE_LIMIT_RPS", 10),
		RateLimitBurst: getEnvInt("RATE_LIMIT_BURST", 20),
		RequestTimeout: getEnvDuration("REQUEST_TIMEOUT", 15*time.Second),

		StripeSecretKey:     getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),
		StripeEnableTestPM:  getEnv("STRIPE_ENABLE_TEST_PM", "true") == "true",
		StripeTestPaymentPM: getEnv("STRIPE_TEST_PAYMENT_METHOD", "pm_card_visa"),

		CBMaxRequests: uint32(getEnvInt("CB_MAX_REQUESTS", 3)),
		CBInterval:    getEnvDuration("CB_INTERVAL", 60*time.Second),
		CBTimeout:     getEnvDuration("CB_TIMEOUT", 8*time.Second),
	}
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}
func getEnvInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok {
		var x int
		_, err := fmt.Sscanf(v, "%d", &x)
		if err == nil {
			return x
		}
	}
	return def
}
func getEnvFloat(key string, def float64) float64 {
	if v, ok := os.LookupEnv(key); ok {
		var x float64
		_, err := fmt.Sscanf(v, "%f", &x)
		if err == nil {
			return x
		}
	}
	return def
}
func getEnvDuration(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
	}
	return def
}
