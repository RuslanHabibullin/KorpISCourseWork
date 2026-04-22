package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	apihttp "github.com/RuslanHabibullin/KorpISCourseWork/internal/api/http"
	"github.com/RuslanHabibullin/KorpISCourseWork/internal/api/mid"
	"github.com/RuslanHabibullin/KorpISCourseWork/internal/repository"
	"github.com/RuslanHabibullin/KorpISCourseWork/internal/repository/postgres"
	clientsvc "github.com/RuslanHabibullin/KorpISCourseWork/internal/service/client"
	ordersvc "github.com/RuslanHabibullin/KorpISCourseWork/internal/service/order"
	stocksvc "github.com/RuslanHabibullin/KorpISCourseWork/internal/service/stock"
	"github.com/RuslanHabibullin/KorpISCourseWork/pkg/config"
	"github.com/RuslanHabibullin/KorpISCourseWork/pkg/logger"

	"github.com/go-chi/chi/v5"
	chimid "github.com/go-chi/chi/v5/middleware"
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

	// ── Database ──────────────────────────────────────────────────────────────
	store, err := repository.NewStore(cfg.DBDSN)
	if err != nil {
		log.Fatal("connect to db", zap.Error(err))
	}
	defer store.DB.Close()
	log.Info("connected to postgres")

	// ── Repositories ──────────────────────────────────────────────────────────
	clientRepo := postgres.NewClientRepository(store.DB)
	vehicleRepo := postgres.NewVehicleRepository(store.DB)
	orderRepo := postgres.NewOrderRepository(store.DB)
	stockRepo := postgres.NewStockRepository(store.DB)
	partRepo := postgres.NewPartRepository(store.DB)
	svcCatalogRepo := postgres.NewServiceCatalogRepository(store.DB)

	// ── TxManager ─────────────────────────────────────────────────────────────
	txm := repository.NewTxManager(store.DB)

	// ── Services ──────────────────────────────────────────────────────────────
	clientService := clientsvc.NewService(clientRepo, vehicleRepo)
	orderService := ordersvc.NewService(orderRepo, stockRepo, partRepo, svcCatalogRepo, txm)
	stockService := stocksvc.NewService(stockRepo, partRepo)

	// ── HTTP handlers ─────────────────────────────────────────────────────────
	handler := apihttp.NewHandler(log, clientService, orderService, stockService)

	r := chi.NewRouter()
	r.Use(chimid.RequestID)
	r.Use(chimid.RealIP)
	r.Use(mid.Recovery(log))
	r.Use(mid.Logger(log))
	if cfg.AuthToken != "" {
		r.Use(mid.Auth(cfg.AuthToken))
	}

	// Health check
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r.Mount("/", handler.Routes())

	// ── Server ────────────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info("starting HTTP server", zap.String("addr", cfg.HTTPAddr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	<-quit
	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server shutdown error", zap.Error(err))
	}
	log.Info("server stopped")
}
