package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nzb3/slogmanager"
	"github.com/tmc/langchaingo/llms/ollama"
	"gorm.io/gorm"

	"github.com/nzb3/diploma/resource-service/internal/controllers"
	"github.com/nzb3/diploma/resource-service/internal/controllers/middleware"
	"github.com/nzb3/diploma/resource-service/internal/controllers/resourcecontroller"
	"github.com/nzb3/diploma/resource-service/internal/domain/services/contentextractor"
	"github.com/nzb3/diploma/resource-service/internal/domain/services/resourceservcie"
	"github.com/nzb3/diploma/resource-service/internal/repository/pgx"
	"github.com/nzb3/diploma/resource-service/internal/repository/pgx/events"
	"github.com/nzb3/diploma/resource-service/internal/repository/pgx/resources"
	"github.com/nzb3/diploma/resource-service/internal/server"
)

// ServiceProvider implementation of DI-container haves method to initialize components of application
type ServiceProvider struct {
	slogManager         *slogmanager.Manager
	embeddingLLM        *ollama.LLM
	generationLLM       *ollama.LLM
	server              *http.Server
	resourceController  *resourcecontroller.Controller
	ginEngine           *gin.Engine
	resourceService     *resourceservcie.Service
	serverConfig        *server.Config
	repositoryConfig    *pgx.Config
	pgxPool             *pgxpool.Pool
	repository          *pgx.Repository
	resourcesRepository *resources.Repository
	eventsRepository    *events.Repository
	gormDB              *gorm.DB
	contentExtractor    *contentextractor.ContentExtractor
	authConfig          *middleware.AuthMiddlewareConfig
	authMiddleware      *middleware.AuthMiddleware
}

// NewServiceProvider creates and returns a new instance of ServiceProvider
func NewServiceProvider() *ServiceProvider {
	return &ServiceProvider{}
}

// Logger returns the application's slog manager, creating it if it doesn't exist
func (sp *ServiceProvider) Logger(ctx context.Context) *slogmanager.Manager {
	if sp.slogManager != nil {
		return sp.slogManager
	}
	manager := slogmanager.New()
	manager.AddWriter("stdout", slogmanager.NewWriter(os.Stdout, slogmanager.WithTextFormat()))
	slog.SetLogLoggerLevel(slog.LevelDebug)
	sp.slogManager = manager
	return sp.slogManager
}

// EmbeddingLLM returns the LLM instance for embeddings, creating it if it doesn't exist
func (sp *ServiceProvider) EmbeddingLLM(ctx context.Context) *ollama.LLM {
	if sp.embeddingLLM != nil {
		return sp.embeddingLLM
	}

	llm, err := ollama.New(
		ollama.WithServerURL("http://ollama-embedder:11434/"),
		ollama.WithModel("bge-m3"),
	)
	if err != nil {
		sp.Logger(ctx).Logger().Error("error creating ollama embedding LLM", "error", err.Error())
		panic(fmt.Errorf("error creating ollama embedding LLM: %w", err))
	}

	sp.embeddingLLM = llm

	return llm
}

// GeneratingLLM returns the LLM instance for text generation, creating it if it doesn't exist
func (sp *ServiceProvider) GeneratingLLM(ctx context.Context) *ollama.LLM {
	if sp.generationLLM != nil {
		return sp.generationLLM
	}

	llm, err := ollama.New(ollama.WithServerURL("http://ollama-generator:11434/"),
		ollama.WithModel("gemma3:4b-it-qat"),
	)
	if err != nil {
		sp.Logger(ctx).Logger().Error("error creating ollama generating LLM", "error", err.Error())
		panic(fmt.Errorf("error creating ollama generating LLM: %w", err))
	}

	sp.generationLLM = llm
	return llm
}

// AuthConfig returns the auth configuration, creating it if it doesn't exist
func (sp *ServiceProvider) AuthConfig(ctx context.Context) *middleware.AuthMiddlewareConfig {
	if sp.authConfig != nil {
		return sp.authConfig
	}

	host := os.Getenv("AUTH_HOST")
	if host == "" {
		slog.Error("AUTH_HOST environment variable not set")
		return nil
	}
	port := os.Getenv("AUTH_PORT")
	if port == "" {
		slog.Error("AUTH_PORT environment variable not set")
		return nil
	}
	realm := os.Getenv("AUTH_REALM")
	if realm == "" {
		slog.Error("AUTH_REALM environment variable not set")
		return nil
	}
	clientID := os.Getenv("AUTH_CLIENT_ID")
	if clientID == "" {
		slog.Error("AUTH_CLIENT_ID environment variable not set")
		return nil
	}
	clientSecret := os.Getenv("AUTH_CLIENT_SECRET")
	if clientSecret == "" {
		slog.Error("AUTH_CLIENT_SECRET environment variable not set")
		return nil
	}

	sp.authConfig = &middleware.AuthMiddlewareConfig{
		Host:         host,
		Port:         port,
		Realm:        realm,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}

	return sp.authConfig
}

// GinEngine returns the configured Gin web engine instance, creating it if it doesn't exist
func (sp *ServiceProvider) GinEngine(ctx context.Context) *gin.Engine {
	if sp.ginEngine != nil {
		return sp.ginEngine
	}
	_ = ctx
	engine := gin.Default()

	corsConfig := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	corsConfig.AllowAllOrigins = true

	engine.Use(cors.New(corsConfig))

	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	engine = sp.setupRoutes(
		ctx,
		engine,
		sp.ResourceController(ctx),
	)

	sp.ginEngine = engine
	return engine
}

func (sp *ServiceProvider) setupRoutes(ctx context.Context, router *gin.Engine, controllers ...controllers.Controller) *gin.Engine {
	api := router.Group("/api")
	v1 := api.Group("/v1")

	v1.Use(sp.AuthMiddleware(ctx).Authenticate())

	for _, controller := range controllers {
		controller.RegisterRoutes(v1)
	}

	return router
}

func (sp *ServiceProvider) AuthMiddleware(ctx context.Context) *middleware.AuthMiddleware {
	if sp.authMiddleware != nil {
		return sp.authMiddleware
	}

	_ = ctx

	authMiddleware := middleware.NewAuthMiddleware(sp.AuthConfig(ctx))

	sp.authMiddleware = authMiddleware
	return authMiddleware
}

// RepositoryConfig returns the repository configuration, creating it if it doesn't exist
func (sp *ServiceProvider) RepositoryConfig(ctx context.Context) *pgx.Config {
	if sp.repositoryConfig != nil {
		return sp.repositoryConfig
	}

	config := pgx.NewConfig()

	sp.repositoryConfig = config

	return config
}

// PgxPool returns the pgx connection pool, creating it if it doesn't exist
func (sp *ServiceProvider) PgxPool(ctx context.Context) *pgxpool.Pool {
	if sp.pgxPool != nil {
		return sp.pgxPool
	}

	pool, err := pgx.NewPgxPool(ctx, sp.RepositoryConfig(ctx))
	if err != nil {
		sp.Logger(ctx).Logger().Error("error creating pgx pool", "error", err.Error())
		panic(fmt.Errorf("error creating pgx pool: %w", err))
	}

	sp.pgxPool = pool
	return pool
}

// Repository returns the pgx repository instance, creating it if it doesn't exist
func (sp *ServiceProvider) Repository(ctx context.Context) *pgx.Repository {
	if sp.repository != nil {
		return sp.repository
	}

	repository, err := pgx.NewRepository(ctx, sp.PgxPool(ctx))
	if err != nil {
		sp.Logger(ctx).Logger().Error("error creating repository", "error", err.Error())
		panic(fmt.Errorf("error creating repository: %w", err))
	}

	sp.repository = repository

	return repository
}

// ResourcesRepository returns the resources repository instance, creating it if it doesn't exist
func (sp *ServiceProvider) ResourcesRepository(ctx context.Context) *resources.Repository {
	if sp.resourcesRepository != nil {
		return sp.resourcesRepository
	}

	resourcesRepository := resources.NewResourceRepository(ctx, sp.Repository(ctx))

	sp.resourcesRepository = resourcesRepository
	return resourcesRepository
}

// EventsRepository returns the events repository instance, creating it if it doesn't exist
func (sp *ServiceProvider) EventsRepository(ctx context.Context) *events.Repository {
	if sp.eventsRepository != nil {
		return sp.eventsRepository
	}

	eventsRepository := events.NewEventRepository(ctx, sp.Repository(ctx))

	sp.eventsRepository = eventsRepository
	return eventsRepository
}

// ResourceProcessor returns the resource processor instance, creating it if it doesn't exist
func (sp *ServiceProvider) ResourceProcessor(ctx context.Context) *contentextractor.ContentExtractor {
	if sp.contentExtractor != nil {
		return sp.contentExtractor
	}

	resourceProcessor := contentextractor.NewResourceProcessor()

	sp.contentExtractor = resourceProcessor

	return resourceProcessor
}

// ResourceService returns the resource service instance, creating it if it doesn't exist
func (sp *ServiceProvider) ResourceService(ctx context.Context) *resourceservcie.Service {
	if sp.resourceService != nil {
		return sp.resourceService
	}

	service := resourceservcie.NewService(sp.Repository(ctx), sp.ResourceProcessor(ctx))

	sp.resourceService = service

	return service
}

// ResourceController returns the resource controller instance, creating it if it doesn't exist
func (sp *ServiceProvider) ResourceController(ctx context.Context) *resourcecontroller.Controller {
	if sp.resourceController != nil {
		return sp.resourceController
	}

	controller := resourcecontroller.NewController(sp.ResourceService(ctx))

	sp.resourceController = controller

	return controller
}

// ServerConfig returns the server configuration, creating it if it doesn't exist
func (sp *ServiceProvider) ServerConfig(ctx context.Context) *server.Config {
	if sp.serverConfig != nil {
		return sp.serverConfig
	}

	config, err := server.NewConfig()
	if err != nil {
		sp.Logger(ctx).Logger().Error("error creating server config", "error", err.Error())
		panic(fmt.Errorf("error creating server config: %w", err))
	}

	sp.serverConfig = config
	return config
}

// Server returns the HTTP server instance, creating it if it doesn't exist
func (sp *ServiceProvider) Server(ctx context.Context) *http.Server {
	if sp.server != nil {
		return sp.server
	}

	s := server.NewServer(
		ctx,
		sp.GinEngine(ctx),
		sp.ServerConfig(ctx),
	)

	sp.server = s
	return s
}
