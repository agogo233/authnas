package handler

import (
	"errors"
	"strings"

	"github.com/authnas/authnas/go-server/internal/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GroupListResponse struct {
	Groups []GroupListItem `json:"groups"`
	Total  int64           `json:"total"`
}

type GroupListItem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	CreatedAt   string  `json:"createdAt"`
}

type CreateGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *AdminHandler) ListGroups(c *gin.Context) {
	groups, total, err := h.groupService.List(0, 100)
	if err != nil {
		response.InternalServerError(c, "failed to list groups")
		return
	}

	var items []GroupListItem
	for _, g := range groups {
		items = append(items, GroupListItem{
			ID:          g.ID,
			Name:        g.Name,
			Description: g.Description,
			CreatedAt:   g.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	response.SuccessList(c, items, total)
}

func (h *AdminHandler) GetGroup(c *gin.Context) {
	id := c.Param("id")

	group, err := h.groupService.GetByID(id)
	if err != nil || group == nil {
		response.NotFound(c, "group not found")
		return
	}

	response.Success(c, GroupListItem{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		CreatedAt:   group.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *AdminHandler) CreateGroup(c *gin.Context) {
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	group, err := h.groupService.Create(req.Name, req.Description)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) || (err.Error() != "" && strings.Contains(err.Error(), "UNIQUE")) {
			response.BadRequest(c, "group with this name already exists")
			return
		}
		response.InternalServerError(c, "failed to create group")
		return
	}

	response.Success(c, GroupListItem{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		CreatedAt:   group.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *AdminHandler) UpdateGroup(c *gin.Context) {
	id := c.Param("id")

	var req UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	existingGroup, err := h.groupService.GetByID(id)
	if err != nil || existingGroup == nil {
		response.NotFound(c, "group not found")
		return
	}

	group, err := h.groupService.Update(id, req.Name, req.Description)
	if err != nil {
		response.InternalServerError(c, "failed to update group")
		return
	}

	response.Success(c, GroupListItem{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		CreatedAt:   group.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *AdminHandler) DeleteGroup(c *gin.Context) {
	id := c.Param("id")

	existingGroup, err := h.groupService.GetByID(id)
	if err != nil || existingGroup == nil {
		response.NotFound(c, "group not found")
		return
	}

	err = h.groupService.Delete(id)
	if err != nil {
		response.InternalServerError(c, "failed to delete group")
		return
	}

	response.SuccessWithMessage(c, "group deleted successfully")
}
