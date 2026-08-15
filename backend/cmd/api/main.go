package main

import (
	"context"
	"errors"
	"log"

	"github.com/Fabien-Halaby/kinspace/backend/internal/auth"
	"github.com/Fabien-Halaby/kinspace/backend/internal/config"
	"github.com/Fabien-Halaby/kinspace/backend/internal/database"
	"github.com/Fabien-Halaby/kinspace/backend/internal/family"
	"github.com/Fabien-Halaby/kinspace/backend/internal/middleware"
	"github.com/Fabien-Halaby/kinspace/backend/internal/relation"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load(); if err != nil { log.Fatal(err) }
	ctx := context.Background()
	db, err := database.Connect(ctx, cfg.DatabaseURL); if err != nil { log.Fatal(err) }; defer db.Close()
	if err := database.Migrate(ctx, db); err != nil { log.Fatal(err) }

	authService := auth.NewService(auth.NewPostgresRepository(db))
	tokens, err := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTTTL); if err != nil { log.Fatal(err) }
	familyService := family.NewService(family.NewPostgresRepository(db))
	relationService := relation.NewService(relation.NewPostgresRepository(db))

	router := gin.Default()
	router.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status":"ok","service":"kinspace-api"}) })
	api := router.Group("/api/v1")
	api.POST("/auth/register", auth.RegisterHandler(authService))
	api.POST("/auth/login", auth.LoginHandler(authService, tokens.Issue))

	protected := api.Group("", middleware.RequireAuth(tokens))
	authUser := func(c *gin.Context) (int64, bool) { v, ok := c.Get("user_id"); if !ok { return 0,false }; id,ok:=v.(int64); return id,ok }
	familyID := func(c *gin.Context, userID int64) (int64,error) { f,err:=familyService.Me(c.Request.Context(),userID); return f.ID,err }

	family.RegisterHandlers(protected, familyService, authUser)
	relation.RegisterHandlers(protected, relationService, authUser, familyID)
	protected.GET("/auth/me", func(c *gin.Context) { id,_:=authUser(c); c.JSON(200,gin.H{"user_id":id}) })
	protected.GET("/families/me", func(c *gin.Context) { id,_:=authUser(c); f,err:=familyService.Me(c.Request.Context(),id); if errors.Is(err,family.ErrFamilyNotFound){c.JSON(404,gin.H{"error":err.Error()});return};if err!=nil{c.JSON(500,gin.H{"error":"could not load family"});return};c.JSON(200,gin.H{"family":f}) })

	if err := router.Run(":" + cfg.Port); err != nil { log.Fatal(err) }
}
