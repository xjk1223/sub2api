package handler

import (
	"errors"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// CheckinHandler handles user-facing check-in requests.
type CheckinHandler struct {
	checkinService *service.CheckinService
}

func NewCheckinHandler(checkinService *service.CheckinService) *CheckinHandler {
	return &CheckinHandler{checkinService: checkinService}
}

// GetCheckinStatus GET /api/v1/user/checkin/status
func (h *CheckinHandler) GetCheckinStatus(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	status, err := h.checkinService.GetStatus(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, status)
}

// DoCheckin POST /api/v1/user/checkin
func (h *CheckinHandler) DoCheckin(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	result, err := h.checkinService.DoCheckin(c.Request.Context(), subject.UserID)
	if err != nil {
		if errors.Is(err, service.ErrCheckinAlreadyDone) {
			response.BadRequest(c, "今日已签到")
			return
		}
		if errors.Is(err, service.ErrCheckinDisabled) {
			response.Forbidden(c, "签到功能未开启")
			return
		}
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

// GetCheckinHistory GET /api/v1/user/checkin/history?page=1&page_size=20
func (h *CheckinHandler) GetCheckinHistory(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	page, pageSize := parseCheckinPagination(c)
	logs, total, err := h.checkinService.ListLogs(c.Request.Context(), subject.UserID, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, logs, total, page, pageSize)
}

func parseCheckinPagination(c *gin.Context) (page, pageSize int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return
}
