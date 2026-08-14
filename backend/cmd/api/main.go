package main

import (
	"context"
	"log"

	"github.com/Fabien-Halaby/kinspace/backend/internal/auth"
	"github.com/Fabien-Halaby/kinspace/backend/internal/config"
	"github.com/Fabien-Halaby/kinspace/backend/internal/database"
	"github.com/Fabien-Halaby/kinspace/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil { log.Fatal(err) }
	ctx := context.Background()
	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil { log.Fatal(err) }
	defer db.Close()
	if err := database.Migrate(ctx, db); err != nil { log.Fatal(err) }

	authService := auth.NewService(auth.NewPostgresRepository(db))
	tokens, err := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTTTL)
	if err != nil { log.Fatal(err) }

	router := gin.Default()
	router.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok", "service": "kinspace-api"}) })
	api := router.Group("/api/v1")
	api.POST("/auth/register", auth.RegisterHandler(authService))
	api.POST("/auth/login", auth.LoginHandler(authService, tokens.Issue))
	protected := api.Group("", middleware.RequireAuth(tokens))
	protected.GET("/auth/me", func(c *gin.Context) {
		id, _ := c.Get("user_id")
		c.JSON(200, gin.H{"user_id": id})
	})

	if err := router.Run(":" + cfg.Port); err != nil { log.Fatal(err) }
}
