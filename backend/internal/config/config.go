package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port            string
	AppEnv          string
	CORSOrigins     []string
	DBDriver        string
	DatabaseURL     string
	JWTSecret       string
	JWTTTL          time.Duration
	BcryptCost      int
	LoginRateLimit  int
	LoginRateWindow time.Duration
	Timezone        string
	SeedAdminEmail  string
	SeedAdminPass   string
	SeedWSName      string
	AttachmentsDir        string
	PlatformAdminEmails []string
	// PublicAPIBase is the externally reachable API root for join links/prompts (ONE_TAP_JOIN).
	// Env PUBLIC_API_BASE; default https://bot.api.sn.cn
	PublicAPIBase string
	// PublicBoardBase optional FE origin for /join deep-link hints (PUBLIC_BOARD_BASE).
	PublicBoardBase string
}

func Load() Config {
	ttlH, _ := strconv.Atoi(getenv("JWT_TTL_HOURS", "168"))
	cost, _ := strconv.Atoi(getenv("BCRYPT_COST", "12"))
	rl, _ := strconv.Atoi(getenv("LOGIN_RATE_LIMIT", "10"))
	rlw, _ := strconv.Atoi(getenv("LOGIN_RATE_WINDOW_SEC", "60"))
	origins := strings.Split(getenv("CORS_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173"), ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}
	return Config{
		Port:            getenv("PORT", "8080"),
		AppEnv:          getenv("APP_ENV", "development"),
		CORSOrigins:     origins,
		DBDriver:        getenv("DB_DRIVER", "sqlite"),
		DatabaseURL:     getenv("DATABASE_URL", "./data/ai-collab.db"),
		JWTSecret:       getenv("JWT_SECRET", "dev-secret-change-me-ai-collab-board"),
		JWTTTL:          time.Duration(ttlH) * time.Hour,
		BcryptCost:      cost,
		LoginRateLimit:  rl,
		LoginRateWindow: time.Duration(rlw) * time.Second,
		Timezone:        getenv("TZ", "Asia/Shanghai"),
		SeedAdminEmail:  getenv("SEED_ADMIN_EMAIL", "admin@yumeng.tech"),
		SeedAdminPass:   getenv("SEED_ADMIN_PASSWORD", "Admin123!"),
		SeedWSName:      getenv("SEED_WORKSPACE_NAME", "育蒙科技"),
		AttachmentsDir:        getenv("ATTACHMENTS_DIR", "./data/attachments"),
		PlatformAdminEmails: parseCSVEmails(getenv("PLATFORM_ADMIN_EMAILS", "")),
		PublicAPIBase:       strings.TrimRight(getenv("PUBLIC_API_BASE", "https://bot.api.sn.cn"), "/"),
		PublicBoardBase:     strings.TrimRight(getenv("PUBLIC_BOARD_BASE", ""), "/"),
	}
}

func parseCSVEmails(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		e := strings.ToLower(strings.TrimSpace(part))
		if e != "" {
			out = append(out, e)
		}
	}
	return out
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
