package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Fabien-Halaby/kinspace/backend/internal/domain"
	"github.com/Fabien-Halaby/kinspace/backend/internal/token"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// --- stubs ---------------------------------------------------------------

type stubAuthService struct {
	registerUser  domain.User
	registerToken string
	registerErr   error
	loginUser     domain.User
	loginToken    string
	loginErr      error
	meUser        domain.User
	meErr         error
}

func (s *stubAuthService) Register(context.Context, string, string, string) (domain.User, string, error) {
	return s.registerUser, s.registerToken, s.registerErr
}
func (s *stubAuthService) Login(context.Context, string, string) (domain.User, string, error) {
	return s.loginUser, s.loginToken, s.loginErr
}
func (s *stubAuthService) Me(context.Context, int64) (domain.User, error) {
	return s.meUser, s.meErr
}

type stubFamilyService struct {
	meFamily  domain.Family
	meErr     error
	created   domain.Family
	createErr error
	joined    domain.Family
	joinErr   error
}

func (s *stubFamilyService) Me(context.Context, int64) (domain.Family, error) {
	return s.meFamily, s.meErr
}
func (s *stubFamilyService) Create(context.Context, int64, string) (domain.Family, error) {
	return s.created, s.createErr
}
func (s *stubFamilyService) Join(context.Context, int64, string) (domain.Family, error) {
	return s.joined, s.joinErr
}

type stubRelationService struct {
	created   domain.Relation
	createErr error
	listed    []domain.Relation
	listErr   error
}

func (s *stubRelationService) Create(context.Context, int64, int64, int64, string) (domain.Relation, error) {
	return s.created, s.createErr
}
func (s *stubRelationService) List(context.Context, int64) ([]domain.Relation, error) {
	return s.listed, s.listErr
}

type stubTokens struct {
	userID   int64
	parseErr error
}

func (t *stubTokens) Issue(int64) (string, error) { return "test-token", nil }
func (t *stubTokens) Parse(string) (int64, error) {
	return t.userID, t.parseErr
}

// --- helpers ---------------------------------------------------------------

func newTestRouter(auth AuthService, families FamilyService, relations RelationService) *gin.Engine {
	return NewRouter(Dependencies{
		Environment: "test",
		Logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
		Tokens:      &stubTokens{userID: 7},
		Auth:        auth,
		Families:    families,
		Relations:   relations,
	})
}

func doRequest(router http.Handler, method, path, body, bearer string) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != "" {
		reader = bytes.NewBufferString(body)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode response body: %v (body: %s)", err, rec.Body.String())
	}
	return out
}

// --- auth handler tests -----------------------------------------------------

func TestRegisterReturnsUserAndToken(t *testing.T) {
	auth := &stubAuthService{
		registerUser:  domain.User{ID: 1, Name: "Fabien", Email: "fabien@example.com"},
		registerToken: "signed-token",
	}
	router := newTestRouter(auth, &stubFamilyService{}, &stubRelationService{})

	rec := doRequest(router, http.MethodPost, "/api/v1/auth/register",
		`{"name":"Fabien","email":"fabien@example.com","password":"correct-password"}`, "")

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (body: %s)", rec.Code, rec.Body.String())
	}
	body := decodeBody(t, rec)
	if body["token"] != "signed-token" {
		t.Fatalf("token = %v, want signed-token", body["token"])
	}
}

func TestRegisterDuplicateReturnsConflict(t *testing.T) {
	auth := &stubAuthService{registerErr: domain.ErrEmailExists}
	router := newTestRouter(auth, &stubFamilyService{}, &stubRelationService{})

	rec := doRequest(router, http.MethodPost, "/api/v1/auth/register",
		`{"name":"Fabien","email":"fabien@example.com","password":"correct-password"}`, "")

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409 (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestRegisterMalformedBodyReturnsBadRequest(t *testing.T) {
	router := newTestRouter(&stubAuthService{}, &stubFamilyService{}, &stubRelationService{})

	rec := doRequest(router, http.MethodPost, "/api/v1/auth/register", `{"name":`, "")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestLoginInvalidCredentialsReturnsUnauthorized(t *testing.T) {
	auth := &stubAuthService{loginErr: domain.ErrInvalidCredentials}
	router := newTestRouter(auth, &stubFamilyService{}, &stubRelationService{})

	rec := doRequest(router, http.MethodPost, "/api/v1/auth/login",
		`{"email":"fabien@example.com","password":"wrong"}`, "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 (body: %s)", rec.Code, rec.Body.String())
	}
}

// --- protected route tests ---------------------------------------------------

func TestProtectedRouteRejectsMissingToken(t *testing.T) {
	router := newTestRouter(&stubAuthService{}, &stubFamilyService{}, &stubRelationService{})

	rec := doRequest(router, http.MethodGet, "/api/v1/auth/me", "", "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestMeReturnsAuthenticatedUser(t *testing.T) {
	auth := &stubAuthService{meUser: domain.User{ID: 7, Name: "Fabien", Email: "fabien@example.com"}}
	router := newTestRouter(auth, &stubFamilyService{}, &stubRelationService{})

	rec := doRequest(router, http.MethodGet, "/api/v1/auth/me", "", "any-token")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
	}
	body := decodeBody(t, rec)
	user := body["user"].(map[string]any)
	if user["id"].(float64) != 7 {
		t.Fatalf("user id = %v, want 7", user["id"])
	}
}

func TestMeWhenNoFamilyReturnsNotFound(t *testing.T) {
	families := &stubFamilyService{meErr: domain.ErrFamilyNotFound}
	router := newTestRouter(&stubAuthService{}, families, &stubRelationService{})

	rec := doRequest(router, http.MethodGet, "/api/v1/families/me", "", "any-token")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestCreateFamilyReturnsCreated(t *testing.T) {
	families := &stubFamilyService{created: domain.Family{ID: 1, Name: "The Smiths", InviteCode: "ABC123"}}
	router := newTestRouter(&stubAuthService{}, families, &stubRelationService{})

	rec := doRequest(router, http.MethodPost, "/api/v1/families",
		`{"name":"The Smiths"}`, "any-token")

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestCreateRelationCrossFamilyReturnsForbidden(t *testing.T) {
	families := &stubFamilyService{meFamily: domain.Family{ID: 10, Name: "A"}}
	relations := &stubRelationService{createErr: domain.ErrNotInSameFamily}
	router := newTestRouter(&stubAuthService{}, families, relations)

	rec := doRequest(router, http.MethodPost, "/api/v1/relations",
		`{"related_user_id":2,"type":"parent"}`, "any-token")

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 (body: %s)", rec.Code, rec.Body.String())
	}
}

func TestListRelationsReturnsRelations(t *testing.T) {
	families := &stubFamilyService{meFamily: domain.Family{ID: 10, Name: "A"}}
	relations := &stubRelationService{
		listed: []domain.Relation{{ID: 1, FamilyID: 10, UserID: 7, RelatedUserID: 8, Type: "spouse"}},
	}
	router := newTestRouter(&stubAuthService{}, families, relations)

	rec := doRequest(router, http.MethodGet, "/api/v1/relations", "", "any-token")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
	}
	body := decodeBody(t, rec)
	relationsList := body["relations"].([]any)
	if len(relationsList) != 1 {
		t.Fatalf("relations length = %d, want 1", len(relationsList))
	}
}

// --- error mapping -----------------------------------------------------------

func TestErrorMappingUnknownErrorIsInternal(t *testing.T) {
	auth := &stubAuthService{registerErr: errors.New("boom")}
	router := newTestRouter(auth, &stubFamilyService{}, &stubRelationService{})

	rec := doRequest(router, http.MethodPost, "/api/v1/auth/register",
		`{"name":"Fabien","email":"fabien@example.com","password":"correct-password"}`, "")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500 (body: %s)", rec.Code, rec.Body.String())
	}
	if body := decodeBody(t, rec); body["error"] != "internal server error" {
		t.Fatalf("error = %v, want generic internal message (leak check)", body["error"])
	}
}

var _ token.Manager = (*stubTokens)(nil)
