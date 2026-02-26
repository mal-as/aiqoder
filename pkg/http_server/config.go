package httpserver

import (
	"log/slog"
	"time"
)

type config struct {
	listen            string
	readTimeout       time.Duration
	readHeaderTimeout time.Duration
	writeTimeout      time.Duration
	idleTimeout       time.Duration
	maxHeaderBytes    int
	gracefulShutdown  time.Duration
	debugGin          bool
	removeExtraSlash  bool
	logger            *slog.Logger
}

func defaultConfig() *config {
	return &config{
		listen:            ":8080",
		readTimeout:       time.Second * 15,
		readHeaderTimeout: time.Second * 15,
		writeTimeout:      time.Second * 15,
		idleTimeout:       time.Second * 60,
		maxHeaderBytes:    1024 * 1024,
		gracefulShutdown:  time.Second * 30,
		debugGin:          false,
		removeExtraSlash:  true,
		logger:            slog.Default(),
	}
}

type OptionFunc func(*config)

func WithListen(listen string) OptionFunc {
	return func(o *config) {
		o.listen = listen
	}
}

func WithReadTimeout(readTimeout time.Duration) OptionFunc {
	return func(o *config) {
		o.readTimeout = readTimeout
	}
}

func WithReadHeaderTimeout(readHeaderTimeout time.Duration) OptionFunc {
	return func(o *config) {
		o.readHeaderTimeout = readHeaderTimeout
	}
}

func WithWriteTimeout(writeTimeout time.Duration) OptionFunc {
	return func(o *config) {
		o.writeTimeout = writeTimeout
	}
}

func WithIdleTimeout(idleTimeout time.Duration) OptionFunc {
	return func(o *config) {
		o.idleTimeout = idleTimeout
	}
}

func WithMaxHeaderBytes(maxHeaderBytes int) OptionFunc {
	return func(o *config) {
		o.maxHeaderBytes = maxHeaderBytes
	}
}

func WithGracefulShutdown(gracefulShutdown time.Duration) OptionFunc {
	return func(o *config) {
		o.gracefulShutdown = gracefulShutdown
	}
}

func WithDebugGin(debugGin bool) OptionFunc {
	return func(o *config) {
		o.debugGin = debugGin
	}
}

func WithRemoveExtraSlash(removeExtraSlash bool) OptionFunc {
	return func(o *config) {
		o.removeExtraSlash = removeExtraSlash
	}
}

func WithLogger(logger *slog.Logger) OptionFunc {
	return func(o *config) {
		o.logger = logger
	}
}
