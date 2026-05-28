package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"quick-ticket/domain"
	"quick-ticket/interfaces/http/middleware"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProfileHandler handles profile and member operations.
type ProfileHandler struct {
	db *pgxpool.Pool
}

func NewProfileHandler(db *pgxpool.Pool) *ProfileHandler {
	return &ProfileHandler{
		db: db,
	}
}

func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromContext(r.Context())
	if tenant == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var profile domain.UserProfile
	err := h.db.QueryRow(r.Context(), `
		SELECT id, tenant_id, email, full_name, role, created_at
		FROM user_profiles
		WHERE tenant_id = $1 LIMIT 1
	`, tenant.ID).Scan(&profile.ID, &profile.TenantID, &profile.Email, &profile.FullName, &profile.Role, &profile.CreatedAt)

	if err != nil {
		// If no profile exists, create a default one
		profile = domain.UserProfile{
			ID:        uuid.New().String(),
			TenantID:  tenant.ID,
			Email:     "admin@" + tenant.Name + ".com",
			FullName:  tenant.Name + " Admin",
			Role:      "admin",
			CreatedAt: time.Now().UTC(),
		}
		_, insertErr := h.db.Exec(r.Context(), `
			INSERT INTO user_profiles (id, tenant_id, email, full_name, role, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT DO NOTHING
		`, profile.ID, profile.TenantID, profile.Email, profile.FullName, profile.Role, profile.CreatedAt)

		if insertErr != nil {
			writeError(w, http.StatusInternalServerError, "failed to create default profile: "+insertErr.Error())
			return
		}
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

	// Update existing profile (assuming user updates their own profile)
	_, err := h.db.Exec(r.Context(), `
		UPDATE user_profiles
		SET full_name = $1
		WHERE email = $2 AND tenant_id = $3
	`, req.FullName, req.Email,  tenant.ID)

	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update profile: "+err.Error())
		return
	}

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

	member := domain.Member{
		ID:        uuid.New().String(),
		TenantID:  tenant.ID,
		Name:      req.Name,
		Role:      req.Role,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
	}

	_, err := h.db.Exec(r.Context(), `
		INSERT INTO members (id, tenant_id, name, role, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, member.ID, member.TenantID, member.Name, member.Role, member.IsActive, member.CreatedAt)

	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create member: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, member)
}

func (h *ProfileHandler) GetMembers(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromContext(r.Context())
	if tenant == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	rows, err := h.db.Query(r.Context(), `
		SELECT id, tenant_id, name, role, is_active, created_at
		FROM members
		WHERE tenant_id = $1
	`, tenant.ID)

	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to query members: "+err.Error())
		return
	}
	defer rows.Close()

	members := []domain.Member{}
	for rows.Next() {
		var m domain.Member
		if err := rows.Scan(&m.ID, &m.TenantID, &m.Name, &m.Role, &m.IsActive, &m.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to scan member: "+err.Error())
			return
		}
		members = append(members, m)
	}

	writeJSON(w, http.StatusOK, members)
}
