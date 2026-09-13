package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindOrCreateByOIDC(ctx context.Context, subject, email, displayName string) (models.User, error) {
	handle := handleFromOIDC(subject, email)
	if displayName == "" {
		displayName = handle
	}

	var user models.User
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO users (id, oidc_subject, email, handle, display_name, bio, profile_completed)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, '', FALSE)
		ON CONFLICT (oidc_subject) DO UPDATE
		SET email = EXCLUDED.email,
		    display_name = EXCLUDED.display_name
		RETURNING id, handle, display_name, bio, created_at, NOT profile_completed
	`, subject, email, handle, displayName).Scan(
		&user.ID,
		&user.Handle,
		&user.DisplayName,
		&user.Bio,
		&user.CreatedAt,
		&user.NeedsProfileSetup,
	)
	if err != nil {
		return models.User{}, classifyPostgresError(err)
	}
	return user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (models.User, error) {
	var user models.User
	err := r.db.QueryRowContext(ctx, `
		SELECT id, handle, display_name, bio, created_at, NOT profile_completed
		FROM users
		WHERE id = $1
	`, id).Scan(&user.ID, &user.Handle, &user.DisplayName, &user.Bio, &user.CreatedAt, &user.NeedsProfileSetup)
	if err != nil {
		return models.User{}, classifyPostgresError(err)
	}
	return user, nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, id string, request models.UpdateUserProfileRequest) (models.User, error) {
	var user models.User
	err := r.db.QueryRowContext(ctx, `
		UPDATE users
		SET handle = $2,
		    display_name = $3,
		    bio = $4,
		    profile_completed = TRUE
		WHERE id = $1
		RETURNING id, handle, display_name, bio, created_at, NOT profile_completed
	`, id, request.Handle, request.DisplayName, request.Bio).Scan(
		&user.ID,
		&user.Handle,
		&user.DisplayName,
		&user.Bio,
		&user.CreatedAt,
		&user.NeedsProfileSetup,
	)
	if err != nil {
		return models.User{}, classifyPostgresError(err)
	}
	return user, nil
}

func handleFromOIDC(subject, email string) string {
	base := strings.Split(email, "@")[0]
	base = strings.ToLower(base)
	cleaned := make([]rune, 0, len(base))
	for _, r := range base {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			cleaned = append(cleaned, r)
		}
	}
	if len(cleaned) == 0 {
		cleaned = []rune("user")
	}
	suffix := subject
	if len(suffix) > 8 {
		suffix = suffix[:8]
	}
	return fmt.Sprintf("user_%s", suffix)[:13]
}
