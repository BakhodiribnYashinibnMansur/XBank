package http

import (
	"net/http"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/dashboard/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

// Handler handles HTTP requests for admin dashboard.
type Handler struct {
	service *command.Service
}

// NewHandler creates a new dashboard HTTP handler.
func NewHandler(service *command.Service) *Handler {
	return &Handler{service: service}
}

// Overview godoc
// @Summary      Get admin dashboard overview statistics
// @Tags         AdminDashboard
// @Produce      json
// @Success      200 {object} OverviewResponse
// @Failure      500 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/dashboard/overview [get]
func (h *Handler) Overview(c *fiber.Ctx) error {
	stats, err := h.service.GetOverview(c.Context())
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, OverviewResponse{
		TotalUsers:       stats.TotalUsers,
		ActiveUsers:      stats.ActiveUsers,
		TotalAccounts:    stats.TotalAccounts,
		TotalTransfers:   stats.TotalTransfers,
		TotalDeposits:    stats.TotalDeposits,
		TotalWithdrawals: stats.TotalWithdrawals,
		GeneratedAt:      stats.GeneratedAt,
	})
}

// PeriodStats godoc
// @Summary      Get dashboard stats for a time period
// @Tags         AdminDashboard
// @Produce      json
// @Param        period query string true "Period: daily, weekly, monthly"
// @Param        start  query string true "Start date (RFC3339)"
// @Param        end    query string true "End date (RFC3339)"
// @Success      200 {object} PeriodStatsResponse
// @Failure      400 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/dashboard/period [get]
func (h *Handler) PeriodStats(c *fiber.Ctx) error {
	period := c.Query("period")
	if period != "daily" && period != "weekly" && period != "monthly" {
		return apperror.ErrValidation.WithMessage("period must be daily, weekly, or monthly")
	}

	startStr := c.Query("start")
	endStr := c.Query("end")
	if startStr == "" || endStr == "" {
		return apperror.ErrMissingField.WithMessage("start and end query parameters are required")
	}

	startDate, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		return apperror.ErrValidation.WithMessage("start must be a valid RFC3339 date")
	}
	endDate, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		return apperror.ErrValidation.WithMessage("end must be a valid RFC3339 date")
	}

	stats, err := h.service.GetPeriodStats(c.Context(), period, startDate, endDate)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, PeriodStatsResponse{
		Period:       stats.Period,
		StartDate:    stats.StartDate,
		EndDate:      stats.EndDate,
		NewUsers:     stats.NewUsers,
		NewAccounts:  stats.NewAccounts,
		Transactions: stats.Transactions,
		Volume:       stats.Volume,
	})
}
