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
	"github.com/nzb3/closer"
	"github.com/nzb3/slogmanager"
	"github.com/tmc/langchaingo/llms/ollama"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nzb3/diploma/search-service/internal/controllers"
	"github.com/nzb3/diploma/search-service/internal/controllers/middleware"
	"github.com/nzb3/diploma/search-service/internal/controllers/resourcecontroller"
	"github.com/nzb3/diploma/search-service/internal/controllers/searchcontroller"
	"github.com/nzb3/diploma/search-service/internal/domain/services/resourceservcie"
	"github.com/nzb3/diploma/search-service/internal/domain/services/searchservice"
	"github.com/nzb3/diploma/search-service/internal/repository/gormpg"
	"github.com/nzb3/diploma/search-service/internal/repository/integration/embedder"
	"github.com/nzb3/diploma/search-service/internal/repository/integration/generator"
	"github.com/nzb3/diploma/search-service/internal/repository/integration/resourceprocessor"
	"github.com/nzb3/diploma/search-service/internal/repository/vectorstorage"
	"github.com/nzb3/diploma/search-service/internal/server"
)

// ServiceProvider implementation of DI-container haves method to initialize components of application
type ServiceProvider struct {
	slogManager         *slogmanager.Manager
	embeddingLLM        *ollama.LLM
	generationLLM       *ollama.LLM
	embedder            *embedder.Embedder
	generator           *generator.Generator
	server              *http.Server
	resourceController  *resourcecontroller.Controller
	ginEngine           *gin.Engine
	vectorStore         *vectorstorage.VectorStorage
	vectorStorageConfig *vectorstorage.Config
	resourceService     *resourceservcie.Service
	serverConfig        *server.Config
	repositoryConfig    *gormpg.Config
	repository          *gormpg.Repository
	gormDB              *gorm.DB
	searchController    *searchcontroller.Controller
	searchService       *searchservice.Service
	resourceProcessor   *resourceprocessor.ResourceProcessor
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
	_ = ctx
	manager := slogmanager.New()
	manager.AddWriter("stdout", slogmanager.NewWriter(os.Stdout, slogmanager.WithTextFormat()))
	slog.SetLogLoggerLevel(slog.LevelDebug)
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

// Embedder returns the embedder service instance, creating it if it doesn't exist
func (sp *ServiceProvider) Embedder(ctx context.Context) *embedder.Embedder {
	if sp.embedder != nil {
		return sp.embedder
	}

	e, err := embedder.NewEmbedder(sp.EmbeddingLLM(ctx))
	if err != nil {
		sp.Logger(ctx).Logger().Error("error creating embedding LLM", "error", err.Error())
		panic(fmt.Errorf("error creating embedding LLM: %w", err))
	}

	sp.embedder = e

	return e
}

// Generator returns the text generator service instance, creating it if it doesn't exist
func (sp *ServiceProvider) Generator(ctx context.Context) *generator.Generator {
	if sp.generator != nil {
		return sp.generator
	}

	g, err := generator.NewGenerator(sp.GeneratingLLM(ctx))
	if err != nil {
		sp.Logger(ctx).Logger().Error("error creating generating LLM", "error", err.Error())
		panic(fmt.Errorf("error creating generating LLM: %w", err))
	}

	sp.generator = g

	return g
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
		sp.SearchController(ctx),
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

// VectorStorageConfig returns the vector storage configuration, creating it if it doesn't exist
func (sp *ServiceProvider) VectorStorageConfig(ctx context.Context) *vectorstorage.Config {
	if sp.vectorStorageConfig != nil {
		return sp.vectorStorageConfig
	}

	config, err := vectorstorage.NewConfig(sp.RepositoryConfig(ctx).DSN)
	if err != nil {
		sp.Logger(ctx).Logger().Error("error creating vector storage config", "error", err.Error())
		panic(fmt.Errorf("error creating vector storage config: %w", err))
	}

	sp.vectorStorageConfig = config

	return config
}

// VectorStore returns the vector storage instance, creating it if it doesn't exist
func (sp *ServiceProvider) VectorStore(ctx context.Context) *vectorstorage.VectorStorage {
	if sp.vectorStore != nil {
		return sp.vectorStore
	}

	vectorStore, err := vectorstorage.NewVectorStorage(
		ctx,
		sp.VectorStorageConfig(ctx),
		sp.Embedder(ctx),
		sp.Generator(ctx),
	)

	if err != nil {
		sp.Logger(ctx).Logger().Error("error creating vector storage", "error", err.Error())
		panic(fmt.Errorf("error creating vector storage: %w", err))
	}

	sp.vectorStore = vectorStore

	return vectorStore
}

// RepositoryConfig returns the repository configuration, creating it if it doesn't exist
func (sp *ServiceProvider) RepositoryConfig(ctx context.Context) *gormpg.Config {
	if sp.repositoryConfig != nil {
		return sp.repositoryConfig
	}

	config, err := gormpg.NewConfig()
	if err != nil {
		sp.Logger(ctx).Logger().Error("error creating repository config", "error", err.Error())
		panic(fmt.Errorf("error creating repository config: %w", err))
	}

	sp.repositoryConfig = config

	return config
}

// GormDB returns the GORM database instance, creating it if it doesn't exist
func (sp *ServiceProvider) GormDB(ctx context.Context) *gorm.DB {
	if sp.gormDB != nil {
		return sp.gormDB
	}

	db, err := gorm.Open(postgres.Open(sp.RepositoryConfig(ctx).DSN))
	if err != nil {
		sp.Logger(ctx).Logger().Error("error creating gorm db", "error", err.Error())
		panic(fmt.Errorf("error creating gorm db: %w", err))
	}

	sp.gormDB = db

	closer.Add(func() error {
		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("failed to get sql.DB instance: %w", err)
		}
		return sqlDB.Close()
	})

	return db
}

// Repository returns the GORM repository instance, creating it if it doesn't exist
func (sp *ServiceProvider) Repository(ctx context.Context) *gormpg.Repository {
	if sp.repository != nil {
		return sp.repository
	}

	repository, err := gormpg.NewRepository(sp.GormDB(ctx))
	if err != nil {
		sp.Logger(ctx).Logger().Error("error creating repository", "error", err.Error())
		panic(fmt.Errorf("error creating repository: %w", err))
	}

	sp.repository = repository

	return repository
}

// ResourceProcessor returns the resource processor instance, creating it if it doesn't exist
func (sp *ServiceProvider) ResourceProcessor(ctx context.Context) *resourceprocessor.ResourceProcessor {
	if sp.resourceProcessor != nil {
		return sp.resourceProcessor
	}

	resourceProcessor := resourceprocessor.NewResourceProcessor(sp.VectorStore(ctx))

	sp.resourceProcessor = resourceProcessor

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

// SearchController returns the search controller instance, creating it if it doesn't exist
func (sp *ServiceProvider) SearchController(ctx context.Context) *searchcontroller.Controller {
	if sp.searchController != nil {
		return sp.searchController
	}

	controller := searchcontroller.NewController(sp.SearchService(ctx))

	sp.searchController = controller

	return controller
}

// SearchService returns the search service instance, creating it if it doesn't exist
func (sp *ServiceProvider) SearchService(ctx context.Context) *searchservice.Service {
	if sp.searchService != nil {
		return sp.searchService
	}

	service := searchservice.NewService(sp.VectorStore(ctx), sp.Repository(ctx))

	sp.searchService = service

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
