package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPAddr                  string
	APIToken                  string
	MySQLDSN                  string
	GoDaddyBaseURL            string
	GoDaddyAPIKey             string
	GoDaddyAPISecret          string
	DomainSyncIntervalSeconds int
	GoDaddyTimeoutSeconds     int
	GoDaddyRateLimitPerMinute int
	RandomSubdomainLength     int
}

func Load(path string) (*Config, error) {
	_ = loadEnvFile(path)
	cfg := &Config{
		HTTPAddr:                  getenv("HTTP_ADDR", ":8080"),
		APIToken:                  getenv("API_TOKEN", ""),
		MySQLDSN:                  getenv("MYSQL_DSN", "root:password@tcp(127.0.0.1:3306)/adnx_dns?parseTime=true&charset=utf8mb4&loc=Local"),
		GoDaddyBaseURL:            getenv("GODADDY_BASE_URL", "https://api.godaddy.com"),
		GoDaddyAPIKey:             getenv("GODADDY_API_KEY", ""),
		GoDaddyAPISecret:          getenv("GODADDY_API_SECRET", ""),
		DomainSyncIntervalSeconds: atoi(getenv("DOMAIN_SYNC_INTERVAL_SECONDS", "300"), 300),
		GoDaddyTimeoutSeconds:     atoi(getenv("GODADDY_REQUEST_TIMEOUT_SECONDS", "15"), 15),
		GoDaddyRateLimitPerMinute: atoi(getenv("GODADDY_RATE_LIMIT_PER_MINUTE", "60"), 60),
		RandomSubdomainLength:     atoi(getenv("RANDOM_SUBDOMAIN_LENGTH", "8"), 8),
	}
	return cfg, nil
}

func getenv(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}

func atoi(s string, def int) int {
	v, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || v <= 0 {
		return def
	}
	return v
}

func loadEnvFile(path string) error {
	if path == "" {
		path = ".env"
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, "\"")
		if _, ok := os.LookupEnv(key); !ok {
			_ = os.Setenv(key, val)
		}
	}
	return s.Err()
}
