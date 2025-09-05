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

	"github.com/nzb3/diploma/search-service/internal/controllers"
	"github.com/nzb3/diploma/search-service/internal/controllers/middleware"
	"github.com/nzb3/diploma/search-service/internal/controllers/searchcontroller"
	"github.com/nzb3/diploma/search-service/internal/domain/services/eventservice"
	"github.com/nzb3/diploma/search-service/internal/domain/services/outboxprocessor"
	"github.com/nzb3/diploma/search-service/internal/domain/services/resourceprocessor"
	"github.com/nzb3/diploma/search-service/internal/domain/services/searchservice"
	"github.com/nzb3/diploma/search-service/internal/repository/embedder"
	"github.com/nzb3/diploma/search-service/internal/repository/events/pgx"
	"github.com/nzb3/diploma/search-service/internal/repository/generator"
	"github.com/nzb3/diploma/search-service/internal/repository/messaging"
	"github.com/nzb3/diploma/search-service/internal/repository/messaging/kafka"
	"github.com/nzb3/diploma/search-service/internal/repository/postgres"
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
	ginEngine           *gin.Engine
	vectorStore         *vectorstorage.VectorStorage
	vectorStorageConfig *vectorstorage.Config
	postgresConfig      *postgres.Config
	serverConfig        *server.Config
	kafkaConfig         *kafka.Config
	authConfig          *middleware.AuthConfig
	gormDB              *gorm.DB
	searchController    *searchcontroller.Controller
	searchService       *searchservice.Service
	authMiddleware      *middleware.AuthMiddleware
	// Event system components
	pgxPool           *pgxpool.Pool
	eventRepository   *pgx.Repository
	kafkaProducer     *kafka.Producer
	kafkaConsumer     messaging.MessageConsumer
	eventService      *eventservice.Service
	outboxProcessor   *outboxprocessor.Processor
	resourceProcessor *resourceprocessor.Processor
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
	sp.slogManager = manager
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

// PostgresConfig returns the PostgreSQL configuration, creating it if it doesn't exist
func (sp *ServiceProvider) PostgresConfig(ctx context.Context) *postgres.Config {
	if sp.postgresConfig != nil {
		return sp.postgresConfig
	}

	config, err := postgres.NewConfig()
	if err != nil {
		sp.Logger(ctx).Logger().Error("error creating postgres config", "error", err.Error())
		panic(fmt.Errorf("error creating postgres config: %w", err))
	}

	sp.postgresConfig = config
	return config
}

// KafkaConfig returns the Kafka configuration, creating it if it doesn't exist
func (sp *ServiceProvider) KafkaConfig(ctx context.Context) *kafka.Config {
	if sp.kafkaConfig != nil {
		return sp.kafkaConfig
	}

	config, err := kafka.NewConfig()
	if err != nil {
		sp.Logger(ctx).Logger().Error("error creating kafka config", "error", err.Error())
		panic(fmt.Errorf("error creating kafka config: %w", err))
	}

	sp.kafkaConfig = config
	return config
}

// AuthConfig returns the auth configuration, creating it if it doesn't exist
func (sp *ServiceProvider) AuthConfig(ctx context.Context) *middleware.AuthConfig {
	if sp.authConfig != nil {
		return sp.authConfig
	}

	config, err := middleware.NewAuthConfig()
	if err != nil {
		sp.Logger(ctx).Logger().Error("error creating auth config", "error", err.Error())
		panic(fmt.Errorf("error creating auth config: %w", err))
	}

	sp.authConfig = config
	return config
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

	config, err := vectorstorage.NewConfig()
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
		sp.PostgresConfig(ctx),
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

	// Create search service with optional event service
	service := searchservice.NewService(
		sp.VectorStore(ctx),
		sp.EventService(ctx),
	)

	sp.searchService = service

	return service
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

// PgxPool returns the PostgreSQL connection pool, creating it if it doesn't exist
func (sp *ServiceProvider) PgxPool(ctx context.Context) *pgxpool.Pool {
	if sp.pgxPool != nil {
		return sp.pgxPool
	}

	postgresConfig := sp.PostgresConfig(ctx)
	dbURL := postgresConfig.GetConnectionString()

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		sp.Logger(ctx).Logger().Error("error parsing database URL", "error", err.Error())
		panic(fmt.Errorf("error parsing database URL: %w", err))
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		sp.Logger(ctx).Logger().Error("error creating database pool", "error", err.Error())
		panic(fmt.Errorf("error creating database pool: %w", err))
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		sp.Logger(ctx).Logger().Error("error pinging database", "error", err.Error())
		panic(fmt.Errorf("error pinging database: %w", err))
	}

	sp.pgxPool = pool
	return pool
}

// EventRepository returns the event repository instance, creating it if it doesn't exist
func (sp *ServiceProvider) EventRepository(ctx context.Context) *pgx.Repository {
	if sp.eventRepository != nil {
		return sp.eventRepository
	}

	repo, err := pgx.NewRepository(ctx, sp.PgxPool(ctx))
	if err != nil {
		sp.Logger(ctx).Logger().Error("error creating event repository", "error", err.Error())
		panic(fmt.Errorf("error creating event repository: %w", err))
	}

	sp.eventRepository = repo
	return repo
}

// KafkaProducer returns the Kafka producer instance, creating it if it doesn't exist
func (sp *ServiceProvider) KafkaProducer(ctx context.Context) *kafka.Producer {
	if sp.kafkaProducer != nil {
		return sp.kafkaProducer
	}

	kafkaConfig := sp.KafkaConfig(ctx)

	producer, err := kafka.NewKafkaProducer(kafkaConfig)
	if err != nil {
		sp.Logger(ctx).Logger().Error("error creating Kafka producer", "error", err.Error())
		panic(fmt.Errorf("error creating Kafka producer: %w", err))
	}

	sp.kafkaProducer = producer
	return producer
}

// EventService returns the event service instance, creating it if it doesn't exist
func (sp *ServiceProvider) EventService(ctx context.Context) *eventservice.Service {
	if sp.eventService != nil {
		return sp.eventService
	}

	service := eventservice.NewEventService(
		sp.EventRepository(ctx),
		sp.KafkaProducer(ctx),
	)

	sp.eventService = service
	return service
}

// OutboxProcessor returns the outbox processor instance, creating it if it doesn't exist
func (sp *ServiceProvider) OutboxProcessor(ctx context.Context) *outboxprocessor.Processor {
	if sp.outboxProcessor != nil {
		return sp.outboxProcessor
	}

	processor := outboxprocessor.NewDefaultOutboxProcessor(
		sp.EventService(ctx),
	)

	sp.outboxProcessor = processor
	return processor
}

// KafkaConsumer returns the Kafka consumer instance, creating it if it doesn't exist
func (sp *ServiceProvider) KafkaConsumer(ctx context.Context) messaging.MessageConsumer {
	if sp.kafkaConsumer != nil {
		return sp.kafkaConsumer
	}

	kafkaConsumerConfig, err := kafka.NewConsumerConfig()
	if err != nil {
		// Log the error if logger is available, otherwise use standard logging
		if sp.slogManager != nil {
			sp.Logger(ctx).Logger().Error("error creating kafka consumer config", "error", err.Error())
		} else {
			slog.Error("error creating kafka consumer config", "error", err.Error())
		}
		panic(fmt.Errorf("error creating kafka consumer config: %w", err))
	}

	consumer, err := kafka.NewKafkaConsumer(kafkaConsumerConfig)
	if err != nil {
		// Log the error if logger is available, otherwise use standard logging
		if sp.slogManager != nil {
			sp.Logger(ctx).Logger().Error("error creating kafka consumer", "error", err.Error())
		} else {
			slog.Error("error creating kafka consumer", "error", err.Error())
		}
		panic(fmt.Errorf("error creating kafka consumer: %w", err))
	}

	sp.kafkaConsumer = consumer
	return consumer
}

// ResourceProcessor returns the resource processor instance, creating it if it doesn't exist
func (sp *ServiceProvider) ResourceProcessor(ctx context.Context) *resourceprocessor.Processor {
	if sp.resourceProcessor != nil {
		return sp.resourceProcessor
	}

	processor := resourceprocessor.NewResourceProcessor(
		sp.VectorStore(ctx),
		sp.EventService(ctx),
		sp.KafkaConsumer(ctx),
	)

	sp.resourceProcessor = processor
	return processor
}
