package handler

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/authnas/authnas/go-server/internal/middleware"
	"github.com/authnas/authnas/go-server/internal/model"
	"github.com/authnas/authnas/go-server/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/nbutton23/zxcvbn-go"
)

var adminUsernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

type UserListResponse struct {
	Users []UserListItem `json:"users"`
	Total int64          `json:"total"`
}

type UserListItem struct {
	ID            string  `json:"id"`
	Email         *string `json:"email"`
	Username      string  `json:"username"`
	Name          *string `json:"name"`
	EmailVerified bool    `json:"emailVerified"`
	Approved      bool    `json:"approved"`
	MFARequired   bool    `json:"mfaRequired"`
	IsAdmin       bool    `json:"isAdmin"`
	CreatedAt     string  `json:"createdAt"`
}

type CreateUserRequest struct {
	Email       string  `json:"email" binding:"required"`
	Username    string  `json:"username" binding:"required"`
	Password    *string `json:"password"`
	Name        string  `json:"name"`
	IsAdmin     bool    `json:"isAdmin"`
	Approved    bool    `json:"approved"`
	MFARequired bool    `json:"mfaRequired"`
}

type UpdateUserRequest struct {
	Email       *string `json:"email"`
	Username    *string `json:"username"`
	Name        *string `json:"name"`
	Approved    *bool   `json:"approved"`
	IsAdmin     *bool   `json:"isAdmin"`
	MFARequired *bool   `json:"mfaRequired"`
}

type ApproveUserRequest struct {
	Approved bool `json:"approved"`
}

type ResetPasswordRequest struct {
	NewPassword string `json:"newPassword" binding:"required"`
}

func (h *AdminHandler) CountUsers(c *gin.Context) {
	total, err := h.userService.Count()
	if err != nil {
		response.InternalServerError(c, "failed to count users")
		return
	}
	response.Success(c, gin.H{"total": total})
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	page := 1
	pageSize := 100
	searchQuery := c.Query("search")

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 1000 {
			pageSize = parsed
		}
	}

	offset := (page - 1) * pageSize
	var users []*model.User
	var total int64
	var err error

	if searchQuery != "" {
		users, total, err = h.userService.Search(searchQuery, offset, pageSize)
	} else {
		users, total, err = h.userService.List(offset, pageSize)
	}
	if err != nil {
		response.InternalServerError(c, "failed to list users")
		return
	}

	var items []UserListItem
	for _, u := range users {
		items = append(items, UserListItem{
			ID:            u.ID,
			Email:         u.Email,
			Username:      u.Username,
			Name:          u.Name,
			EmailVerified: u.EmailVerified,
			Approved:      u.Approved,
			MFARequired:   u.MFARequired,
			IsAdmin:       u.IsAdmin,
			CreatedAt:     u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	response.SuccessPaginated(c, items, total, page, pageSize)
}

func (h *AdminHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	password := ""
	if req.Password != nil && *req.Password != "" {
		password = *req.Password
		result := zxcvbn.PasswordStrength(password, nil)
		if h.userService.GetConfig().Security.PasswordStrength > 0 && result.Score < h.userService.GetConfig().Security.PasswordStrength {
			response.BadRequest(c, "password is too weak")
			return
		}
		if h.userService.GetConfig().Security.PasswordMinLength > 0 && len(password) < h.userService.GetConfig().Security.PasswordMinLength {
			response.BadRequest(c, fmt.Sprintf("password must be at least %d characters", h.userService.GetConfig().Security.PasswordMinLength))
			return
		}
	}

	user, err := h.userService.Create(req.Email, req.Username, password)
	if err != nil {
		response.BadRequest(c, safeErrorMessage(err, "admin create user"))
		return
	}

	if req.Name != "" {
		user.Name = &req.Name
	}
	user.IsAdmin = req.IsAdmin
	user.Approved = req.Approved
	user.MFARequired = req.MFARequired

	if err := h.userService.Update(user); err != nil {
		response.InternalServerError(c, "failed to set user properties")
		return
	}

	response.Success(c, gin.H{
		"id":            user.ID,
		"email":         user.Email,
		"username":      user.Username,
		"name":          user.Name,
		"emailVerified": user.EmailVerified,
		"isAdmin":       user.IsAdmin,
		"approved":      user.Approved,
		"mfaRequired":   user.MFARequired,
		"createdAt":     user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *AdminHandler) GetUser(c *gin.Context) {
	id := c.Param("id")

	user, err := h.userService.GetByID(id)
	if err != nil || user == nil {
		response.NotFound(c, "user not found")
		return
	}

	response.Success(c, gin.H{
		"id":            user.ID,
		"email":         user.Email,
		"username":      user.Username,
		"name":          user.Name,
		"emailVerified": user.EmailVerified,
		"isAdmin":       user.IsAdmin,
		"approved":      user.Approved,
		"mfaRequired":   user.MFARequired,
		"createdAt":     user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *AdminHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	user, err := h.userService.GetByID(id)
	if err != nil || user == nil {
		response.NotFound(c, "user not found")
		return
	}

	if req.Email != nil {
		if *req.Email != "" {
			available, _ := h.userService.CheckEmailAvailable(*req.Email, user.ID)
			if !available {
				response.BadRequest(c, "email already in use by another user")
				return
			}
		}
		user.Email = req.Email
	}
	if req.Username != nil {
		newUsername := *req.Username
		if len(newUsername) < 3 || len(newUsername) > 32 {
			response.BadRequest(c, "username must be between 3 and 32 characters")
			return
		}
		if !adminUsernameRegex.MatchString(newUsername) {
			response.BadRequest(c, "username can only contain letters, numbers, and underscores")
			return
		}
		user.Username = newUsername
	}
	if req.Name != nil {
		user.Name = req.Name
	}
	if req.Approved != nil {
		user.Approved = *req.Approved
	}
	if req.IsAdmin != nil {
		user.IsAdmin = *req.IsAdmin
	}
	if req.MFARequired != nil {
		user.MFARequired = *req.MFARequired
	}

	err = h.userService.Update(user)
	if err != nil {
		response.InternalServerError(c, "failed to update user")
		return
	}

	response.SuccessWithMessage(c, "user updated successfully")
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	currentUser := middleware.GetCurrentUser(c)

	if currentUser.ID == id {
		response.BadRequest(c, "cannot delete yourself")
		return
	}

	existingUser, err := h.userService.GetByID(id)
	if err != nil || existingUser == nil {
		response.NotFound(c, "user not found")
		return
	}

	if existingUser.IsAdmin {
		adminCount, err := h.userService.CountAdmins()
		if err == nil && adminCount <= 1 {
			response.BadRequest(c, "cannot delete the last admin")
			return
		}
	}

	err = h.userService.Delete(id)
	if err != nil {
		response.InternalServerError(c, "failed to delete user")
		return
	}

	response.SuccessWithMessage(c, "user deleted successfully")
}

func (h *AdminHandler) ApproveUser(c *gin.Context) {
	id := c.Param("id")

	var req ApproveUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	user, err := h.userService.GetByID(id)
	if err != nil || user == nil {
		response.NotFound(c, "user not found")
		return
	}

	user.Approved = req.Approved
	err = h.userService.Update(user)
	if err != nil {
		response.InternalServerError(c, "failed to update user")
		return
	}

	response.SuccessWithMessage(c, "user approval status updated")
}

func (h *AdminHandler) ResetPassword(c *gin.Context) {
	id := c.Param("id")

	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	err := h.userService.ResetPassword(id, req.NewPassword)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "password is too weak") {
			response.BadRequest(c, "password strength is insufficient, please use a stronger password")
		} else if strings.Contains(errMsg, "password is too short") {
			response.BadRequest(c, "password is too short")
		} else {
			response.InternalServerError(c, "failed to reset password")
		}
		return
	}

	response.SuccessWithMessage(c, "password reset successfully")
}
