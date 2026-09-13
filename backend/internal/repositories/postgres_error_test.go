package repositories

import (
	"context"
	"errors"
	"testing"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/apperrors"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestClassifyPostgresError(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantCode    apperrors.ErrCode
		wantMessage string
		wantSameErr bool
	}{
		{
			name:        "context canceled",
			err:         context.Canceled,
			wantSameErr: true,
		},
		{
			name:        "context deadline exceeded",
			err:         context.DeadlineExceeded,
			wantCode:    apperrors.DependencyUnavailable,
			wantMessage: "temporarily unavailable",
		},
		{
			name:        "database error",
			err:         errors.New("database unavailable"),
			wantCode:    apperrors.DependencyUnavailable,
			wantMessage: "temporarily unavailable",
		},
		{
			name:        "foreign key violation",
			err:         &pgconn.PgError{Code: "23503"},
			wantCode:    apperrors.NotFound,
			wantMessage: "referenced resource not found",
		},
		{
			name:        "unique violation",
			err:         &pgconn.PgError{Code: "23505"},
			wantCode:    apperrors.BadParam,
			wantMessage: "resource already exists",
		},
		{
			name:        "not null violation",
			err:         &pgconn.PgError{Code: "23502"},
			wantCode:    apperrors.BadParam,
			wantMessage: "required value is missing",
		},
		{
			name:        "check violation",
			err:         &pgconn.PgError{Code: "23514"},
			wantCode:    apperrors.BadParam,
			wantMessage: "value violates constraint",
		},
		{
			name:        "string data too long",
			err:         &pgconn.PgError{Code: "22001"},
			wantCode:    apperrors.BadParam,
			wantMessage: "invalid value",
		},
		{
			name:        "invalid text representation",
			err:         &pgconn.PgError{Code: "22P02"},
			wantCode:    apperrors.BadParam,
			wantMessage: "invalid value",
		},
		{
			name:        "undefined column",
			err:         &pgconn.PgError{Code: "42703"},
			wantCode:    apperrors.DataMappingFailed,
			wantMessage: "database schema is incompatible",
		},
		{
			name:        "undefined table",
			err:         &pgconn.PgError{Code: "42P01"},
			wantCode:    apperrors.DataMappingFailed,
			wantMessage: "database schema is incompatible",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyPostgresError(tt.err)
			if tt.wantSameErr {
				if !errors.Is(got, tt.err) {
					t.Fatalf("expected original error, got %v", got)
				}
				return
			}

			var appErr *apperrors.Error
			if !errors.As(got, &appErr) || appErr.ErrCode != string(tt.wantCode) {
				t.Fatalf("expected application error %q, got %v", tt.wantCode, got)
			}
			if appErr.Message != tt.wantMessage {
				t.Fatalf("expected message %q, got %q", tt.wantMessage, appErr.Message)
			}
		})
	}
}
