package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HTTP     HTTP
	Ollama   Ollama
	PG       PG
	Log      Log
	Splitter Splitter
}

type HTTP struct {
	Listen           string        `env:"HTTP_LISTEN"            env-default:"localhost:8001"`
	ReadTimeout      time.Duration `env:"HTTP_READ_TIMEOUT"      env-default:"60s"`
	WriteTimeout     time.Duration `env:"HTTP_WRITE_TIMEOUT"     env-default:"60s"`
	IdleTimeout      time.Duration `env:"HTTP_IDLE_TIMEOUT"      env-default:"60s"`
	GracefulShutdown time.Duration `env:"HTTP_GRACEFUL_SHUTDOWN" env-default:"30s"`
	Debug            bool          `env:"HTTP_DEBUG"             env-default:"false"`
}

type Ollama struct {
	ServerAddress   string `env:"OLLAMA_SERVER_ADDRESS"   env-default:"http://127.0.0.1:11434"`
	EmbeddingModel  string `env:"OLLAMA_EMBEDDING_MODEL"  env-default:"all-minilm:33m"`
	GenerativeModel string `env:"OLLAMA_GENERATIVE_MODEL" env-default:"qwen3-coder:480b-cloud"`
}

type PG struct {
	Host          string `env:"PG_HOST"           env-default:"localhost:5432"`
	User          string `env:"PG_USER"           env-required:"true"`
	Password      string `env:"PG_PASSWORD"       env-required:"true"`
	UserAdmin     string `env:"PG_USER_ADMIN"`
	PasswordAdmin string `env:"PG_PASSWORD_ADMIN"`
	Database      string `env:"PG_DATABASE"       env-required:"true"`
}

type Log struct {
	Level string `env:"LOG_LEVEL" env-default:"info"`
}

type Splitter struct {
	ChunkSize    int `env:"SPLITTER_CHUNK_SIZE" env-default:"200"`
	ChunkOverlap int `env:"SPLITTER_CHUNK_OVERLAP" env-default:"20"`
}

func Load() (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, fmt.Errorf("cleanenv.ReadEnv: %w", err)
	}

	return cfg, nil
}
