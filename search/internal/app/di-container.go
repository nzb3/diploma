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
	"github.com/tmc/langchaingo/llms/ollama"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nzb3/slogmanager"

	"github.com/nzb3/diploma/search/internal/controllers/resourcecontroller"
	"github.com/nzb3/diploma/search/internal/controllers/searchcontroller"
	"github.com/nzb3/diploma/search/internal/domain/services/resourceservcie"
	"github.com/nzb3/diploma/search/internal/domain/services/searchservice"
	"github.com/nzb3/diploma/search/internal/integration/generator"
	"github.com/nzb3/diploma/search/internal/repository/gormpg"
	"github.com/nzb3/diploma/search/internal/repository/vectorstorage"
	"github.com/nzb3/diploma/search/internal/server"

	"github.com/nzb3/diploma/search/internal/integration/embedder"
)

type serviceProvider struct {
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
}

func NewServiceProvider() *serviceProvider {
	return &serviceProvider{}
}

func (sp *serviceProvider) Logger(ctx context.Context) *slogmanager.Manager {
	if sp.slogManager != nil {
		return sp.slogManager
	}
	manager := slogmanager.New()
	manager.AddWriter("stdout", slogmanager.NewWriter(os.Stdout, slogmanager.WithTextFormat()))

	slog.SetLogLoggerLevel(slog.LevelDebug)
	return sp.slogManager
}

func (sp *serviceProvider) EmbeddingLLM(ctx context.Context) *ollama.LLM {
	if sp.embeddingLLM != nil {
		return sp.embeddingLLM
	}

	llm, err := ollama.New(
		ollama.WithServerURL("http://ollama-embedder:11434/"),
		ollama.WithModel("all-minilm"),
	)
	if err != nil {
		sp.Logger(ctx).Logger().Error("error creating ollama embedding LLM", err)
		panic(fmt.Errorf("error creating ollama embedding LLM: %w", err))
	}

	sp.embeddingLLM = llm

	return llm
}

func (sp *serviceProvider) GeneratingLLM(ctx context.Context) *ollama.LLM {
	if sp.generationLLM != nil {
		return sp.generationLLM
	}

	llm, err := ollama.New(
		ollama.WithServerURL("http://ollama-generator:11434/"),
		ollama.WithModel("llama3"),
	)

	if err != nil {
		sp.Logger(ctx).Logger().Error("error creating ollama generating LLM", err)
		panic(fmt.Errorf("error creating ollama generating LLM: %w", err))
	}

	sp.generationLLM = llm
	return llm
}

func (sp *serviceProvider) Embedder(ctx context.Context) *embedder.Embedder {
	if sp.embedder != nil {
		return sp.embedder
	}

	e, err := embedder.NewEmbedder(sp.EmbeddingLLM(ctx))
	if err != nil {
		sp.Logger(ctx).Logger().Error("error creating embedding LLM", err)
		panic(fmt.Errorf("error creating embedding LLM: %w", err))
	}

	sp.embedder = e

	return e
}

func (sp *serviceProvider) Generator(ctx context.Context) *generator.Generator {
	if sp.generator != nil {
		return sp.generator
	}

	g, err := generator.NewGenerator(sp.GeneratingLLM(ctx))
	if err != nil {
		sp.Logger(ctx).Logger().Error("error creating generating LLM", err)
		panic(fmt.Errorf("error creating generating LLM: %w", err))
	}

	sp.generator = g

	return g
}

func (sp *serviceProvider) GinEngine(ctx context.Context) *gin.Engine {
	if sp.ginEngine != nil {
		return sp.ginEngine
	}

	engine := gin.Default()

	// Configure CORS to allow frontend requests
	corsConfig := cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:5174", "http://localhost:5175", "http://localhost", "http://localhost:80", "http://front", "http://front:80"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	engine.Use(cors.New(corsConfig))

	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	sp.ginEngine = engine
	return engine
}

func (sp *serviceProvider) VectorStorageConfig(ctx context.Context) *vectorstorage.Config {
	if sp.vectorStorageConfig != nil {
		return sp.vectorStorageConfig
	}

	config, err := vectorstorage.NewConfig()
	if err != nil {
		sp.Logger(ctx).Logger().Error("error creating vector storage config", err)
		panic(fmt.Errorf("error creating vector storage config: %w", err))
	}

	sp.vectorStorageConfig = config

	return config
}

func (sp *serviceProvider) VectorStore(ctx context.Context) *vectorstorage.VectorStorage {
	if sp.vectorStore != nil {
		return sp.vectorStore
	}

	vectorStore, err := vectorstorage.NewVectorStorage(ctx, sp.VectorStorageConfig(ctx), sp.Embedder(ctx), sp.Generator(ctx))
	if err != nil {
		sp.Logger(ctx).Logger().Error("error creating vector storage", err)
		panic(fmt.Errorf("error creating vector storage: %w", err))
	}

	sp.vectorStore = vectorStore

	return vectorStore
}

func (sp *serviceProvider) RepositoryConfig(ctx context.Context) *gormpg.Config {
	if sp.repositoryConfig != nil {
		return sp.repositoryConfig
	}

	config, err := gormpg.NewConfig()
	if err != nil {
		sp.Logger(ctx).Logger().Error("error creating repository config", err)
		panic(fmt.Errorf("error creating repository config: %w", err))
	}

	sp.repositoryConfig = config

	return config
}

func (sp *serviceProvider) GormDB(ctx context.Context) *gorm.DB {
	if sp.gormDB != nil {
		return sp.gormDB
	}

	db, err := gorm.Open(postgres.Open(sp.RepositoryConfig(ctx).DSN))
	if err != nil {
		sp.Logger(ctx).Logger().Error("error creating gorm db", err)
		panic(fmt.Errorf("error creating gorm db: %w", err))
	}

	sp.gormDB = db

	return db
}

func (sp *serviceProvider) Repository(ctx context.Context) *gormpg.Repository {
	if sp.repository != nil {
		return sp.repository
	}

	repository, err := gormpg.NewRepository(sp.GormDB(ctx))
	if err != nil {
		sp.Logger(ctx).Logger().Error("error creating repository", err)
		panic(fmt.Errorf("error creating repository: %w", err))
	}

	sp.repository = repository

	return repository
}

func (sp *serviceProvider) ResourceService(ctx context.Context) *resourceservcie.Service {
	if sp.resourceService != nil {
		return sp.resourceService
	}

	service := resourceservcie.NewService(sp.VectorStore(ctx), sp.Repository(ctx))

	sp.resourceService = service

	return service
}

func (sp *serviceProvider) SearchController(ctx context.Context) *searchcontroller.Controller {
	if sp.searchController != nil {
		return sp.searchController
	}

	controller := searchcontroller.NewController(sp.SearchService(ctx))

	sp.searchController = controller

	return controller
}

func (sp *serviceProvider) SearchService(ctx context.Context) *searchservice.Service {
	if sp.searchService != nil {
		return sp.searchService
	}

	service := searchservice.NewService(sp.VectorStore(ctx), sp.Repository(ctx))

	sp.searchService = service

	return service
}

func (sp *serviceProvider) ResourceController(ctx context.Context) *resourcecontroller.Controller {
	if sp.resourceController != nil {
		return sp.resourceController
	}

	controller := resourcecontroller.NewController(sp.ResourceService(ctx))

	sp.resourceController = controller

	return controller
}

func (sp *serviceProvider) ServerConfig(ctx context.Context) *server.Config {
	if sp.serverConfig != nil {
		return sp.serverConfig
	}

	config, err := server.NewConfig()
	if err != nil {
		sp.Logger(ctx).Logger().Error("error creating server config", err)
		panic(fmt.Errorf("error creating server config: %w", err))
	}

	sp.serverConfig = config
	return config
}

func (sp *serviceProvider) Server(ctx context.Context) *http.Server {
	if sp.server != nil {
		return sp.server
	}

	s := server.NewServer(
		ctx,
		sp.GinEngine(ctx),
		sp.ServerConfig(ctx),
		sp.ResourceController(ctx),
		sp.SearchController(ctx),
	)

	sp.server = s
	return s
}
