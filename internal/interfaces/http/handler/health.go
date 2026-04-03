package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// HealthHandler provides liveness and readiness probes.
type HealthHandler struct {
	dbPool      *pgxpool.Pool
	mongoClient *mongo.Client
	kafkaBroker string
	redisClient RedisClient
}

// RedisClient is the subset of redis.Client used for health checks.
type RedisClient interface {
	Ping(ctx context.Context) *goredis.StatusCmd
}

// NewHealthHandler creates a new HealthHandler.
// mongoClient, redisClient and kafkaBroker may be nil/empty if not configured.
func NewHealthHandler(dbPool *pgxpool.Pool, mongoClient *mongo.Client, kafkaBroker string, redisClient RedisClient) *HealthHandler {
	return &HealthHandler{
		dbPool:      dbPool,
		mongoClient: mongoClient,
		kafkaBroker: kafkaBroker,
		redisClient: redisClient,
	}
}

// Live godoc
// @Summary      Liveness probe
// @Description  Returns 200 if the process is alive
// @Tags         Health
// @Produce      json
// @Success      200 {object} map[string]string
// @Router       /health [get]
func (h *HealthHandler) Live(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

// Ready godoc
// @Summary      Readiness probe
// @Description  Checks all dependencies (PostgreSQL, MongoDB, Kafka)
// @Tags         Health
// @Produce      json
// @Success      200 {object} map[string]interface{}
// @Failure      503 {object} map[string]interface{}
// @Router       /health/ready [get]
func (h *HealthHandler) Ready(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()

	checks := fiber.Map{}
	allOk := true

	// PostgreSQL
	if err := h.dbPool.Ping(ctx); err != nil {
		checks["postgres"] = "down"
		allOk = false
	} else {
		checks["postgres"] = "up"
	}

	// MongoDB
	if h.mongoClient != nil {
		if err := h.mongoClient.Ping(ctx, readpref.Primary()); err != nil {
			checks["mongodb"] = "down"
			allOk = false
		} else {
			checks["mongodb"] = "up"
		}
	}

	// Redis
	if h.redisClient != nil {
		if err := h.redisClient.Ping(ctx).Err(); err != nil {
			checks["redis"] = "down"
			allOk = false
		} else {
			checks["redis"] = "up"
		}
	}

	// Kafka
	if h.kafkaBroker != "" {
		conn, err := kafka.DialContext(ctx, "tcp", h.kafkaBroker)
		if err != nil {
			checks["kafka"] = "down"
			allOk = false
		} else {
			conn.Close()
			checks["kafka"] = "up"
		}
	}

	status := "ready"
	httpStatus := http.StatusOK
	if !allOk {
		status = "not_ready"
		httpStatus = http.StatusServiceUnavailable
	}

	return c.Status(httpStatus).JSON(fiber.Map{
		"status": status,
		"checks": checks,
	})
}
