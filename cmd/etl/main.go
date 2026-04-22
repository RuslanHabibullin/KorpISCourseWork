package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RuslanHabibullin/KorpISCourseWork/internal/etl"
	"github.com/RuslanHabibullin/KorpISCourseWork/internal/repository"
	"github.com/RuslanHabibullin/KorpISCourseWork/pkg/config"
	"github.com/RuslanHabibullin/KorpISCourseWork/pkg/logger"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("load config: " + err.Error())
	}

	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		panic("init logger: " + err.Error())
	}
	defer log.Sync() //nolint:errcheck

	store, err := repository.NewStore(cfg.DBDSN)
	if err != nil {
		log.Fatal("connect to db", zap.Error(err))
	}
	defer store.DB.Close()

	outDir := os.Getenv("ETL_OUT_DIR")

	extractor := etl.NewExtractor(store.DB)
	transformer := etl.NewTransformer()
	loader := etl.NewLoader(log, outDir)
	pipeline := etl.NewPipeline(extractor, transformer, loader, log)

	// Интервал между запусками
	interval := 1 * time.Hour
	if v := os.Getenv("ETL_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			interval = d
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Info("etl worker started", zap.Duration("interval", interval))

	// Первый запуск сразу
	if err := pipeline.Run(ctx); err != nil {
		log.Error("pipeline run failed", zap.Error(err))
	}

	for {
		select {
		case <-ticker.C:
			if err := pipeline.Run(ctx); err != nil {
				log.Error("pipeline run failed", zap.Error(err))
			}
		case <-quit:
			log.Info("etl worker stopping")
			cancel()
			return
		}
	}
}
