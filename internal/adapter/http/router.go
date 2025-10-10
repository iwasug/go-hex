// Package http provides HTTP routing and middleware setup for the application.
// It wires together all dependencies (repositories, services, handlers) and
// configures the router with appropriate middleware and route protection.
package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"

	"golang-chi-hexagonal-base/internal/adapter/crypto"
	"golang-chi-hexagonal-base/internal/adapter/http/handler"
	mediaadapter "golang-chi-hexagonal-base/internal/adapter/media"
	custommiddleware "golang-chi-hexagonal-base/internal/adapter/http/middleware"
	"golang-chi-hexagonal-base/internal/adapter/postgres"
	"golang-chi-hexagonal-base/internal/adapter/webhook"
	"golang-chi-hexagonal-base/internal/adapter/whatsapp"
	"golang-chi-hexagonal-base/internal/config"
	"golang-chi-hexagonal-base/internal/domain/service"
	"golang-chi-hexagonal-base/internal/pkg/logger"
	"golang-chi-hexagonal-base/internal/pkg/media"
	webhookpkg "golang-chi-hexagonal-base/internal/pkg/webhook"
)

// Router holds all dependencies for HTTP routing
type Router struct {
	config        *config.Config
	db            *postgres.DB
	clientManager *whatsapp.ClientManager
}

// NewRouter creates a new Router instance
func NewRouter(config *config.Config, db *postgres.DB) *Router {
	return &Router{
		config: config,
		db:     db,
	}
}

// Setup sets up all routes and middleware
func (rt *Router) Setup() http.Handler {
	r := chi.NewRouter()

	rateLimiter := custommiddleware.NewRateLimiter(100, 200)

	// Setup middleware chain (order matters!)
	r.Use(custommiddleware.Recovery)
	r.Use(custommiddleware.RequestID)
	r.Use(custommiddleware.Logger)
	r.Use(custommiddleware.SecurityHeaders)
	r.Use(custommiddleware.Metrics)
	r.Use(custommiddleware.Tracing("go-hex-api"))
	r.Use(rateLimiter.Limit)
	r.Use(custommiddleware.Compress)
	r.Use(rt.corsMiddleware())
	r.Use(middleware.Timeout(60 * time.Second))

	// Wire dependencies
	// Repositories
	userRepo := postgres.NewUserRepository(rt.db)
	roleRepo := postgres.NewRoleRepository(rt.db)
	permissionRepo := postgres.NewPermissionRepository(rt.db)
	tokenRepo := postgres.NewTokenRepository(rt.db)
	whatsappSessionRepo := postgres.NewWhatsAppSessionRepository(rt.db)
	whatsappMessageRepo := postgres.NewWhatsAppMessageRepository(rt.db)

	// Crypto adapters
	bcryptHasher := crypto.NewBcryptHasher(rt.config.Bcrypt.Cost)
	jwtManager := crypto.NewJWTManager(
		rt.config.JWT.Secret,
		rt.config.JWT.AccessTokenExpiry,
		rt.config.JWT.RefreshTokenExpiry,
	)

	// Services
	authService := service.NewAuthService(userRepo, roleRepo, tokenRepo, bcryptHasher, jwtManager)
	userService := service.NewUserService(userRepo, roleRepo)
	roleService := service.NewRoleService(roleRepo, permissionRepo)

	// WhatsApp adapters and services
	log := logger.NewFromConfig(rt.config.Logging.Level, rt.config.Logging.Format)
	
	// Initialize WhatsApp client manager
	rt.clientManager = whatsapp.NewClientManager(
		rt.db.DB, // Get underlying *sql.DB
		whatsappSessionRepo,
		whatsappMessageRepo,
		log,
	)
	
	// Wrap client manager with adapter to satisfy service interface
	clientManager := whatsapp.NewClientManagerAdapter(rt.clientManager)
	
	// Initialize webhook client
	webhookConfig := webhookpkg.DefaultConfig()
	webhookClient := webhook.NewClientAdapter(webhookConfig, log)
	
	// Initialize media manager
	mediaManager, err := media.NewManager(
		rt.config.WhatsApp.MediaStoragePath,
		rt.config.WhatsApp.MediaMaxSizeMB,
	)
	if err != nil {
		log.Error("Failed to initialize media manager", logger.Fields{
			"error": err.Error(),
		})
		panic(err)
	}
	mediaAdapter := mediaadapter.NewManagerAdapter(mediaManager)
	
	// Initialize WhatsApp service
	whatsappService := service.NewWhatsAppService(
		whatsappSessionRepo,
		whatsappMessageRepo,
		userRepo,
		clientManager,
		webhookClient,
		mediaAdapter,
		log,
	)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	roleHandler := handler.NewRoleHandler(roleService)
	healthHandler := handler.NewHealthHandler(rt.db)
	healthHandler.SetClientManager(rt.clientManager) // Set WhatsApp client manager for health checks
	whatsappHandler := handler.NewWhatsAppHandler(whatsappService)

	// Health check and observability endpoints
	r.Get("/health", healthHandler.Health)
	r.Get("/health/ready", healthHandler.Readiness)
	r.Get("/health/live", healthHandler.Liveness)
	r.Handle("/metrics", promhttp.Handler())

	// Swagger documentation
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes (auth endpoints)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.RefreshToken)
			r.Post("/forgot-password", authHandler.ForgotPassword)
			r.Post("/reset-password", authHandler.ResetPassword)
			r.Post("/logout", authHandler.Logout)
		})

		// Protected routes (require authentication)
		r.Group(func(r chi.Router) {
			// Apply auth middleware
			r.Use(custommiddleware.Auth(jwtManager))
			// Inject admin status for all protected routes
			r.Use(custommiddleware.InjectAdminStatus(userRepo))

			// User routes
			r.Route("/users", func(r chi.Router) {
				// List users - requires users:read permission
				r.With(custommiddleware.RequirePermission(userRepo, "users:read")).Get("/", userHandler.List)

				// Get user by ID - requires users:read permission
				r.With(custommiddleware.RequirePermission(userRepo, "users:read")).Get("/{id}", userHandler.GetByID)

				// Update user - requires users:write permission
				r.With(custommiddleware.RequirePermission(userRepo, "users:write")).Put("/{id}", userHandler.Update)

				// Delete user - requires users:delete permission
				r.With(custommiddleware.RequirePermission(userRepo, "users:delete")).Delete("/{id}", userHandler.Delete)

				// Assign role to user - requires users:write permission
				r.With(custommiddleware.RequirePermission(userRepo, "users:write")).Post("/{id}/roles", userHandler.AssignRole)

				// Remove role from user - requires users:write permission
				r.With(custommiddleware.RequirePermission(userRepo, "users:write")).Delete("/{id}/roles/{role_id}", userHandler.RemoveRole)
			})

			// Role routes (admin only)
			r.Route("/roles", func(r chi.Router) {
				// List roles - requires roles:read permission
				r.With(custommiddleware.RequirePermission(userRepo, "roles:read")).Get("/", roleHandler.List)

				// Create role - requires roles:write permission
				r.With(custommiddleware.RequirePermission(userRepo, "roles:write")).Post("/", roleHandler.Create)

				// Get role by ID - requires roles:read permission
				r.With(custommiddleware.RequirePermission(userRepo, "roles:read")).Get("/{id}", roleHandler.GetByID)

				// Update role - requires roles:write permission
				r.With(custommiddleware.RequirePermission(userRepo, "roles:write")).Put("/{id}", roleHandler.Update)

				// Delete role - requires roles:delete permission
				r.With(custommiddleware.RequirePermission(userRepo, "roles:delete")).Delete("/{id}", roleHandler.Delete)
			})

			// WhatsApp routes
			r.Route("/whatsapp", func(r chi.Router) {
				r.Route("/sessions", func(r chi.Router) {
					// Create session - requires whatsapp:write permission
					r.With(custommiddleware.RequirePermission(userRepo, "whatsapp:write")).Post("/", whatsappHandler.CreateSession)

					// List sessions - requires whatsapp:read permission
					r.With(custommiddleware.RequirePermission(userRepo, "whatsapp:read")).Get("/", whatsappHandler.ListSessions)

					// Session-specific routes with ownership validation
					r.Group(func(r chi.Router) {
						// Apply session ownership validation middleware
						r.Use(custommiddleware.ValidateSessionOwnership(whatsappSessionRepo))

						// Get session by ID - requires whatsapp:read permission
						r.With(custommiddleware.RequirePermission(userRepo, "whatsapp:read")).Get("/{id}", whatsappHandler.GetSession)

						// Delete session - requires whatsapp:delete permission
						r.With(custommiddleware.RequirePermission(userRepo, "whatsapp:delete")).Delete("/{id}", whatsappHandler.DeleteSession)

						// Get QR code - requires whatsapp:read permission
						r.With(custommiddleware.RequirePermission(userRepo, "whatsapp:read")).Get("/{id}/qr", whatsappHandler.GetQRCode)
						r.With(custommiddleware.RequirePermission(userRepo, "whatsapp:read")).Get("/{id}/pairing-code", whatsappHandler.GetPairingCode)

						// Send text message - requires whatsapp:write permission
						r.With(custommiddleware.RequirePermission(userRepo, "whatsapp:write")).Post("/{id}/messages", whatsappHandler.SendTextMessage)

						// Send media message - requires whatsapp:write permission
						r.With(custommiddleware.RequirePermission(userRepo, "whatsapp:write")).Post("/{id}/media", whatsappHandler.SendMediaMessage)

						// Get message history - requires whatsapp:read permission
						r.With(custommiddleware.RequirePermission(userRepo, "whatsapp:read")).Get("/{id}/messages", whatsappHandler.GetMessages)

						// Reconnect session - requires whatsapp:write permission
						r.With(custommiddleware.RequirePermission(userRepo, "whatsapp:write")).Post("/{id}/reconnect", whatsappHandler.ReconnectSession)

						// Get session status - requires whatsapp:read permission
						r.With(custommiddleware.RequirePermission(userRepo, "whatsapp:read")).Get("/{id}/status", whatsappHandler.GetSessionStatus)
					})
				})
			})
		})
	})

	return r
}

// corsMiddleware configures CORS settings
func (rt *Router) corsMiddleware() func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		// AllowedOrigins: []string{"https://example.com"}, // Use specific origins in production
		AllowedOrigins:   []string{"*"}, // Allow all origins in development
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
}

// InitializeWhatsApp initializes the WhatsApp service and restores connected clients
func (rt *Router) InitializeWhatsApp(ctx context.Context) error {
	if rt.clientManager == nil {
		return fmt.Errorf("WhatsApp client manager not initialized")
	}

	// Restore all connected sessions from database
	if err := rt.clientManager.RestoreClients(ctx); err != nil {
		return fmt.Errorf("failed to restore WhatsApp clients: %w", err)
	}

	return nil
}

// ShutdownWhatsApp gracefully shuts down all WhatsApp clients
func (rt *Router) ShutdownWhatsApp(ctx context.Context) error {
	if rt.clientManager == nil {
		// WhatsApp not initialized, nothing to shutdown
		return nil
	}

	// Shutdown all active clients
	if err := rt.clientManager.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown WhatsApp clients: %w", err)
	}

	return nil
}

// GetClientManager returns the WhatsApp client manager for health checks
func (rt *Router) GetClientManager() *whatsapp.ClientManager {
	return rt.clientManager
}
