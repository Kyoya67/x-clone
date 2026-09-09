package repositories

import (
	"context"
	"errors"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/apperrors"
)

func classifyPostgresError(err error) error {
	if errors.Is(err, context.Canceled) {
		return err
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return apperrors.DependencyUnavailable.Wrap(err, "temporarily unavailable")
	}
	return apperrors.DependencyUnavailable.Wrap(err, "temporarily unavailable")
}
