package mongodb

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/circuitbreaker"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// AuditLog - MongoDB implementation of domain.AuditLog
// Writes audit entries async (non-blocking goroutine)
// Protected by circuit breaker to prevent cascading failures when MongoDB is down.
type AuditLog struct {
	collection *mongo.Collection
	breaker    *circuitbreaker.Breaker
}

// NewAuditLog - create audit log writer
func NewAuditLog(db *mongo.Database) *AuditLog {
	return &AuditLog{
		collection: db.Collection("audit_log"),
		breaker:    circuitbreaker.New("mongodb-audit", 5, 30*time.Second),
	}
}

// Log - write audit entry to MongoDB (async, non-blocking)
func (a *AuditLog) Log(ctx context.Context, entry domain.AuditEntry) error {
	// Fire and forget - don't block the main request
	go func() {
		doc := bson.M{
			"aggregate_type": entry.AggregateType,
			"aggregate_id":   entry.AggregateID,
			"action":         entry.Action,
			"actor_id":       entry.ActorID,
			"attributes":     entry.Attributes,
			"ip_address":     entry.IPAddress,
			"user_agent":     entry.UserAgent,
			"timestamp":      entry.Timestamp,
			"created_at":     time.Now(),
		}

		err := a.breaker.Execute(func() error {
			_, insertErr := a.collection.InsertOne(context.Background(), doc)
			return insertErr
		})
		if err != nil {
			logger.Log.Error("audit log write failed",
				zap.String("aggregate_type", entry.AggregateType),
				zap.String("aggregate_id", entry.AggregateID),
				zap.String("action", entry.Action),
				zap.Error(err),
			)
		}
	}()

	return nil
}
