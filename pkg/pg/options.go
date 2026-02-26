package pg

import (
	"io/fs"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	pgx "github.com/jackc/pgx/v5"
)

// psqlOptions настройки подключения к базе
type psqlOptions struct {
	host              string
	port              uint16
	database          string
	user              string
	password          string
	userAdmin         string
	passwordAdmin     string
	connectionWaiting pingOptions
	queryExecMode     pgx.QueryExecMode
	maxOpenConns      int32
	maxIdleConns      int32
	connMaxLifetime   time.Duration
	connMaxIdleTime   time.Duration
	migrations        fs.FS
	logger            *slog.Logger
}

func initOptions() psqlOptions {
	host, port := parseHostPort("PG_HOST")
	if host == "" {
		host = "localhost"
	}
	if port == 0 {
		port = 5432
	}

	return psqlOptions{
		host:              host,
		port:              port,
		user:              os.Getenv("PG_USER"),
		password:          os.Getenv("PG_PASSWORD"),
		userAdmin:         os.Getenv("PG_USER_ADMIN"),
		passwordAdmin:     os.Getenv("PG_PASSWORD_ADMIN"),
		database:          os.Getenv("PG_DATABASE"),
		queryExecMode:     pgx.QueryExecModeExec,
		connectionWaiting: defaultPingOptions,
		maxOpenConns:      5,
		maxIdleConns:      5,
		connMaxLifetime:   5 * time.Minute,
		connMaxIdleTime:   5 * time.Minute,
	}
}

// OptionFunc - тип опции
type OptionFunc func(*psqlOptions)

// WithHost - устанавливает host для подключения к базе
func WithHost(host string) OptionFunc {
	return func(o *psqlOptions) {
		o.host = host
	}
}

// WithHostPort - устанавливает host и port из строки вида "localhost:5432"
func WithHostPort(hostPort string) OptionFunc {
	return func(o *psqlOptions) {
		o.host, o.port = parseHostPort(hostPort)
	}
}

// WithPort - устанавливает port для подключения к базе
func WithPort(port uint16) OptionFunc {
	return func(o *psqlOptions) {
		o.port = port
	}
}

// WithUser - устанавливает юзера для подключения к базе
func WithUser(user string) OptionFunc {
	return func(o *psqlOptions) {
		o.user = user
	}
}

func WithUserAdmin(user string) OptionFunc {
	return func(o *psqlOptions) {
		o.userAdmin = user
	}
}

// WithPassword - устанавливает пароль для подключения к базе
func WithPassword(password string) OptionFunc {
	return func(o *psqlOptions) {
		o.password = password
	}
}

func WithPasswordAdmin(password string) OptionFunc {
	return func(o *psqlOptions) {
		o.passwordAdmin = password
	}
}

// WithDatabase - устанавливает имя базы для подключения
func WithDatabase(database string) OptionFunc {
	return func(o *psqlOptions) {
		o.database = database
	}
}

// WithSimpleProtocol - использовать ли QueryExecModeSimpleProtocol
func WithSimpleProtocol(simple bool) OptionFunc {
	return func(o *psqlOptions) {
		if simple {
			o.queryExecMode = pgx.QueryExecModeSimpleProtocol
		}
	}
}

func WithoutConnectionWaiting() OptionFunc {
	return func(o *psqlOptions) {
		o.connectionWaiting.enabled = false
	}
}

// WithConnectionWaiting - опция, которая добавит ожидание коннекта к БД через пинг
// Доступна модификация параметров пинга функциями PingTick, PingDeadLine
func WithConnectionWaiting(opts ...OptionPingFunc) OptionFunc {
	return func(o *psqlOptions) {
		o.connectionWaiting = defaultPingOptions

		// применяем опции из переданных опций
		for _, opt := range opts {
			opt(o)
		}
	}
}

func WithMaxOpenConns(conns int32) OptionFunc {
	return func(options *psqlOptions) {
		options.maxOpenConns = conns
	}
}

func WithMaxIdleConns(conns int32) OptionFunc {
	return func(options *psqlOptions) {
		options.maxIdleConns = conns
	}
}

func WithConnMaxLifetime(d time.Duration) OptionFunc {
	return func(options *psqlOptions) {
		options.connMaxLifetime = d
	}
}

func WithConnMaxIdleTime(d time.Duration) OptionFunc {
	return func(options *psqlOptions) {
		options.connMaxIdleTime = d
	}
}

func WithMigrationsDir(dir string) OptionFunc {
	return func(options *psqlOptions) {
		options.migrations = os.DirFS(dir)
	}
}

func WithLogger(logger *slog.Logger) OptionFunc {
	return func(options *psqlOptions) {
		options.logger = logger
	}
}

func parseHostPort(hostPort string) (host string, port uint16) {
	s := strings.Split(hostPort, ":")
	if len(s) != 2 {
		return host, port
	}

	host = s[0]

	uintPort, err := strconv.ParseUint(s[1], 10, 16)
	if err != nil {
		return host, port
	}
	port = uint16(uintPort)

	return host, port
}
