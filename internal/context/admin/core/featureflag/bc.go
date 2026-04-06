// Package featureflag provides the Feature Flag Bounded Context.
// Supports feature toggles with rollout percentage, rule-based evaluation,
// and LRU-cached lookups with Pub/Sub invalidation.
package featureflag

import (
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/featureflag/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/featureflag/application/query"
	flagCache "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/featureflag/infrastructure/cache"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/featureflag/infrastructure/postgres"
	httpHandler "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/featureflag/interfaces/http"
	appKernel "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/application"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires all Feature Flag components.
type BoundedContext struct {
	Handler   *httpHandler.Handler
	Evaluator *flagCache.CachedEvaluator
}

// NewBoundedContext creates the Feature Flag BC with all dependencies.
func NewBoundedContext(pool *pgxpool.Pool, eventBus appKernel.EventBus) *BoundedContext {
	writeRepo := postgres.NewWriteRepo(pool)
	readRepo := postgres.NewReadRepo(pool)

	createHandler := command.NewCreateHandler(writeRepo, eventBus)
	updateHandler := command.NewUpdateHandler(writeRepo, eventBus)
	deleteHandler := command.NewDeleteHandler(writeRepo, eventBus)
	getHandler := query.NewGetHandler(readRepo)
	listHandler := query.NewListHandler(readRepo)
	evaluateHandler := query.NewEvaluateHandler(writeRepo)

	handler := httpHandler.NewHandler(createHandler, updateHandler, deleteHandler, getHandler, listHandler, evaluateHandler)

	evaluator := flagCache.NewCachedEvaluator(writeRepo, 256, 5*time.Minute)

	return &BoundedContext{
		Handler:   handler,
		Evaluator: evaluator,
	}
}
