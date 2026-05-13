package handler

import (
	"net/http"
	"strconv"

	"github.com/authnas/authnas/go-server/internal/repository"
	"github.com/authnas/authnas/go-server/internal/response"
	"github.com/authnas/authnas/go-server/internal/service"
	"github.com/gin-gonic/gin"
)

type AdminAuditHandler struct {
	auditService *service.AuditService
	auditRepo    *repository.AuditLogRepository
}

func NewAdminAuditHandler(auditService *service.AuditService, auditRepo *repository.AuditLogRepository) *AdminAuditHandler {
	return &AdminAuditHandler{
		auditService: auditService,
		auditRepo:    auditRepo,
	}
}

func (h *AdminAuditHandler) ListAuditLogs(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")
	eventType := c.Query("event_type")
	userID := c.Query("user_id")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	logs, total, err := h.auditRepo.List(offset, pageSize, eventType, userID)
	if err != nil {
		response.InternalServerError(c, "failed to list audit logs")
		return
	}

	c.JSON(http.StatusOK, response.PaginatedResponse{
		Success:  true,
		Data:     logs,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}
