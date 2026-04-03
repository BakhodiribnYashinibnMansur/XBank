package apperror

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// MapRepoError converts common database errors to AppError with source context.
// Use this in repository implementations to standardize error handling:
//
//	if err != nil {
//	    return apperror.MapRepoError(err, "UserRepo.GetByID")
//	}
func MapRepoError(err error, operation string) *AppError {
	if err == nil {
		return nil
	}

	// No rows → not found
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound.Wrap(err).WithSource(operation)
	}

	// PostgreSQL-specific errors
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return ErrConcurrencyConflict.Wrap(err).WithSource(operation).
				WithDetails(pgErr.Detail)
		case "23503": // foreign_key_violation
			return ErrBadRequest.Wrap(err).WithSource(operation).
				WithDetails("referenced resource does not exist")
		case "23514": // check_violation
			return ErrValidation.Wrap(err).WithSource(operation).
				WithDetails(pgErr.Message)
		case "40001": // serialization_failure
			return ErrConcurrencyConflict.Wrap(err).WithSource(operation).
				WithDetails("serialization conflict, please retry")
		case "57014": // query_canceled
			return ErrDBTimeout.Wrap(err).WithSource(operation)
		}
	}

	// Context errors
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrDBTimeout.Wrap(err).WithSource(operation)
	}
	if errors.Is(err, context.Canceled) {
		return ErrDatabase.Wrap(err).WithSource(operation).
			WithDetails("request canceled")
	}

	// Fallback: unknown database error
	return ErrDatabase.Wrap(err).WithSource(operation).WithSeverity(SeverityHigh)
}
