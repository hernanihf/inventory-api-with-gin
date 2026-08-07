package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"inventory_api/config"
	"inventory_api/database"
	_ "inventory_api/docs"
	"inventory_api/handler"
	"inventory_api/middleware"
	"inventory_api/store"
	"inventory_api/worker"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const serverPort = "8080"

// @title           Inventory API
// @version         1.0
// @description     Inventory API — Project to learn Go + Gin.
// @host            localhost:8080
// @BasePath        /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	r := gin.New()
	apiKey := readApiKey()
	db, done := connectDb()
	if done {
		return
	}
	defer db.Close()
	newStore := store.InitStore(db)

	productStoreHandler := handler.NewStore(newStore)
	r.Use(middleware.RequestLogger(), gin.Recovery())
	r.GET("/", productStoreHandler.HelloWorld)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/favicon.ico", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	r.GET("/products/:name", productStoreHandler.GetProduct)
	r.GET("/products", productStoreHandler.GetProducts)

	protected := r.Group("/products")
	protected.Use(middleware.Auth(apiKey))
	protected.POST("", productStoreHandler.AddProduct)
	protected.PUT("/:name/sell", productStoreHandler.SellProduct)
	protected.PUT("/:name/addStock", productStoreHandler.AddProductStock)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go worker.StartLowStockMonitor(ctx, newStore)
	startServer(r, ctx)
}

func startServer(r *gin.Engine, ctx context.Context) {
	srv := &http.Server{
		Addr:    ":" + serverPort,
		Handler: r,
	}

	go getStartServerFc(srv)()

	<-ctx.Done()
	fmt.Println("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		fmt.Println("Error shutting down:", err)
	}
}

func getStartServerFc(srv *http.Server) func() {
	return func() {
		fmt.Printf("Listening on port %s\n", serverPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Println(err)
		}
	}
}

func readApiKey() string {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		fmt.Println("Missing API_KEY environment variable")
		os.Exit(1)
	}
	return apiKey
}

func connectDb() (*sql.DB, bool) {
	readConfig, err := config.ReadConfig("config.json")
	if err != nil {
		fmt.Println(err)
		return nil, true
	}

	db, err2 := database.ConnectToDatabase(readConfig)
	if err2 != nil {
		fmt.Println(err2)
		return nil, true
	}
	return db, false
}
