package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/Fabien-Halaby/kinspace/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// testDatabaseURL returns the DSN for integration tests. Set
// TEST_DATABASE_URL to run them (see .github/workflows/backend.yml).
func testDatabaseURL(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}
	return dsn
}

// setupTestDB connects to the test database and resets it to a clean
// migrated state for every test.
func setupTestDB(t *testing.T) *testDB {
	t.Helper()
	ctx := context.Background()

	pool, err := Connect(ctx, testDatabaseURL(t))
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)

	if _, err := pool.Exec(ctx, `
		DROP TABLE IF EXISTS relations, families, users, schema_migrations CASCADE
	`); err != nil {
		t.Fatalf("reset schema: %v", err)
	}
	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	return &testDB{pool: pool}
}

type testDB struct {
	pool *pgxpool.Pool
}

func TestUserRepositoryRoundTrip(t *testing.T) {
	db := setupTestDB(t)
	users := NewUserRepository(db.pool)

	ctx := context.Background()
	created, err := users.Create(ctx, "Fabien", "fabien@example.com", "a-hash")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID == 0 || created.Name != "Fabien" {
		t.Fatalf("created = %+v, want populated user", created)
	}

	found, err := users.FindByEmail(ctx, "fabien@example.com")
	if err != nil {
		t.Fatalf("FindByEmail() error = %v", err)
	}
	if found.ID != created.ID || found.PasswordHash != "a-hash" {
		t.Fatalf("found = %+v, want stored user", found)
	}

	byID, err := users.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if byID.Email != "fabien@example.com" {
		t.Fatalf("byID.Email = %q, want fabien@example.com", byID.Email)
	}
}

func TestUserRepositoryCreateDuplicateEmail(t *testing.T) {
	db := setupTestDB(t)
	users := NewUserRepository(db.pool)
	ctx := context.Background()

	if _, err := users.Create(ctx, "One", "dup@example.com", "h1"); err != nil {
		t.Fatalf("first Create() error = %v", err)
	}
	if _, err := users.Create(ctx, "Two", "dup@example.com", "h2"); !errors.Is(err, domain.ErrEmailExists) {
		t.Fatalf("second Create() error = %v, want ErrEmailExists", err)
	}
}

func TestUserRepositoryFindByEmailUnknown(t *testing.T) {
	db := setupTestDB(t)
	users := NewUserRepository(db.pool)

	_, err := users.FindByEmail(context.Background(), "nobody@example.com")
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("error = %v, want ErrUserNotFound", err)
	}
}

func TestFamilyRepositoryCreateAndJoin(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	users := NewUserRepository(db.pool)
	families := NewFamilyRepository(db.pool)

	owner, err := users.Create(ctx, "Owner", "owner@example.com", "h")
	if err != nil {
		t.Fatalf("create owner: %v", err)
	}
	joiner, err := users.Create(ctx, "Joiner", "joiner@example.com", "h")
	if err != nil {
		t.Fatalf("create joiner: %v", err)
	}

	family, err := families.Create(ctx, "The Smiths", "INVITE01", owner.ID)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	me, err := families.FindByUserID(ctx, owner.ID)
	if err != nil {
		t.Fatalf("FindByUserID(owner) error = %v", err)
	}
	if me.ID != family.ID {
		t.Fatalf("family id = %d, want %d", me.ID, family.ID)
	}

	joined, err := families.Join(ctx, joiner.ID, "INVITE01")
	if err != nil {
		t.Fatalf("Join() error = %v", err)
	}
	if joined.ID != family.ID {
		t.Fatalf("joined id = %d, want %d", joined.ID, family.ID)
	}
}

func TestFamilyRepositoryCreateRejectsSecondMembership(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	users := NewUserRepository(db.pool)
	families := NewFamilyRepository(db.pool)

	user, err := users.Create(ctx, "User", "user@example.com", "h")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if _, err := families.Create(ctx, "One", "AAA111", user.ID); err != nil {
		t.Fatalf("first Create() error = %v", err)
	}
	if _, err := families.Create(ctx, "Two", "BBB222", user.ID); !errors.Is(err, domain.ErrAlreadyInFamily) {
		t.Fatalf("second Create() error = %v, want ErrAlreadyInFamily", err)
	}
}

func TestFamilyRepositoryJoinInvalidCode(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	users := NewUserRepository(db.pool)
	families := NewFamilyRepository(db.pool)

	user, err := users.Create(ctx, "User", "user2@example.com", "h")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if _, err := families.Join(ctx, user.ID, "NOPE"); !errors.Is(err, domain.ErrInviteCodeInvalid) {
		t.Fatalf("error = %v, want ErrInviteCodeInvalid", err)
	}
}

func TestRelationRepositoryScopedToFamily(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	users := NewUserRepository(db.pool)
	families := NewFamilyRepository(db.pool)
	relations := NewRelationRepository(db.pool)

	members := make([]domain.User, 3)
	for i := 0; i < 3; i++ {
		user, err := users.Create(ctx, "Member", fmt.Sprintf("member%d@example.com", i), "h")
		if err != nil {
			t.Fatalf("create member %d: %v", i, err)
		}
		members[i] = user
	}

	family, err := families.Create(ctx, "The Smiths", "REL001", members[0].ID)
	if err != nil {
		t.Fatalf("create family: %v", err)
	}
	for _, member := range members[1:] {
		if _, err := families.Join(ctx, member.ID, "REL001"); err != nil {
			t.Fatalf("join member: %v", err)
		}
	}

	relation, err := relations.Create(ctx, family.ID, members[0].ID, members[1].ID, "parent")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if relation.Type != "parent" || relation.FamilyID != family.ID {
		t.Fatalf("relation = %+v, want scoped parent edge", relation)
	}

	listed, err := relations.List(ctx, family.ID)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(listed) != 1 || listed[0].RelatedUserID != members[1].ID {
		t.Fatalf("listed = %+v, want single relation", listed)
	}

	// A relation cannot be created across families: user in a second
	// family must not be linkable.
	other, err := users.Create(ctx, "Outsider", "outsider@example.com", "h")
	if err != nil {
		t.Fatalf("create outsider: %v", err)
	}
	if _, err := families.Create(ctx, "Other", "REL002", other.ID); err != nil {
		t.Fatalf("create other family: %v", err)
	}
	otherFamilyID, err := relations.UserFamilyID(ctx, other.ID)
	if err != nil {
		t.Fatalf("UserFamilyID() error = %v", err)
	}
	if otherFamilyID == family.ID {
		t.Fatal("outsider must belong to a different family")
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	if err := Migrate(ctx, db.pool); err != nil {
		t.Fatalf("second Migrate() error = %v", err)
	}
}
