package data

import (
	"context"
	"time"

	"github.com/Blue-Davinci/musical-zoe/internal/validator"
)

// contextGenerator() is a helper function that generates a new context.Context from a
// context.Context and a timeout duration.
func contextGenerator(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}

func ValidateURLID(v *validator.Validator, stockID int64, fieldName string) {
	v.Check(stockID > 0, fieldName, "must be a valid ID")
}
