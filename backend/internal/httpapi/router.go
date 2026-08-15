package httpapi

import (
	"context"
	"log/slog"

	"github.com/Fabien-Halaby/kinspace/backend/internal/domain"
	"github.com/Fabien-Halaby/kinspace/backend/internal/token"
	"github.com/gin-gonic/gin"
)

// The following ports describe exactly what the HTTP layer needs from
// the application layer. Concrete services satisfy them implicitly,
// which keeps this layer decoupled and easy to test with stubs.

type AuthService interface {
	Register(ctx context.Context, name, email, password string) (domain.User, string, error)
	Login(ctx context.Context, email, password string) (domain.User, string, error)
	Me(ctx context.Context, userID int64) (domain.User, error)
}

type FamilyService interface {
	Me(ctx context.Context, userID int64) (domain.Family, error)
	Create(ctx context.Context, userID int64, name string) (domain.Family, error)
	Join(ctx context.Context, userID int64, inviteCode string) (domain.Family, error)
}

type RelationService interface {
	Create(ctx context.Context, familyID, userID, relatedUserID int64, relationType string) (domain.Relation, error)
	List(ctx context.Context, familyID int64) ([]domain.Relation, error)
}

// Dependencies are injected by the composition root (cmd/api).
type Dependencies struct {
	Environment string
	Logger      *slog.Logger
	Tokens      token.Manager
	Auth        AuthService
	Families    FamilyService
	Relations   RelationService
}

// NewRouter assembles the HTTP application: middleware, routes and
// handler wiring. No business logic lives here.
func NewRouter(deps Dependencies) *gin.Engine {
	if deps.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery(), RequestLogger(deps.Logger), SecurityHeaders())

	auth := AuthHandler{service: deps.Auth}
	families := FamilyHandler{service: deps.Families}
	relations := RelationHandler{families: deps.Families, relations: deps.Relations}

	router.GET("/health", healthHandler)

	api := router.Group("/api/v1")
	api.POST("/auth/register", auth.Register)
	api.POST("/auth/login", auth.Login)

	protected := api.Group("")
	protected.Use(RequireAuth(deps.Tokens))
	protected.GET("/auth/me", auth.Me)
	protected.POST("/families", families.Create)
	protected.POST("/families/join", families.Join)
	protected.GET("/families/me", families.Me)
	protected.POST("/relations", relations.Create)
	protected.GET("/relations", relations.List)

	return router
}

func healthHandler(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok", "service": "kinspace-api"})
}
