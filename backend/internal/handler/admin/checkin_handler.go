package admin

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// CheckinHandler handles admin check-in management.
type CheckinHandler struct {
	checkinService *service.CheckinService
}

func NewCheckinHandler(checkinService *service.CheckinService) *CheckinHandler {
	return &CheckinHandler{checkinService: checkinService}
}

// GetSettings GET /api/v1/admin/checkin/settings
func (h *CheckinHandler) GetSettings(c *gin.Context) {
	settings, err := h.checkinService.GetSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, settings)
}

// UpdateCheckinSettingsRequest is the admin settings update payload.
type UpdateCheckinSettingsRequest struct {
	Enabled          bool    `json:"enabled"`
	BaseAmount       float64 `json:"base_amount" binding:"min=0"`
	ConsecutiveBonus bool    `json:"consecutive_bonus"`
	BonusPerDay      float64 `json:"bonus_per_day" binding:"min=0"`
	MaxBonusDays     int     `json:"max_bonus_days" binding:"min=0"`
}

// UpdateSettings PUT /api/v1/admin/checkin/settings
func (h *CheckinHandler) UpdateSettings(c *gin.Context) {
	var req UpdateCheckinSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	settings := &service.CheckinSettings{
		Enabled:          req.Enabled,
		BaseAmount:       req.BaseAmount,
		ConsecutiveBonus: req.ConsecutiveBonus,
		BonusPerDay:      req.BonusPerDay,
		MaxBonusDays:     req.MaxBonusDays,
	}
	if err := h.checkinService.UpdateSettings(c.Request.Context(), settings); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, settings)
}

// ListLogs GET /api/v1/admin/checkin/logs?user_id=&page=1&page_size=20
func (h *CheckinHandler) ListLogs(c *gin.Context) {
	var userID int64
	if s := c.Query("user_id"); s != "" {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			response.BadRequest(c, "Invalid user_id")
			return
		}
		userID = v
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	logs, total, err := h.checkinService.ListLogs(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, logs, total, page, pageSize)
}
