package statistics

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/generic/statistics/application/query"
	httpHandler "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/generic/statistics/interfaces/http"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BoundedContext struct {
	Handler *httpHandler.Handler
}

func NewBoundedContext(pool *pgxpool.Pool) *BoundedContext {
	return &BoundedContext{
		Handler: httpHandler.NewHandler(query.NewOverviewHandler(pool)),
	}
}
