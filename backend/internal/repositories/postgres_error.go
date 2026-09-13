package repositories

import (
	"context"
	"errors"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/apperrors"
	"github.com/jackc/pgx/v5/pgconn"
)

func classifyPostgresError(err error) error {
	if errors.Is(err, context.Canceled) {
		return err
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return apperrors.DependencyUnavailable.Wrap(err, "temporarily unavailable")
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503":
			return apperrors.NotFound.Wrap(err, "referenced resource not found")
		case "23505":
			return apperrors.BadParam.Wrap(err, "resource already exists")
		case "23502":
			return apperrors.BadParam.Wrap(err, "required value is missing")
		case "23514":
			return apperrors.BadParam.Wrap(err, "value violates constraint")
		case "22001", "22P02":
			return apperrors.BadParam.Wrap(err, "invalid value")
		case "42703", "42P01":
			return apperrors.DataMappingFailed.Wrap(err, "database schema is incompatible")
		}
	}
	return apperrors.DependencyUnavailable.Wrap(err, "temporarily unavailable")
}
