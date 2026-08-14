package relation

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidType = errors.New("invalid relation type")
var ErrNotInSameFamily = errors.New("users must belong to the same family")
var ErrUserNotFound = errors.New("user not found")

type Relation struct {
	ID int64 `json:"id"`
	UserID int64 `json:"user_id"`
	RelatedUserID int64 `json:"related_user_id"`
	Type string `json:"type"`
}

type Repository interface {
	Create(ctx context.Context, familyID, userID, relatedUserID int64, relationType string) (Relation,error)
	List(ctx context.Context, familyID int64) ([]Relation,error)
	UserFamilyID(ctx context.Context,userID int64)(int64,error)
}

type Service struct{repo Repository}
func NewService(repo Repository)*Service{return &Service{repo:repo}}
func (s *Service) Create(ctx context.Context,familyID,userID,relatedUserID int64, relationType string)(Relation,error){
	relationType=strings.ToLower(strings.TrimSpace(relationType))
	switch relationType {case "parent","child","spouse","sibling":default:return Relation{},ErrInvalidType}
	if userID==relatedUserID{return Relation{},fmt.Errorf("cannot relate a user to themselves")}
	otherFamily,err:=s.repo.UserFamilyID(ctx,relatedUserID);if err!=nil{return Relation{},err}
	if otherFamily!=familyID{return Relation{},ErrNotInSameFamily}
	return s.repo.Create(ctx,familyID,userID,relatedUserID,relationType)
}
func(s *Service)List(ctx context.Context,familyID int64)([]Relation,error){return s.repo.List(ctx,familyID)}
