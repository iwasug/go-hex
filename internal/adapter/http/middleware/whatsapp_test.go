package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"golang-chi-hexagonal-base/internal/domain/entity"
	"golang-chi-hexagonal-base/internal/domain/repository"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string, deletedBy string) error {
	args := m.Called(ctx, id, deletedBy)
	return args.Error(0)
}

func (m *MockUserRepository) Restore(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, params repository.ListUsersParams) ([]*entity.User, int, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*entity.User), args.Int(1), args.Error(2)
}

func (m *MockUserRepository) GetUserRoles(ctx context.Context, userID string) ([]*entity.Role, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Role), args.Error(1)
}

func (m *MockUserRepository) AssignRole(ctx context.Context, userID, roleID string) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockUserRepository) RemoveRole(ctx context.Context, userID, roleID string) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserPermissions(ctx context.Context, userID string) ([]*entity.Permission, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Permission), args.Error(1)
}

// MockWhatsAppSessionRepository is a mock implementation of WhatsAppSessionRepository
type MockWhatsAppSessionRepository struct {
	mock.Mock
}

func (m *MockWhatsAppSessionRepository) Create(ctx context.Context, session *entity.WhatsAppSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockWhatsAppSessionRepository) FindByID(ctx context.Context, id string) (*entity.WhatsAppSession, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WhatsAppSession), args.Error(1)
}

func (m *MockWhatsAppSessionRepository) FindByUserID(ctx context.Context, userID string) ([]*entity.WhatsAppSession, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.WhatsAppSession), args.Error(1)
}

func (m *MockWhatsAppSessionRepository) Update(ctx context.Context, session *entity.WhatsAppSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockWhatsAppSessionRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWhatsAppSessionRepository) List(ctx context.Context, params repository.ListSessionsParams) ([]*entity.WhatsAppSession, int, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*entity.WhatsAppSession), args.Int(1), args.Error(2)
}

func (m *MockWhatsAppSessionRepository) UpdateStatus(ctx context.Context, id string, status entity.SessionStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockWhatsAppSessionRepository) UpdateQRCode(ctx context.Context, id string, qrCode string, expiry time.Time) error {
	args := m.Called(ctx, id, qrCode, expiry)
	return args.Error(0)
}

func (m *MockWhatsAppSessionRepository) ClearQRCode(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWhatsAppSessionRepository) UpdatePhoneNumber(ctx context.Context, id string, phoneNumber string) error {
	args := m.Called(ctx, id, phoneNumber)
	return args.Error(0)
}

func (m *MockWhatsAppSessionRepository) UpdateLastActivity(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWhatsAppSessionRepository) UpdateDeviceInfo(ctx context.Context, id string, deviceInfo string) error {
	args := m.Called(ctx, id, deviceInfo)
	return args.Error(0)
}

func (m *MockWhatsAppSessionRepository) FindActiveByUserID(ctx context.Context, userID string) ([]*entity.WhatsAppSession, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.WhatsAppSession), args.Error(1)
}

func (m *MockWhatsAppSessionRepository) FindConnectedSessions(ctx context.Context) ([]*entity.WhatsAppSession, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.WhatsAppSession), args.Error(1)
}

func (m *MockWhatsAppSessionRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockWhatsAppSessionRepository) FindByUserIDAndName(ctx context.Context, userID, sessionName string) (*entity.WhatsAppSession, error) {
	args := m.Called(ctx, userID, sessionName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WhatsAppSession), args.Error(1)
}

func TestInjectAdminStatus(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		setupMock      func(*MockUserRepository)
		expectedAdmin  bool
		shouldHaveFlag bool
	}{
		{
			name:   "User with admin role",
			userID: "user-123",
			setupMock: func(m *MockUserRepository) {
				m.On("GetUserRoles", mock.Anything, "user-123").Return([]*entity.Role{
					{Name: "admin"},
				}, nil)
			},
			expectedAdmin:  true,
			shouldHaveFlag: true,
		},
		{
			name:   "User without admin role",
			userID: "user-456",
			setupMock: func(m *MockUserRepository) {
				m.On("GetUserRoles", mock.Anything, "user-456").Return([]*entity.Role{
					{Name: "user"},
				}, nil)
			},
			expectedAdmin:  false,
			shouldHaveFlag: true,
		},
		{
			name:   "User with multiple roles including admin",
			userID: "user-789",
			setupMock: func(m *MockUserRepository) {
				m.On("GetUserRoles", mock.Anything, "user-789").Return([]*entity.Role{
					{Name: "user"},
					{Name: "admin"},
					{Name: "moderator"},
				}, nil)
			},
			expectedAdmin:  true,
			shouldHaveFlag: true,
		},
		{
			name:   "Error getting user roles",
			userID: "user-error",
			setupMock: func(m *MockUserRepository) {
				m.On("GetUserRoles", mock.Anything, "user-error").Return(nil, errors.New("database error"))
			},
			expectedAdmin:  false,
			shouldHaveFlag: false,
		},
		{
			name:           "No user ID in context",
			userID:         "",
			setupMock:      func(m *MockUserRepository) {},
			expectedAdmin:  false,
			shouldHaveFlag: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockUserRepo := new(MockUserRepository)
			tt.setupMock(mockUserRepo)

			// Create test handler
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.shouldHaveFlag {
					isAdmin := r.Context().Value("is_admin")
					assert.NotNil(t, isAdmin)
					assert.Equal(t, tt.expectedAdmin, isAdmin.(bool))
				} else {
					isAdmin := r.Context().Value("is_admin")
					if isAdmin != nil {
						assert.False(t, isAdmin.(bool))
					}
				}
				w.WriteHeader(http.StatusOK)
			})

			// Create middleware
			middleware := InjectAdminStatus(mockUserRepo)
			handler := middleware(testHandler)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.userID != "" {
				ctx := context.WithValue(req.Context(), "user_id", tt.userID)
				req = req.WithContext(ctx)
			}

			// Execute
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			// Assert
			assert.Equal(t, http.StatusOK, rr.Code)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestValidateSessionOwnership(t *testing.T) {
	tests := []struct {
		name           string
		sessionID      string
		userID         string
		isAdmin        bool
		setupMock      func(*MockWhatsAppSessionRepository)
		expectedStatus int
	}{
		{
			name:      "User owns session",
			sessionID: "session-123",
			userID:    "user-123",
			isAdmin:   false,
			setupMock: func(m *MockWhatsAppSessionRepository) {
				m.On("FindByID", mock.Anything, "session-123").Return(&entity.WhatsAppSession{
					ID:     "session-123",
					UserID: "user-123",
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "User does not own session",
			sessionID: "session-123",
			userID:    "user-456",
			isAdmin:   false,
			setupMock: func(m *MockWhatsAppSessionRepository) {
				m.On("FindByID", mock.Anything, "session-123").Return(&entity.WhatsAppSession{
					ID:     "session-123",
					UserID: "user-123",
				}, nil)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:      "Admin can access any session",
			sessionID: "session-123",
			userID:    "admin-user",
			isAdmin:   true,
			setupMock: func(m *MockWhatsAppSessionRepository) {
				// Admin bypasses ownership check, so no mock call expected
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "Session not found",
			sessionID: "session-999",
			userID:    "user-123",
			isAdmin:   false,
			setupMock: func(m *MockWhatsAppSessionRepository) {
				m.On("FindByID", mock.Anything, "session-999").Return(nil, errors.New("not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "No user ID in context",
			sessionID:      "session-123",
			userID:         "",
			isAdmin:        false,
			setupMock:      func(m *MockWhatsAppSessionRepository) {},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockSessionRepo := new(MockWhatsAppSessionRepository)
			tt.setupMock(mockSessionRepo)

			// Create test handler
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Create middleware
			middleware := ValidateSessionOwnership(mockSessionRepo)

			// Create router with chi to support URL params
			r := chi.NewRouter()
			r.Route("/sessions/{id}", func(r chi.Router) {
				r.Use(middleware)
				r.Get("/", testHandler)
			})

			// Create request
			url := "/sessions/" + tt.sessionID
			req := httptest.NewRequest(http.MethodGet, url, nil)

			// Build context
			ctx := req.Context()
			if tt.userID != "" {
				ctx = context.WithValue(ctx, "user_id", tt.userID)
			}
			if tt.isAdmin {
				ctx = context.WithValue(ctx, "is_admin", true)
			}
			req = req.WithContext(ctx)

			// Execute
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockSessionRepo.AssertExpectations(t)
		})
	}
}
