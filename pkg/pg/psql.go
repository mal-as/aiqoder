package pg

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
	"github.com/pressly/goose/v3/lock"
)

// Connect возвращает настроенный *pgxpool.Pool к БД
//
// При инициализации *pgxpool.Pool конфигурируются на основе переменных окружения
//   - PG_HOST
//   - PG_USER
//   - PG_PASSWORD
//   - PG_USER_ADMIN
//   - PG_PASSWORD_ADMIN
//   - PG_DATABASE
//
// Так же эти значения можно переопределить с помощью OptionFunc
func Connect(ctx context.Context, opts ...OptionFunc) (*pgxpool.Pool, error) {
	options := initOptions()

	for _, opt := range opts {
		opt(&options)
	}

	connString := (&url.URL{
		Scheme: "postgres",
		User:   getUserInfo(options.user, options.password),
		Host:   fmt.Sprintf("%s:%d", options.host, options.port),
		Path:   options.database,
	}).String()

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.ParseConfig: %w", err)
	}

	poolConfig.MinConns = options.maxIdleConns
	poolConfig.MaxConns = options.maxOpenConns
	poolConfig.MaxConnIdleTime = options.connMaxIdleTime
	poolConfig.MaxConnLifetime = options.connMaxLifetime
	poolConfig.ConnConfig.DefaultQueryExecMode = options.queryExecMode

	var db *pgxpool.Pool
	db, err = pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.NewWithConfig: %w", err)
	}

	if options.connectionWaiting.enabled {
		if err = options.connectionWaiting.pingDB(ctx, db); err != nil {
			return nil, fmt.Errorf("failed to ping database: %w", err)
		}
	}

	if options.migrations != nil {
		migrationConnString := (&url.URL{
			Scheme: "postgres",
			User:   getUserInfo(options.userAdmin, options.passwordAdmin),
			Host:   fmt.Sprintf("%s:%d", options.host, options.port),
			Path:   options.database,
		}).String()

		config, err := pgx.ParseConfig(migrationConnString)
		if err != nil {
			return nil, fmt.Errorf("pgx.ParseConfig: %w", err)
		}

		stdDb := stdlib.OpenDB(*config)
		stdDb.SetMaxIdleConns(0)
		defer func() { _ = stdDb.Close() }()

		if err = stdDb.Ping(); err != nil {
			return nil, fmt.Errorf("stdDb.Ping: %w", err)
		}

		locker, err := lock.NewPostgresSessionLocker()
		if err != nil {
			return nil, fmt.Errorf("lock.NewPostgresSessionLocker: %w", err)
		}

		g, err := goose.NewProvider(
			database.DialectPostgres,
			stdDb,
			options.migrations,
			goose.WithSessionLocker(locker), // нужно чтобы несколько инстансов не катили миграции одновременно
			goose.WithAllowOutofOrder(true), // не валиться если миграция скомиченная раньше катится позже последней
		)
		if err != nil {
			if errors.Is(err, goose.ErrNoMigrations) {
				if options.logger != nil {
					options.logger.WarnContext(ctx, "no postgres migrations found")
				}
				return db, nil
			}

			return nil, fmt.Errorf("goose.NewProvider: %w", err)
		}
		defer func() {
			_ = g.Close()
		}()

		if options.logger != nil {
			var version int64
			version, err = g.GetDBVersion(ctx)
			if err != nil {
				return nil, fmt.Errorf("g.GetDBVersion: %w", err)
			}

			options.logger.InfoContext(ctx, fmt.Sprintf("db migrations version %d", version))
		}

		res, err := g.Up(ctx)
		if err != nil {
			return nil, fmt.Errorf("g.Up: %w", err)
		}

		if options.logger != nil {
			for _, i := range res {
				options.logger.InfoContext(ctx, "migration "+i.Source.Path,
					slog.String("file", i.Source.Path),
					slog.String("duration", i.Duration.String()),
				)
			}
		}
	}

	return db, nil
}

func getUserInfo(user, password string) *url.Userinfo {
	if user != "" {
		if password != "" {
			return url.UserPassword(user, password)
		}

		return url.User(user)
	}

	return nil
}
