package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"quick-ticket/domain"
	"quick-ticket/interfaces/http/middleware"

	"github.com/google/uuid"
)

// ProfileHandler handles profile and member operations.
type ProfileHandler struct {
	// In-memory stores for testing purposes
	profiles map[string]*domain.UserProfile
	members  map[string][]*domain.Member
}

func NewProfileHandler() *ProfileHandler {
	return &ProfileHandler{
		profiles: make(map[string]*domain.UserProfile),
		members:  make(map[string][]*domain.Member),
	}
}

func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromContext(r.Context())
	if tenant == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	profile, exists := h.profiles[tenant.ID]
	if !exists {
		// Create a default profile if it doesn't exist
		profile = &domain.UserProfile{
			ID:        uuid.New().String(),
			TenantID:  tenant.ID,
			Email:     "admin@" + tenant.Name + ".com",
			FullName:  tenant.Name + " Admin",
			CreatedAt: time.Now().UTC(),
		}
		h.profiles[tenant.ID] = profile
	}

	writeJSON(w, http.StatusOK, profile)
}

func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromContext(r.Context())
	if tenant == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req domain.UserProfile
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	// Ensure ID and TenantID are preserved
	existing, exists := h.profiles[tenant.ID]
	if exists {
		req.ID = existing.ID
		req.TenantID = existing.TenantID
		req.CreatedAt = existing.CreatedAt
	} else {
		req.ID = uuid.New().String()
		req.TenantID = tenant.ID
		req.CreatedAt = time.Now().UTC()
	}

	h.profiles[tenant.ID] = &req
	writeJSON(w, http.StatusOK, req)
}

type CreateMemberRequest struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

func (h *ProfileHandler) CreateMember(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromContext(r.Context())
	if tenant == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	member := &domain.Member{
		ID:        uuid.New().String(),
		TenantID:  tenant.ID,
		Name:      req.Name,
		Role:      req.Role,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
	}

	h.members[tenant.ID] = append(h.members[tenant.ID], member)
	writeJSON(w, http.StatusCreated, member)
}

func (h *ProfileHandler) GetMembers(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromContext(r.Context())
	if tenant == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	members := h.members[tenant.ID]
	if members == nil {
		members = []*domain.Member{}
	}

	writeJSON(w, http.StatusOK, members)
}
