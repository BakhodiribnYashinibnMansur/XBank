package mongodb

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/shared"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// AuditLog - MongoDB implementation of shared.AuditLog
// Writes audit entries async (non-blocking goroutine)
type AuditLog struct {
	collection *mongo.Collection
}

// NewAuditLog - create audit log writer
func NewAuditLog(db *mongo.Database) *AuditLog {
	return &AuditLog{
		collection: db.Collection("audit_log"),
	}
}

// Log - write audit entry to MongoDB (async, non-blocking)
func (a *AuditLog) Log(ctx context.Context, entry shared.AuditEntry) error {
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

		_, err := a.collection.InsertOne(context.Background(), doc)
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
