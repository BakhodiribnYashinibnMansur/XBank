package middleware

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// DynamicIPWhitelist - loads allowed IPs from DB, refreshes every interval
type DynamicIPWhitelist struct {
	pool     *pgxpool.Pool
	mu       sync.RWMutex
	allowed  []string // cached IPs from DB
	interval time.Duration
}

func NewDynamicIPWhitelist(pool *pgxpool.Pool, refreshInterval time.Duration) *DynamicIPWhitelist {
	w := &DynamicIPWhitelist{
		pool:     pool,
		interval: refreshInterval,
	}
	w.refresh()
	if refreshInterval > 0 {
		go w.startRefreshLoop()
	}
	return w
}

func (w *DynamicIPWhitelist) refresh() {
	if w.pool == nil {
		return
	}
	rows, err := w.pool.Query(context.Background(), `SELECT ip_address FROM admin_whitelist_ips`)
	if err != nil {
		logger.Log.Warn("failed to load admin whitelist IPs", zap.Error(err))
		return
	}
	defer rows.Close()

	var ips []string
	for rows.Next() {
		var ip string
		if err := rows.Scan(&ip); err == nil {
			ips = append(ips, ip)
		}
	}

	w.mu.Lock()
	w.allowed = ips
	w.mu.Unlock()
}

func (w *DynamicIPWhitelist) startRefreshLoop() {
	ticker := time.NewTicker(w.interval)
	for range ticker.C {
		w.refresh()
	}
}

// Middleware - Fiber middleware that checks client IP against DB whitelist
func (w *DynamicIPWhitelist) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		w.mu.RLock()
		allowed := w.allowed
		w.mu.RUnlock()

		// If no IPs in DB, allow all (not configured yet)
		if len(allowed) == 0 {
			return c.Next()
		}

		clientIP := net.ParseIP(c.IP())
		if clientIP == nil {
			return apperror.ErrForbidden.WithMessage("Invalid IP address")
		}

		for _, a := range allowed {
			// Check CIDR
			if _, cidr, err := net.ParseCIDR(a); err == nil {
				if cidr.Contains(clientIP) {
					return c.Next()
				}
				continue
			}
			// Check exact IP
			if ip := net.ParseIP(a); ip != nil && ip.Equal(clientIP) {
				return c.Next()
			}
		}

		logger.Log.Warn("admin access denied: IP not whitelisted",
			zap.String("ip", c.IP()),
			zap.String("path", c.Path()),
		)
		return apperror.ErrForbidden.WithMessage("Access denied: IP not whitelisted")
	}
}
