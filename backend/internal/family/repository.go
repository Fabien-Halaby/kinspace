package family

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Fabien-Halaby/kinspace/backend/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresRepository struct{ db *pgxpool.Pool }
func NewPostgresRepository(db *pgxpool.Pool) Repository { return &postgresRepository{db: db} }

func (r *postgresRepository) Create(ctx context.Context, name, inviteCode string, ownerID int64) (Family, error) {
	tx, err := r.db.Begin(ctx); if err != nil { return Family{}, fmt.Errorf("begin family transaction: %w", err) }
	defer tx.Rollback(ctx)
	var f Family
	if err := tx.QueryRow(ctx, `INSERT INTO families(name, invite_code) VALUES($1,$2) RETURNING id,name,invite_code`, name, inviteCode).Scan(&f.ID,&f.Name,&f.InviteCode); err != nil { return Family{}, fmt.Errorf("create family: %w", err) }
	if _, err := tx.Exec(ctx, `UPDATE users SET family_id=$1 WHERE id=$2 AND family_id IS NULL`, f.ID, ownerID); err != nil { return Family{}, fmt.Errorf("assign family owner: %w", err) }
	if err := tx.Commit(ctx); err != nil { return Family{}, fmt.Errorf("commit family: %w", err) }
	return f,nil
}
func (r *postgresRepository) FindByUserID(ctx context.Context, userID int64) (Family,error) {
	var f Family
	err:=r.db.QueryRow(ctx, `SELECT f.id,f.name,f.invite_code FROM families f JOIN users u ON u.family_id=f.id WHERE u.id=$1`,userID).Scan(&f.ID,&f.Name,&f.InviteCode)
	if errors.Is(err,pgx.ErrNoRows){return Family{},ErrFamilyNotFound}; if err!=nil{return Family{},fmt.Errorf("find family: %w",err)}; return f,nil
}
func (r *postgresRepository) Join(ctx context.Context,userID int64,inviteCode string)(Family,error){
	var f Family
	err:=r.db.QueryRow(ctx,`SELECT id,name,invite_code FROM families WHERE invite_code=$1`,inviteCode).Scan(&f.ID,&f.Name,&f.InviteCode)
	if errors.Is(err,pgx.ErrNoRows){return Family{},ErrInviteCodeInvalid}; if err!=nil{return Family{},fmt.Errorf("find invite: %w",err)}
	result,err:=r.db.Exec(ctx,`UPDATE users SET family_id=$1 WHERE id=$2 AND family_id IS NULL`,f.ID,userID); if err!=nil{return Family{},fmt.Errorf("join family: %w",err)}
	if result.RowsAffected()!=1{return Family{},ErrAlreadyInFamily}; return f,nil
}

func RegisterHandlers(router *gin.RouterGroup, service *Service, authUser func(*gin.Context)(int64,bool)) {
	router.POST("/families", func(c *gin.Context){
		userID,ok:=authUser(c); if !ok { c.JSON(http.StatusUnauthorized,gin.H{"error":"unauthorized"}); return }
		var req struct{Name string `json:"name" binding:"required"`}; if err:=c.ShouldBindJSON(&req);err!=nil{c.JSON(http.StatusBadRequest,gin.H{"error":"invalid request"});return}
		f,err:=service.Create(c.Request.Context(),userID,req.Name); if errors.Is(err,ErrAlreadyInFamily){c.JSON(http.StatusConflict,gin.H{"error":err.Error()});return};if err!=nil{c.JSON(http.StatusBadRequest,gin.H{"error":err.Error()});return};c.JSON(http.StatusCreated,gin.H{"family":f})
	})
	router.POST("/families/join",func(c *gin.Context){userID,ok:=authUser(c);if !ok{c.JSON(http.StatusUnauthorized,gin.H{"error":"unauthorized"});return};var req struct{InviteCode string `json:"invite_code" binding:"required"`};if err:=c.ShouldBindJSON(&req);err!=nil{c.JSON(http.StatusBadRequest,gin.H{"error":"invalid request"});return};f,err:=service.Join(c.Request.Context(),userID,req.InviteCode);if errors.Is(err,ErrInviteCodeInvalid){c.JSON(http.StatusNotFound,gin.H{"error":err.Error()});return};if errors.Is(err,ErrAlreadyInFamily){c.JSON(http.StatusConflict,gin.H{"error":err.Error()});return};if err!=nil{c.JSON(http.StatusBadRequest,gin.H{"error":err.Error()});return};c.JSON(http.StatusOK,gin.H{"family":f})})
	router.GET("/families/me",func(c *gin.Context){userID,ok:=authUser(c);if !ok{c.JSON(http.StatusUnauthorized,gin.H{"error":"unauthorized"});return};f,err:=service.repo.FindByUserID(c.Request.Context(),userID);if errors.Is(err,ErrFamilyNotFound){c.JSON(http.StatusNotFound,gin.H{"error":err.Error()});return};if err!=nil{c.JSON(http.StatusInternalServerError,gin.H{"error":"could not load family"});return};c.JSON(http.StatusOK,gin.H{"family":f})})
}

var _ auth.User
var _ = strings.TrimSpace
