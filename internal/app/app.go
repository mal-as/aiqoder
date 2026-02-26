package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/firebase/genkit/go/core/api"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/ollama"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/textsplitter"

	"github.com/mal-as/aiqoder/internal/config"
	"github.com/mal-as/aiqoder/internal/flows"
	"github.com/mal-as/aiqoder/internal/infrastructure/gogit"
	"github.com/mal-as/aiqoder/internal/infrastructure/repository/repos"
	"github.com/mal-as/aiqoder/internal/services/documents"
	"github.com/mal-as/aiqoder/internal/services/retriever"
	"github.com/mal-as/aiqoder/internal/services/scanner"
	httpserver "github.com/mal-as/aiqoder/pkg/http_server"
	"github.com/mal-as/aiqoder/pkg/http_server/middleware/logging"
	"github.com/mal-as/aiqoder/pkg/logger"
	"github.com/mal-as/aiqoder/pkg/pg"
	"github.com/mal-as/aiqoder/pkg/pg/transaction"
)

type App struct {
	cfg    *config.Config
	logger *slog.Logger
	srv    *httpserver.Server
	fm     *flows.Manager

	repos struct {
		gitStorage *repos.Repository
	}
	services struct {
		doc  *documents.Service
		scan *scanner.Service
	}
	flows struct {
		index api.Action
		query api.Action
	}

	closeFuncs []func()
}

func (a *App) AddCloseFunc(f func()) {
	a.closeFuncs = append(a.closeFuncs, f)
}

func (a *App) close() {
	for i := len(a.closeFuncs) - 1; i >= 0; i-- {
		f := a.closeFuncs[i]
		if f != nil {
			f()
		}
	}
}

func Run() {
	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	app, err := initApp(ctx)
	if err != nil {
		slog.Default().Error(err.Error())
		return
	}
	defer app.close()

	go app.srv.Start()

	<-ctx.Done()
}

func initApp(ctx context.Context) (*App, error) {
	app := &App{}

	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("config init error: %w", err)
	}

	app.cfg = cfg
	app.logger = logger.New(cfg.Log)

	pool, err := pg.Connect(ctx,
		pg.WithHostPort(cfg.PG.Host),
		pg.WithUser(cfg.PG.User),
		pg.WithPassword(cfg.PG.Password),
		pg.WithUserAdmin(cfg.PG.UserAdmin),
		pg.WithPasswordAdmin(cfg.PG.PasswordAdmin),
		pg.WithDatabase(cfg.PG.Database),
		pg.WithMigrationsDir("migrations"),
		pg.WithLogger(app.logger),
	)
	if err != nil {
		return nil, fmt.Errorf("init pg error: %w", err)
	}
	app.AddCloseFunc(func() {
		pool.Close()
	})

	tx := transaction.New(pool, app.logger)
	app.repos.gitStorage = repos.NewRepository(tx)

	ollamaPlugin := &ollama.Ollama{ServerAddress: cfg.Ollama.ServerAddress}
	g := genkit.Init(ctx,
		genkit.WithPlugins(ollamaPlugin),
		genkit.WithPromptDir("prompts"),
	)

	model := ollamaPlugin.DefineModel(g, ollama.ModelDefinition{Name: cfg.Ollama.GenerativeModel}, nil)
	embedder := ollamaPlugin.DefineEmbedder(g, cfg.Ollama.ServerAddress, cfg.Ollama.EmbeddingModel, nil)
	ret := retriever.Define(g, app.repos.gitStorage, embedder)
	prompt := genkit.LookupPrompt(g, "query_repo")
	if prompt == nil {
		return nil, fmt.Errorf("prompt 'query_repo' not found — ensure prompts/query_repo.prompt exists")
	}

	cloner := gogit.New()

	app.initServices()
	app.fm = flows.NewManager(g, app.repos.gitStorage, tx, cloner, app.services.scan, embedder, model, ret, prompt)
	app.flows.index = app.fm.DefineIndexFlow()
	app.flows.query = app.fm.DefineQueryFlow()
	app.initHTTPServer() //nolint:contextcheck // gin router setup does not require a context

	return app, nil
}

func (a *App) initServices() {
	a.services.doc = documents.NewService(textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(a.cfg.Splitter.ChunkSize),
		textsplitter.WithChunkOverlap(a.cfg.Splitter.ChunkOverlap),
	))
	a.services.scan = scanner.New(a.services.doc)
}

func (a *App) initHTTPServer() {
	a.srv = httpserver.NewServer(
		httpserver.WithListen(a.cfg.HTTP.Listen),
		httpserver.WithReadTimeout(a.cfg.HTTP.ReadTimeout),
		httpserver.WithWriteTimeout(a.cfg.HTTP.WriteTimeout),
		httpserver.WithIdleTimeout(a.cfg.HTTP.IdleTimeout),
		httpserver.WithGracefulShutdown(a.cfg.HTTP.GracefulShutdown),
		httpserver.WithDebugGin(a.cfg.HTTP.Debug),
		httpserver.WithLogger(a.logger),
	)
	a.AddCloseFunc(func() {
		a.srv.Close()
	})

	a.srv.Router().Use(logging.LoggerMiddleware(a.logger))

	apiRouter := a.srv.Router().Group("/api/v1")
	flowsRouter := apiRouter.Group("/flows")

	flowsRouter.POST("/indexRepository", gin.WrapF(genkit.Handler(a.flows.index)))
	flowsRouter.POST("/queryRepository", gin.WrapF(genkit.Handler(a.flows.query)))
}
