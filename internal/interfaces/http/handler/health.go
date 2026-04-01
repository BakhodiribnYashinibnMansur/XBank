package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// HealthHandler provides liveness and readiness probes.
type HealthHandler struct {
	dbPool      *pgxpool.Pool
	mongoClient *mongo.Client
	kafkaBroker string
}

// NewHealthHandler creates a new HealthHandler.
// mongoClient and kafkaBroker may be nil/empty if not configured.
func NewHealthHandler(dbPool *pgxpool.Pool, mongoClient *mongo.Client, kafkaBroker string) *HealthHandler {
	return &HealthHandler{
		dbPool:      dbPool,
		mongoClient: mongoClient,
		kafkaBroker: kafkaBroker,
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
