package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mafzaidi/stackforge/internal/delivery/http/middleware"
	"github.com/mafzaidi/stackforge/internal/delivery/http/serializer"
	"github.com/mafzaidi/stackforge/internal/pkg/response"
	"github.com/mafzaidi/stackforge/internal/usecase/userprofiles"
)

// CreateUserProfileRequest is intentionally empty — profile is created automatically on first login.
type CreateUserProfileRequest struct{}

// UserProfilesHandler handles user profiles HTTP requests.
type UserProfilesHandler struct {
	createUC              userprofiles.CreateUseCase
	getByUserIDUC         userprofiles.GetByUserIDUseCase
	setupMasterPasswordUC userprofiles.SetupMasterPasswordUseCase
}

// NewUserProfilesHandler creates a new UserProfilesHandler.
func NewUserProfilesHandler(
	createUC userprofiles.CreateUseCase,
	getByUserIDUC userprofiles.GetByUserIDUseCase,
	setupMasterPasswordUC userprofiles.SetupMasterPasswordUseCase,
) *UserProfilesHandler {
	return &UserProfilesHandler{
		createUC:              createUC,
		getByUserIDUC:         getByUserIDUC,
		setupMasterPasswordUC: setupMasterPasswordUC,
	}
}

// Create handles POST /api/user-profiles
// Called automatically on first login — no request body needed.
func (h *UserProfilesHandler) Create(c *gin.Context) {
	claims, err := middleware.GetClaims(c)
	if err != nil {
		response.Unauthorized(c, "Missing or invalid authentication token")
		return
	}

	profile, err := h.createUC.Execute(c.Request.Context(), claims.Subject)
	if err != nil {
		response.InternalServerError(c, "Failed to create user profile")
		return
	}

	response.Created(c, "User profile created successfully", serializer.FromUserProfileCreated(profile))
}

func (h *UserProfilesHandler) GetByUserID(c *gin.Context) {
	claims, err := middleware.GetClaims(c)
	if err != nil {
		response.Unauthorized(c, "Missing or invalid authentication token")
		return
	}

	profile, err := h.getByUserIDUC.Execute(c.Request.Context(), claims.Subject)
	if err != nil {
		response.InternalServerError(c, "Failed to fetch user profile")
		return
	}

	if profile == nil {
		response.NotFound(c, "User profile not found")
		return
	}

	response.Success(c, "User profile retrieved successfully", serializer.FromUserProfile(profile))
}

// SetupMasterPasswordRequest represents the request body for setting up master password.
type SetupMasterPasswordRequest struct {
	MasterPassword string `json:"master_password" binding:"required"`
}

// SetupMasterPassword handles POST /api/user-profiles/master-password
// Called when user creates their first credential.
func (h *UserProfilesHandler) SetupMasterPassword(c *gin.Context) {
	claims, err := middleware.GetClaims(c)
	if err != nil {
		response.Unauthorized(c, "Missing or invalid authentication token")
		return
	}

	var req SetupMasterPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "master_password is required")
		return
	}

	if err := h.setupMasterPasswordUC.Execute(c.Request.Context(), claims.Subject, req.MasterPassword); err != nil {
		if err.Error() == "master password already set" {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalServerError(c, "Failed to setup master password")
		return
	}

	response.Success(c, "Master password configured successfully", nil)
}
