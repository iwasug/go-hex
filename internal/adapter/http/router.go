// Package http provides HTTP routing and middleware setup for the application.
// It wires together all dependencies (repositories, services, handlers) and
// configures the router with appropriate middleware and route protection.
package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"

	"golang-chi-hexagonal-base/internal/adapter/crypto"
	"golang-chi-hexagonal-base/internal/adapter/http/handler"
	custommiddleware "golang-chi-hexagonal-base/internal/adapter/http/middleware"
	"golang-chi-hexagonal-base/internal/adapter/postgres"
	"golang-chi-hexagonal-base/internal/config"
	"golang-chi-hexagonal-base/internal/domain/service"
)

// Router holds all dependencies for HTTP routing
type Router struct {
	config *config.Config
	db     *postgres.DB
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

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	roleHandler := handler.NewRoleHandler(roleService)
	healthHandler := handler.NewHealthHandler(rt.db)

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
