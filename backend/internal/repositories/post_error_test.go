package repositories

import (
	"context"
	"errors"
	"testing"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/apperrors"
)

func TestClassifyPostgresError(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantCode    apperrors.ErrCode
		wantSameErr bool
	}{
		{
			name:        "context canceled",
			err:         context.Canceled,
			wantSameErr: true,
		},
		{
			name:     "context deadline exceeded",
			err:      context.DeadlineExceeded,
			wantCode: apperrors.DependencyUnavailable,
		},
		{
			name:     "database error",
			err:      errors.New("database unavailable"),
			wantCode: apperrors.DependencyUnavailable,
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
		})
	}
}
