package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Database struct {
		URL string
	}
	API struct {
		Port int
	}
	SMTP struct {
		Host     string
		Port     string
		Username string
		Password string
		From     string
	}
	Temporal struct {
		Host string
		Port int
	}
	Redis struct {
		URL string
	}
	GitHub struct {
		ClientID     string
		ClientSecret string
	}
	Google struct {
		ClientID     string
		ClientSecret string
	}
	GoogleConnect struct {
		ClientID     string
		ClientSecret string
	}
	Notion struct {
		ClientID     string
		ClientSecret string
	}
	Discord struct {
		ClientID     string
		ClientSecret string
	}

	Slack struct {
		ClientID     string
		ClientSecret string
	}
	OAuth struct {
		RedirectURL                  string
		GoogleIntegrationRedirectURL string
		GitHubIntegrationRedirectURL string
		NotionIntegrationRedirectURL string
	}
	JWT struct {
		Secret string
	}
	EncryptionKey   string
	Environment     string
	ExcelOutputPath string
}

var cfg *Config

func Load() (*Config, error) {
	godotenv.Load()

	cfg = &Config{
		Database: struct {
			URL string
		}{
			URL: getEnv("DATABASE_URL", ""),
		},
		API: struct {
			Port int
		}{
			Port: getEnvInt("API_PORT", 8080),
		},
		Temporal: struct {
			Host string
			Port int
		}{
			Host: getEnv("TEMPORAL_HOST", "localhost"),
			Port: getEnvInt("TEMPORAL_PORT", 7233),
		},
		Redis: struct {
			URL string
		}{
			URL: getEnv("REDIS_URL", "redis://localhost:6379"),
		},
		GitHub: struct {
			ClientID     string
			ClientSecret string
		}{
			ClientID:     getEnv("GITHUB_CLIENT_ID", ""),
			ClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		},
		Google: struct {
			ClientID     string
			ClientSecret string
		}{
			ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		},
		GoogleConnect: struct {
			ClientID     string
			ClientSecret string
		}{
			ClientID:     getEnv("GOOGLE_CONNECT_CLIENT_ID", ""),
			ClientSecret: getEnv("GOOGLE_CONNECT_CLIENT_SECRET", ""),
		},
		Notion: struct {
			ClientID     string
			ClientSecret string
		}{
			ClientID:     getEnv("NOTION_CLIENT_ID", ""),
			ClientSecret: getEnv("NOTION_CLIENT_SECRET", ""),
		},
		Discord: struct {
			ClientID     string
			ClientSecret string
		}{
			ClientID:     getEnv("DISCORD_CLIENT_ID", ""),
			ClientSecret: getEnv("DISCORD_CLIENT_SECRET", ""),
		},
		Slack: struct {
			ClientID     string
			ClientSecret string
		}{
			ClientID:     getEnv("SLACK_CLIENT_ID", ""),
			ClientSecret: getEnv("SLACK_CLIENT_SECRET", ""),
		},
		OAuth: struct {
			RedirectURL                  string
			GoogleIntegrationRedirectURL string
			GitHubIntegrationRedirectURL string
			NotionIntegrationRedirectURL string
		}{
			RedirectURL:                  getEnv("OAUTH_REDIRECT_URL", ""),
			GoogleIntegrationRedirectURL: getEnv("GOOGLE_INTEGRATION_REDIRECT_URL", ""),
			GitHubIntegrationRedirectURL: getEnv("GITHUB_INTEGRATION_REDIRECT_URL", ""),
			NotionIntegrationRedirectURL: getEnv("NOTION_INTEGRATION_REDIRECT_URL", ""),
		},
		JWT: struct {
			Secret string
		}{
			Secret: getEnv("JWT_SECRET", "dev-secret-key"),
		},
		SMTP: struct {
			Host     string
			Port     string
			Username string
			Password string
			From     string
		}{
			Host:     getEnv("SMTP_HOST", ""),
			Port:     getEnv("SMTP_PORT", ""),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", ""),
		},
		EncryptionKey:   getEnv("ENCRYPTION_KEY", ""),
		Environment:     getEnv("ENV", "development"),
		ExcelOutputPath: getEnv("EXCEL_OUTPUT_PATH", "./output"),
	}

	return cfg, nil
}
func Get() *Config {
	return cfg
}
func getEnv(key, defaultVal string) string {
	if value, exist := os.LookupEnv(key); exist {
		return value
	}
	return defaultVal
}
func getEnvInt(key string, defaultVal int) int {
	val := getEnv(key, "")
	if val == "" {
		return defaultVal
	}
	IntVal, _ := strconv.Atoi(val)
	return IntVal
}
