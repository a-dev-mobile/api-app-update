// Package main is the entry point of the github.com/a-dev-mobile/app-update-api application.
// It sets up and starts a web server using the Gin framework, establishes database connections,
// and initializes various HTTP handlers for the application's routes.
//
// The package orchestrates the overall flow of the application, including:
// - Loading and validating configuration settings.
// - Setting up logging and database connections.
// - Initializing and configuring the Gin router with necessary middleware.
// - Defining routes for app update checks, uploads, and downloads.
// - Gracefully handling server shutdowns.
//
// This package is designed to bootstrap the necessary components and start the web service,
// making the github.com/a-dev-mobile/app-update-api ready to serve incoming HTTP requests.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	docs "github.com/a-dev-mobile/app-update-api/docs"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/a-dev-mobile/app-update-api/database/mongo"
	"github.com/a-dev-mobile/app-update-api/internal/config"

	"golang.org/x/exp/slog"

	"github.com/a-dev-mobile/app-update-api/internal/api/v1/check"
	"github.com/a-dev-mobile/app-update-api/internal/api/v1/download"
	"github.com/a-dev-mobile/app-update-api/internal/api/v1/upload"
	"github.com/a-dev-mobile/app-update-api/internal/constants"
	"github.com/a-dev-mobile/app-update-api/internal/logging"
	"github.com/a-dev-mobile/app-update-api/internal/middleware"

	mongodriver "go.mongodb.org/mongo-driver/mongo"
)

func main() {
	cfg, lg := initializeApp()

	setGinMode(cfg)

	db, cleanup := setupDatabase(cfg, lg)
	defer cleanup()

	router := setupRouter(cfg, lg)

	hctxCheck := check.NewHandlerContext(db, lg, cfg)
	router.POST(constants.CheckAPIEndpoint, hctxCheck.Check)

	hctxUpload := upload.NewHandlerContext(db, lg, cfg)
	router.POST(constants.UploadAPIEndpoint, hctxUpload.Upload)

	hctxDownload := download.NewHandlerContext(lg, cfg)
	router.GET(constants.DownloadAPIEndpoint, hctxDownload.DownloadApk)

	docs.SwaggerInfo.Title = "App Update API"
	docs.SwaggerInfo.Description = "Application update control"
	docs.SwaggerInfo.Host = cfg.Swagger.Host
	docs.SwaggerInfo.BasePath = "/app-update-api/v1"
	docs.SwaggerInfo.Version = "v1"

	router.GET("app-update-api/v1/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	lg.Info("Server starting", "env", string(cfg.Environment))

	startServer(cfg, router, lg)
}

// initializeApp sets up the application environment, configuration, and logger.
// It determines the application's running environment, loads the appropriate configuration,
// and initializes the logging system.
func initializeApp() (*config.Config, *slog.Logger) {

	cfg := getConfigOrFail()

	lg := logging.SetupLogger(cfg)

	return cfg, lg
}
func getConfigOrFail() *config.Config {

	cfg, err := config.LoadConfig("../config", "config.yaml")

	if err != nil {
		log.Fatalf("Error loading config: %s", err)
	}

	return cfg
}

// setGinMode configures the Gin mode (debug or release) based on the application's configuration.
func setGinMode(cfg *config.Config) {
	switch cfg.ClientConnection.GinMode {
	case "release":
		gin.SetMode(gin.ReleaseMode)
	default:
		gin.SetMode(gin.DebugMode)
	}
}

// setupDatabase initializes the MongoDB database connection using the provided configuration and logger.
// It returns a MongoDB client and a cleanup function to disconnect from the database.
func setupDatabase(cfg *config.Config, lg *slog.Logger) (*mongodriver.Client, func()) {
	dbUser := cfg.Database.User
	dbPassword := cfg.Database.Password
	dbPort := cfg.Database.Port
	dbHost := cfg.Database.Host

	db, err := mongo.GetDB(context.Background(), dbUser, dbPassword, dbHost, dbPort)
	if err != nil {
		lg.Error("Error initializing database", "error", err.Error())
		os.Exit(1)
	}
	return db, func() {
		if err := db.Disconnect(context.Background()); err != nil {
			lg.Error("Error disconnecting from database", "error", err.Error())
		}
	}
}

// setupRouter initializes and returns a new Gin router configured with middleware and routes.
func setupRouter(cfg *config.Config, lg *slog.Logger) *gin.Engine {
	// Create a new router
	router := gin.New()
	// Apply global middleware
	router.Use(gin.Recovery()) // Recovery middleware от Gin
	router.Use(gin.Logger())   // Logging middleware от Gin
	router.Use(middleware.CORSMiddleware(*cfg, lg))
	router.Use(middleware.TrustProxyHeader())

	// Adding custom middleware to recover from a panic
	router.Use(middleware.RecoveryMiddleware(lg))

	return router
}

// startServer starts the HTTP server using the provided configuration and router,
// and handles graceful shutdown on receiving quit signals.
func startServer(cfg *config.Config, router *gin.Engine, lg *slog.Logger) {
	serverAddr := fmt.Sprintf(":%s", cfg.ClientConnection.Port)
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: router,

		WriteTimeout: 1 * time.Minute,
		ReadTimeout:  1 * time.Minute,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			lg.Error("Error running server", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	gracefulShutdown(srv, lg)
}
func gracefulShutdown(srv *http.Server, lg *slog.Logger) {
	quit := make(chan os.Signal, 1)
	// signal.Notify(quit, os.Interrupt) // Uncomment this line if you want to listen to OS interrupt signals

	<-quit
	lg.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultServerTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		lg.Error("Server forced to shutdown:", slog.String("error", err.Error()))
		os.Exit(1)
	}

	lg.Info("Server exiting")
}
func LogHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		currentTime := time.Now().Format("2006/01/02 15:04:05")
		log.Printf("\n[%s] Headers for %s %s:\n", currentTime, c.Request.Method, c.Request.URL.Path)
		for k, v := range c.Request.Header {
			log.Printf("  %s: %s", k, strings.Join(v, ","))
		}
		c.Next()
	}
}
