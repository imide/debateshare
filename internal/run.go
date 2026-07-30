package internal

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/imide/debateshare/internal/api"
	"github.com/imide/debateshare/internal/config"
	"github.com/imide/debateshare/internal/logger"
	"github.com/imide/debateshare/internal/models"
	"github.com/imide/debateshare/internal/sse"
	"github.com/imide/debateshare/internal/storage"
	"github.com/imide/debateshare/internal/worker"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var RootDir string

func init() {
	exe, err := os.Executable()
	if err == nil {
		RootDir = filepath.Dir(filepath.Dir(exe))
	} else {
		RootDir, _ = os.Getwd()
	}
}

func Run(ctx context.Context, migrate bool) {
	// initialize logger with default settings first (silent)
	logger.Init(&config.Config{
		Log: config.LogConfig{Level: "info"},
	}, true)

	cfg, err := config.Load(RootDir)
	if err != nil {
		log.Fatal().Err(err).Msg("fatal error while loading config")
		os.Exit(1)
	}

	// reinit logger with actual config
	logger.Init(cfg, false)

	db, err := gorm.Open(postgres.Open(cfg.DB.DSN), &gorm.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("fatal error while connecting to database")
		os.Exit(1)
	}

	// migrate database
	if migrate {
		log.Info().Msg("migrating database")
		err = db.AutoMigrate(
			models.Room{},
			models.File{},
			models.Subscriber{},
		)

		if err != nil {
			log.Fatal().Err(err).Msg("fatal error while migrating database")
			os.Exit(1)
		}
		log.Info().Msg("database migration completed")
	} else {
		log.Info().Msg("skipped database migration")
	}

	store, err := storage.New(ctx, cfg.S3)
	if err != nil {
		log.Fatal().Err(err).Msg("fatal error while connecting to storage")
		os.Exit(1)
	}

	hub := sse.NewHub()

	expirer := worker.NewExpirer(db, store, cfg, hub)
	go expirer.Run(ctx)

	serverHandler := api.NewServer(db, store, cfg, hub)

	serverHandler.Init()

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: serverHandler.Router.Handler(),
	}

	go func() {
		log.Info().Str("addr", addr).Msg("server listening")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("fatal error while serving server")
			os.Exit(1)
		}
	}()

	// wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server...")

	// the context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("server forced to shutdown")
		os.Exit(1)
	}

	log.Info().Msg("server shutdown complete")
	os.Exit(0)

	//s3Client := newS3Client(cfg.S3)

	// app logic
	//roomRepo := repository.NewRoomRepository(db)
	//uploadRepo := repository.NewUploadRepository(db, s3Client)
	//wsRepo := repository.NewWebsocketRepository()
	//
	//roomService := service.NewRoomService(roomRepo)
	//uploadService := service.NewUploadService(uploadRepo)
	//wsService := service.NewWebsocketService(wsRepo)

}

//func newS3Client(cfg config.S3Config) *s3.Client {
//	awsCfg := aws.Config{
//		Region: "auto",
//		Credentials: credentials.NewStaticCredentialsProvider(
//			cfg.AWSAccessKeyID,
//			cfg.AWSSecretAccessKey,
//			"",
//		),
//	}
//
//	return s3.NewFromConfig(awsCfg, func(o *s3.Options) {
//		o.BaseEndpoint = &cfg.Endpoint
//		o.UsePathStyle = true
//	})
//}
