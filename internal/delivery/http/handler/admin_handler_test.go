package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prassaaa/itemly-backend/internal/model"
	"github.com/prassaaa/itemly-backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- mock ---

type mockAdminUsecase struct {
	assignRoleFn func(targetUserID uuid.UUID, role string) (*model.User, error)
}

func (m *mockAdminUsecase) AssignRole(targetUserID uuid.UUID, role string) (*model.User, error) {
	return m.assignRoleFn(targetUserID, role)
}

// --- Dashboard ---

func TestAdminDashboard_200(t *testing.T) {
	uc := &mockAdminUsecase{}
	h := NewAdminHandler(uc)
	r := gin.New()
	r.GET("/dashboard", func(c *gin.Context) {
		c.Set("username", "adminuser")
		c.Next()
	}, h.AdminDashboard)

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "welcome to admin dashboard", resp["message"])
	assert.Equal(t, "adminuser", resp["user"])
}

// --- AssignRole ---

func TestAssignRole_200(t *testing.T) {
	userID := uuid.New()
	uc := &mockAdminUsecase{
		assignRoleFn: func(targetUserID uuid.UUID, role string) (*model.User, error) {
			return &model.User{
				ID:       targetUserID,
				Username: "target",
				Email:    "target@example.com",
				Role:     model.Role(role),
			}, nil
		},
	}
	h := NewAdminHandler(uc)
	r := gin.New()
	r.PUT("/users/:id/role", h.AssignRole)

	body, _ := json.Marshal(map[string]string{"role": "manager"})
	req := httptest.NewRequest(http.MethodPut, "/users/"+userID.String()+"/role", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "manager", resp["role"])
}

func TestAssignRole_400_InvalidUUID(t *testing.T) {
	uc := &mockAdminUsecase{}
	h := NewAdminHandler(uc)
	r := gin.New()
	r.PUT("/users/:id/role", h.AssignRole)

	body, _ := json.Marshal(map[string]string{"role": "admin"})
	req := httptest.NewRequest(http.MethodPut, "/users/not-a-uuid/role", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAssignRole_400_InvalidRole(t *testing.T) {
	userID := uuid.New()
	uc := &mockAdminUsecase{
		assignRoleFn: func(targetUserID uuid.UUID, role string) (*model.User, error) {
			return nil, usecase.ErrInvalidRole
		},
	}
	h := NewAdminHandler(uc)
	r := gin.New()
	r.PUT("/users/:id/role", h.AssignRole)

	// Note: "admin" passes DTO validation (oneof=admin manager staff)
	// but we test the usecase returning ErrInvalidRole for a different scenario
	body, _ := json.Marshal(map[string]string{"role": "admin"})
	req := httptest.NewRequest(http.MethodPut, "/users/"+userID.String()+"/role", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAssignRole_404(t *testing.T) {
	userID := uuid.New()
	uc := &mockAdminUsecase{
		assignRoleFn: func(targetUserID uuid.UUID, role string) (*model.User, error) {
			return nil, usecase.ErrUserNotFound
		},
	}
	h := NewAdminHandler(uc)
	r := gin.New()
	r.PUT("/users/:id/role", h.AssignRole)

	body, _ := json.Marshal(map[string]string{"role": "admin"})
	req := httptest.NewRequest(http.MethodPut, "/users/"+userID.String()+"/role", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAssignRole_400_MissingBody(t *testing.T) {
	userID := uuid.New()
	uc := &mockAdminUsecase{}
	h := NewAdminHandler(uc)
	r := gin.New()
	r.PUT("/users/:id/role", h.AssignRole)

	req := httptest.NewRequest(http.MethodPut, "/users/"+userID.String()+"/role", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAssignRole_500(t *testing.T) {
	userID := uuid.New()
	uc := &mockAdminUsecase{
		assignRoleFn: func(targetUserID uuid.UUID, role string) (*model.User, error) {
			return nil, errors.New("db error")
		},
	}
	h := NewAdminHandler(uc)
	r := gin.New()
	r.PUT("/users/:id/role", h.AssignRole)

	body, _ := json.Marshal(map[string]string{"role": "admin"})
	req := httptest.NewRequest(http.MethodPut, "/users/"+userID.String()+"/role", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
