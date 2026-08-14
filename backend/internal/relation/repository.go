package relation

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresRepository struct{db *pgxpool.Pool}
func NewPostgresRepository(db *pgxpool.Pool)Repository{return &postgresRepository{db:db}}
func(r *postgresRepository)Create(ctx context.Context,familyID,userID,relatedUserID int64,relationType string)(Relation,error){var out Relation;err:=r.db.QueryRow(ctx,`INSERT INTO relations(family_id,user_id,related_user_id,type) VALUES($1,$2,$3,$4) RETURNING id,user_id,related_user_id,type`,familyID,userID,relatedUserID,relationType).Scan(&out.ID,&out.UserID,&out.RelatedUserID,&out.Type);if err!=nil{return Relation{},fmt.Errorf("create relation: %w",err)};return out,nil}
func(r *postgresRepository)List(ctx context.Context,familyID int64)([]Relation,error){rows,err:=r.db.Query(ctx,`SELECT id,user_id,related_user_id,type FROM relations WHERE family_id=$1 ORDER BY id`,familyID);if err!=nil{return nil,fmt.Errorf("list relations: %w",err)};defer rows.Close();out:=make([]Relation,0);for rows.Next(){var v Relation;if err:=rows.Scan(&v.ID,&v.UserID,&v.RelatedUserID,&v.Type);err!=nil{return nil,fmt.Errorf("scan relation: %w",err)};out=append(out,v)};if err:=rows.Err();err!=nil{return nil,fmt.Errorf("iterate relations: %w",err)};return out,nil}
func(r *postgresRepository)UserFamilyID(ctx context.Context,userID int64)(int64,error){var id *int64;err:=r.db.QueryRow(ctx,`SELECT family_id FROM users WHERE id=$1`,userID).Scan(&id);if errors.Is(err,pgx.ErrNoRows){return 0,ErrUserNotFound};if err!=nil{return 0,fmt.Errorf("find user family: %w",err)};if id==nil{return 0,ErrNotInSameFamily};return *id,nil}
func RegisterHandlers(router *gin.RouterGroup,service *Service,authUser func(*gin.Context)(int64,bool),familyID func(*gin.Context,int64)(int64,error)){router.POST("/relations",func(c *gin.Context){userID,ok:=authUser(c);if !ok{c.JSON(http.StatusUnauthorized,gin.H{"error":"unauthorized"});return};fid,err:=familyID(c,userID);if err!=nil{c.JSON(http.StatusNotFound,gin.H{"error":"family not found"});return};var req struct{RelatedUserID int64 `json:"related_user_id" binding:"required"`;Type string `json:"type" binding:"required"`};if err:=c.ShouldBindJSON(&req);err!=nil{c.JSON(http.StatusBadRequest,gin.H{"error":"invalid request"});return};v,err:=service.Create(c.Request.Context(),fid,userID,req.RelatedUserID,req.Type);if errors.Is(err,ErrNotInSameFamily){c.JSON(http.StatusForbidden,gin.H{"error":err.Error()});return};if errors.Is(err,ErrInvalidType){c.JSON(http.StatusBadRequest,gin.H{"error":err.Error()});return};if errors.Is(err,ErrUserNotFound){c.JSON(http.StatusNotFound,gin.H{"error":err.Error()});return};if err!=nil{c.JSON(http.StatusBadRequest,gin.H{"error":err.Error()});return};c.JSON(http.StatusCreated,gin.H{"relation":v})});router.GET("/relations",func(c *gin.Context){userID,ok:=authUser(c);if !ok{c.JSON(http.StatusUnauthorized,gin.H{"error":"unauthorized"});return};fid,err:=familyID(c,userID);if err!=nil{c.JSON(http.StatusNotFound,gin.H{"error":"family not found"});return};v,err:=service.List(c.Request.Context(),fid);if err!=nil{c.JSON(http.StatusInternalServerError,gin.H{"error":"could not load relations"});return};c.JSON(http.StatusOK,gin.H{"relations":v})})}
