package mongodb

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AuditReader - reads audit logs from MongoDB for admin API
type AuditReader struct {
	collection *mongo.Collection
}

func NewAuditReader(db *mongo.Database) *AuditReader {
	return &AuditReader{collection: db.Collection("audit_log")}
}

// AuditFilter - query parameters for filtering audit logs
type AuditFilter struct {
	AggregateType string
	AggregateID   string
	Action        string
	ActorID       string
	From          *time.Time
	To            *time.Time
	Limit         int64
	Offset        int64
}

// List - retrieve audit entries with filtering and pagination
func (r *AuditReader) List(ctx context.Context, filter AuditFilter) ([]domain.AuditEntry, int64, error) {
	bsonFilter := bson.M{}
	if filter.AggregateType != "" {
		bsonFilter["aggregate_type"] = filter.AggregateType
	}
	if filter.AggregateID != "" {
		bsonFilter["aggregate_id"] = filter.AggregateID
	}
	if filter.Action != "" {
		bsonFilter["action"] = filter.Action
	}
	if filter.ActorID != "" {
		bsonFilter["actor_id"] = filter.ActorID
	}
	if filter.From != nil || filter.To != nil {
		timeFilter := bson.M{}
		if filter.From != nil {
			timeFilter["$gte"] = *filter.From
		}
		if filter.To != nil {
			timeFilter["$lte"] = *filter.To
		}
		bsonFilter["timestamp"] = timeFilter
	}

	// Count
	total, err := r.collection.CountDocuments(ctx, bsonFilter)
	if err != nil {
		return nil, 0, err
	}

	// Query
	opts := options.Find().
		SetSort(bson.M{"timestamp": -1}).
		SetLimit(filter.Limit).
		SetSkip(filter.Offset)

	cursor, err := r.collection.Find(ctx, bsonFilter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var entries []domain.AuditEntry
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		entry := domain.AuditEntry{
			AggregateType: getString(doc, "aggregate_type"),
			AggregateID:   getString(doc, "aggregate_id"),
			Action:        getString(doc, "action"),
			ActorID:       getString(doc, "actor_id"),
			IPAddress:     getString(doc, "ip_address"),
			UserAgent:     getString(doc, "user_agent"),
		}
		if ts, ok := doc["timestamp"].(time.Time); ok {
			entry.Timestamp = ts
		}
		if attrs, ok := doc["attributes"].(bson.M); ok {
			entry.Attributes = make(map[string]string)
			for k, v := range attrs {
				if s, ok := v.(string); ok {
					entry.Attributes[k] = s
				}
			}
		}
		entries = append(entries, entry)
	}

	return entries, total, nil
}

func getString(doc bson.M, key string) string {
	if v, ok := doc[key].(string); ok {
		return v
	}
	return ""
}
