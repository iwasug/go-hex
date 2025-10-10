package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/skip2/go-qrcode"

	"golang-chi-hexagonal-base/internal/adapter/http/request"
	"golang-chi-hexagonal-base/internal/adapter/http/response"
	"golang-chi-hexagonal-base/internal/domain/entity"
	"golang-chi-hexagonal-base/internal/domain/service"
	apperrors "golang-chi-hexagonal-base/internal/pkg/errors"
	"golang-chi-hexagonal-base/internal/pkg/validator"
)

// WhatsAppHandler handles WhatsApp-related HTTP requests
type WhatsAppHandler struct {
	whatsappService *service.WhatsAppService
}

// NewWhatsAppHandler creates a new WhatsAppHandler instance
func NewWhatsAppHandler(whatsappService *service.WhatsAppService) *WhatsAppHandler {
	return &WhatsAppHandler{
		whatsappService: whatsappService,
	}
}

// safeString safely dereferences a string pointer, returning empty string if nil
func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// handleWhatsAppError handles WhatsAppError responses
func handleWhatsAppError(w http.ResponseWriter, err error) {
	// Check if it's a WhatsAppError
	if waErr := apperrors.GetWhatsAppError(err); waErr != nil {
		respondError(w, waErr.HTTPStatus, waErr.Code, waErr.Message, waErr.Details)
		return
	}
	
	// Check for service errors
	if errors.Is(err, service.ErrUserNotFound) {
		respondError(w, http.StatusNotFound, response.ErrCodeNotFound, "User not found", nil)
		return
	}
	
	// Default to internal server error
	respondError(w, http.StatusInternalServerError, response.ErrCodeInternal, "Internal server error", nil)
}

// CreateSession handles creating a new WhatsApp session
// @Summary Create WhatsApp session
// @Description Create a new WhatsApp session for the authenticated user
// @Tags WhatsApp
// @Accept json
// @Produce json
// @Param request body request.CreateSessionRequest true "Session creation data"
// @Success 201 {object} response.WhatsAppSessionResponse "Session created successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid input"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - requires whatsapp:write permission"
// @Failure 409 {object} response.ErrorResponse "Session with this name already exists"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /whatsapp/sessions [post]
func (h *WhatsAppHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID := r.Context().Value("user_id")
	if userID == nil {
		respondError(w, http.StatusUnauthorized, response.ErrCodeUnauthorized, "User ID not found in context", nil)
		return
	}

	// Parse request body
	var req request.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, response.ErrCodeValidation, "Invalid request body", nil)
		return
	}

	// Validate input
	v := validator.New()
	v.ValidateRequired("session_name", req.SessionName)
	v.ValidateMaxLength("session_name", req.SessionName, 255)
	
	// Validate webhook URL if provided
	if req.WebhookURL != "" {
		v.ValidateURL("webhook_url", req.WebhookURL)
		v.ValidateMaxLength("webhook_url", req.WebhookURL, 2048)
	}
	
	// Validate webhook secret if provided
	if req.WebhookSecret != "" {
		v.ValidateMaxLength("webhook_secret", req.WebhookSecret, 255)
	}

	if v.HasErrors() {
		respondValidationError(w, v.Errors())
		return
	}

	// Call service to create session
	session, err := h.whatsappService.CreateSession(r.Context(), userID.(string), service.CreateSessionInput{
		SessionName:   req.SessionName,
		WebhookURL:    req.WebhookURL,
		WebhookSecret: req.WebhookSecret,
	})

	if err != nil {
		handleWhatsAppError(w, err)
		return
	}

	// Return response
	respondSuccess(w, http.StatusCreated, "Session created successfully", toWhatsAppSessionResponse(session))
}

// ListSessions handles listing WhatsApp sessions
// @Summary List WhatsApp sessions
// @Description Get a paginated list of WhatsApp sessions with optional filtering
// @Tags WhatsApp
// @Accept json
// @Produce json
// @Param status query string false "Filter by status (connected, disconnected, connecting, authenticating, failed)"
// @Param is_active query boolean false "Filter by active status"
// @Param search query string false "Search by session name or phone number"
// @Param order_by query string false "Sort field (session_name, status, created_at, last_activity)"
// @Param order query string false "Sort direction (asc, desc)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.PaginatedResponse "List of sessions"
// @Failure 400 {object} response.ErrorResponse "Invalid parameters"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - requires whatsapp:read permission"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /whatsapp/sessions [get]
func (h *WhatsAppHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID := r.Context().Value("user_id")
	if userID == nil {
		respondError(w, http.StatusUnauthorized, response.ErrCodeUnauthorized, "User ID not found in context", nil)
		return
	}

	// Check if user is admin (from context, set by RBAC middleware)
	isAdmin := false
	if adminVal := r.Context().Value("is_admin"); adminVal != nil {
		isAdmin = adminVal.(bool)
	}

	// Parse query parameters
	statusStr := r.URL.Query().Get("status")
	isActiveStr := r.URL.Query().Get("is_active")
	search := r.URL.Query().Get("search")
	orderBy := r.URL.Query().Get("order_by")
	order := r.URL.Query().Get("order")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	// Parse page and limit
	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Parse is_active
	var isActive *bool
	if isActiveStr != "" {
		if isActiveStr == "true" {
			val := true
			isActive = &val
		} else if isActiveStr == "false" {
			val := false
			isActive = &val
		}
	}

	// Parse status
	var status entity.SessionStatus
	if statusStr != "" {
		status = entity.SessionStatus(statusStr)
	}

	// Validate order_by field
	v := validator.New()
	allowedOrderByFields := []string{"session_name", "status", "created_at", "last_activity"}
	if orderBy != "" {
		v.ValidateOrderBy("order_by", orderBy, allowedOrderByFields)
	}
	if order != "" {
		v.ValidateOrderDirection("order", order)
	}

	if v.HasErrors() {
		respondValidationError(w, v.Errors())
		return
	}

	// Call service
	result, err := h.whatsappService.ListSessions(r.Context(), userID.(string), isAdmin, service.ListSessionsInput{
		Status:   status,
		IsActive: isActive,
		Search:   search,
		OrderBy:  orderBy,
		Order:    order,
		Page:     page,
		Limit:    limit,
	})

	if err != nil {
		respondError(w, http.StatusInternalServerError, response.ErrCodeInternal, "Failed to list sessions", nil)
		return
	}

	// Convert sessions to response format
	sessionResponses := make([]response.WhatsAppSessionResponse, len(result.Sessions))
	for i, session := range result.Sessions {
		sessionResponses[i] = toWhatsAppSessionResponse(session)
	}

	// Return paginated response
	respondJSON(w, http.StatusOK, response.NewSuccessPaginatedResponse(
		"Sessions retrieved successfully",
		sessionResponses,
		response.Pagination{
			Total:      result.Total,
			Page:       result.Page,
			Limit:      result.Limit,
			TotalPages: result.TotalPages,
		},
	))
}

// GetSession handles retrieving a WhatsApp session by ID
// @Summary Get WhatsApp session by ID
// @Description Retrieve detailed information about a specific WhatsApp session
// @Tags WhatsApp
// @Accept json
// @Produce json
// @Param id path string true "Session ID"
// @Success 200 {object} response.WhatsAppSessionResponse "Session details"
// @Failure 400 {object} response.ErrorResponse "Invalid session ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - requires whatsapp:read permission or session ownership"
// @Failure 404 {object} response.ErrorResponse "Session not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /whatsapp/sessions/{id} [get]
func (h *WhatsAppHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	// Get session ID from URL
	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		respondError(w, http.StatusBadRequest, response.ErrCodeValidation, "Session ID is required", nil)
		return
	}

	// Get user ID from context
	userID := r.Context().Value("user_id")
	if userID == nil {
		respondError(w, http.StatusUnauthorized, response.ErrCodeUnauthorized, "User ID not found in context", nil)
		return
	}

	// Check if user is admin
	isAdmin := false
	if adminVal := r.Context().Value("is_admin"); adminVal != nil {
		isAdmin = adminVal.(bool)
	}

	// Call service
	session, err := h.whatsappService.GetSession(r.Context(), sessionID, userID.(string), isAdmin)
	if err != nil {
		handleWhatsAppError(w, err)
		return
	}

	// Return response
	respondSuccess(w, http.StatusOK, "Session retrieved successfully", toWhatsAppSessionResponse(session))
}

// DeleteSession handles deleting a WhatsApp session
// @Summary Delete WhatsApp session
// @Description Delete a WhatsApp session and disconnect the client
// @Tags WhatsApp
// @Accept json
// @Produce json
// @Param id path string true "Session ID"
// @Success 200 {object} response.BaseResponse "Session deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid session ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - requires whatsapp:delete permission or session ownership"
// @Failure 404 {object} response.ErrorResponse "Session not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /whatsapp/sessions/{id} [delete]
func (h *WhatsAppHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	// Get session ID from URL
	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		respondError(w, http.StatusBadRequest, response.ErrCodeValidation, "Session ID is required", nil)
		return
	}

	// Get user ID from context
	userID := r.Context().Value("user_id")
	if userID == nil {
		respondError(w, http.StatusUnauthorized, response.ErrCodeUnauthorized, "User ID not found in context", nil)
		return
	}

	// Check if user is admin
	isAdmin := false
	if adminVal := r.Context().Value("is_admin"); adminVal != nil {
		isAdmin = adminVal.(bool)
	}

	// Call service
	err := h.whatsappService.DeleteSession(r.Context(), sessionID, userID.(string), isAdmin)
	if err != nil {
		handleWhatsAppError(w, err)
		return
	}

	// Return response
	respondSuccess(w, http.StatusOK, "Session deleted successfully", nil)
}

// GetQRCode handles generating and retrieving a QR code for WhatsApp authentication
// @Summary Get QR code for WhatsApp authentication
// @Description Generate a QR code for linking a WhatsApp account to the session. Returns PNG file download by default, or JSON if format=json.
// @Tags WhatsApp
// @Accept json
// @Produce image/png
// @Produce application/json
// @Param id path string true "Session ID"
// @Param timeout query int false "QR code timeout in seconds" default(60)
// @Param format query string false "Response format: 'image' (default) or 'json'" default(image) Enums(image, json)
// @Param size query int false "QR code image size in pixels (only for image format, 128-1024)" default(256) minimum(128) maximum(1024)
// @Success 200 {string} string "PNG image file (binary)" 
// @Header 200 {string} Content-Type "image/png"
// @Header 200 {string} Content-Disposition "attachment; filename=\"whatsapp-qr-{session_id}.png\""
// @Success 200 {object} response.QRCodeResponse "QR code data (JSON)"
// @Failure 400 {object} response.ErrorResponse "Invalid session ID or parameters"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - requires whatsapp:read permission or session ownership"
// @Failure 404 {object} response.ErrorResponse "Session not found"
// @Failure 409 {object} response.ErrorResponse "Session is already connected"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /whatsapp/sessions/{id}/qr [get]
func (h *WhatsAppHandler) GetQRCode(w http.ResponseWriter, r *http.Request) {
	// Get session ID from URL
	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		respondError(w, http.StatusBadRequest, response.ErrCodeValidation, "Session ID is required", nil)
		return
	}

	// Get user ID from context
	userID := r.Context().Value("user_id")
	if userID == nil {
		respondError(w, http.StatusUnauthorized, response.ErrCodeUnauthorized, "User ID not found in context", nil)
		return
	}

	// Check if user is admin
	isAdmin := false
	if adminVal := r.Context().Value("is_admin"); adminVal != nil {
		isAdmin = adminVal.(bool)
	}

	// Parse format from query parameter
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "image" // Default to image
	}
	if format != "image" && format != "json" {
		respondError(w, http.StatusBadRequest, response.ErrCodeValidation, "Format must be 'image' or 'json'", nil)
		return
	}

	// Parse size from query parameter (only for image format)
	size := 256 // Default size
	if format == "image" {
		sizeStr := r.URL.Query().Get("size")
		if sizeStr != "" {
			if parsedSize, err := strconv.Atoi(sizeStr); err == nil && parsedSize >= 128 && parsedSize <= 1024 {
				size = parsedSize
			}
		}
	}

	// Parse timeout from query parameter
	timeoutStr := r.URL.Query().Get("timeout")
	timeout := 60 * time.Second // Default 60 seconds
	if timeoutStr != "" {
		if timeoutSec, err := strconv.Atoi(timeoutStr); err == nil && timeoutSec > 0 {
			timeout = time.Duration(timeoutSec) * time.Second
		}
	}

	// Validate timeout (max 5 minutes)
	if timeout > 5*time.Minute {
		respondError(w, http.StatusBadRequest, response.ErrCodeValidation, "Timeout cannot exceed 5 minutes", nil)
		return
	}

	// Call service to generate QR code
	qrCode, err := h.whatsappService.GenerateQRCode(r.Context(), sessionID, userID.(string), isAdmin, service.GenerateQRCodeInput{
		Timeout: timeout,
	})

	if err != nil {
		handleWhatsAppError(w, err)
		return
	}

	// Debug: Check QR code string
	if qrCode == "" {
		respondError(w, http.StatusInternalServerError, response.ErrCodeInternal, "QR code is empty", nil)
		return
	}

	// Get session to retrieve status
	session, err := h.whatsappService.GetSession(r.Context(), sessionID, userID.(string), isAdmin)
	if err != nil {
		respondError(w, http.StatusInternalServerError, response.ErrCodeInternal, "Failed to get session status", nil)
		return
	}

	// Return response based on format
	if format == "image" {
		// Generate QR code image
		qrImage, err := qrcode.Encode(qrCode, qrcode.Medium, size)
		if err != nil {
			respondError(w, http.StatusInternalServerError, response.ErrCodeInternal, fmt.Sprintf("Failed to generate QR code image: %v", err), nil)
			return
		}

		// Debug: Check image size
		if len(qrImage) == 0 {
			respondError(w, http.StatusInternalServerError, response.ErrCodeInternal, "Generated QR image is empty", nil)
			return
		}

		// Set headers for file download
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"whatsapp-qr-%s.png\"", sessionID))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(qrImage)))
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("X-QR-Expires-At", time.Now().Add(timeout).Format(time.RFC3339))
		w.Header().Set("X-Session-ID", sessionID)
		w.Header().Set("X-Session-Status", string(session.Status))
		
		// Write image
		w.WriteHeader(http.StatusOK)
		w.Write(qrImage)
	} else {
		// Return JSON response
		qrResponse := response.QRCodeResponse{
			SessionID: sessionID,
			QRCode:    qrCode,
			ExpiresAt: time.Now().Add(timeout),
			Status:    string(session.Status),
		}
		respondSuccess(w, http.StatusOK, "QR code generated successfully", qrResponse)
	}
}

// GetPairingCode handles generating and retrieving a pairing code for WhatsApp authentication
// @Summary Get pairing code for WhatsApp authentication
// @Description Generate an 8-digit pairing code for linking a WhatsApp account via phone number. Alternative to QR code scanning.
// @Tags WhatsApp
// @Accept json
// @Produce application/json
// @Param id path string true "Session ID"
// @Param phone_number query string true "Phone number with country code (e.g., 628123456789)"
// @Param timeout query int false "Pairing code timeout in seconds" default(60)
// @Success 200 {object} response.PairingCodeResponse "Pairing code generated successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid session ID or parameters"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - requires whatsapp:read permission or session ownership"
// @Failure 404 {object} response.ErrorResponse "Session not found"
// @Failure 409 {object} response.ErrorResponse "Session is already connected"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /whatsapp/sessions/{id}/pairing-code [get]
func (h *WhatsAppHandler) GetPairingCode(w http.ResponseWriter, r *http.Request) {
	// Get session ID from URL
	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		respondError(w, http.StatusBadRequest, response.ErrCodeValidation, "Session ID is required", nil)
		return
	}

	// Get user ID from context
	userID := r.Context().Value("user_id")
	if userID == nil {
		respondError(w, http.StatusUnauthorized, response.ErrCodeUnauthorized, "User ID not found in context", nil)
		return
	}

	// Check if user is admin
	isAdmin := false
	if adminVal := r.Context().Value("is_admin"); adminVal != nil {
		isAdmin = adminVal.(bool)
	}

	// Parse phone number from query parameter
	phoneNumber := r.URL.Query().Get("phone_number")
	if phoneNumber == "" {
		respondError(w, http.StatusBadRequest, response.ErrCodeValidation, "Phone number is required", nil)
		return
	}

	// Parse timeout from query parameter
	timeoutStr := r.URL.Query().Get("timeout")
	timeout := 60 * time.Second // Default 60 seconds
	if timeoutStr != "" {
		if timeoutSec, err := strconv.Atoi(timeoutStr); err == nil && timeoutSec > 0 {
			timeout = time.Duration(timeoutSec) * time.Second
		}
	}

	// Validate timeout (max 5 minutes)
	if timeout > 5*time.Minute {
		respondError(w, http.StatusBadRequest, response.ErrCodeValidation, "Timeout cannot exceed 5 minutes", nil)
		return
	}

	// Call service to generate pairing code
	pairingCode, err := h.whatsappService.GeneratePairingCode(r.Context(), sessionID, userID.(string), isAdmin, service.GeneratePairingCodeInput{
		PhoneNumber: phoneNumber,
		Timeout:     timeout,
	})

	if err != nil {
		handleWhatsAppError(w, err)
		return
	}

	// Debug: Check pairing code string
	if pairingCode == "" {
		respondError(w, http.StatusInternalServerError, response.ErrCodeInternal, "Pairing code is empty", nil)
		return
	}

	// Get session to retrieve status
	session, err := h.whatsappService.GetSession(r.Context(), sessionID, userID.(string), isAdmin)
	if err != nil {
		respondError(w, http.StatusInternalServerError, response.ErrCodeInternal, "Failed to get session status", nil)
		return
	}

	// Prepare response
	pairingResponse := response.PairingCodeResponse{
		SessionID:   sessionID,
		PairingCode: pairingCode,
		PhoneNumber: phoneNumber,
		ExpiresAt:   time.Now().Add(timeout),
		Status:      string(session.Status),
	}

	// Return response
	respondSuccess(w, http.StatusOK, "Pairing code generated successfully", pairingResponse)
}

// SendTextMessage handles sending a text message through WhatsApp
// @Summary Send text message
// @Description Send a text message through WhatsApp to a recipient
// @Tags WhatsApp
// @Accept json
// @Produce json
// @Param id path string true "Session ID"
// @Param request body request.SendTextMessageRequest true "Message data"
// @Success 200 {object} response.MessageSentResponse "Message sent successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid input"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - requires whatsapp:write permission"
// @Failure 404 {object} response.ErrorResponse "Session not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /whatsapp/sessions/{id}/messages [post]
func (h *WhatsAppHandler) SendTextMessage(w http.ResponseWriter, r *http.Request) {
	// Get session ID from URL
	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		respondError(w, http.StatusBadRequest, response.ErrCodeValidation, "Session ID is required", nil)
		return
	}

	// Get user ID from context
	userID := r.Context().Value("user_id")
	if userID == nil {
		respondError(w, http.StatusUnauthorized, response.ErrCodeUnauthorized, "User ID not found in context", nil)
		return
	}

	// Check if user is admin
	isAdmin := false
	if adminVal := r.Context().Value("is_admin"); adminVal != nil {
		isAdmin = adminVal.(bool)
	}

	// Parse request body
	var req request.SendTextMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, response.ErrCodeValidation, "Invalid request body", nil)
		return
	}

	// Validate input
	v := validator.New()
	v.ValidateRequired("to", req.To)
	v.ValidateRequired("content", req.Content)
	v.ValidateMaxLength("content", req.Content, 65536)

	if v.HasErrors() {
		respondValidationError(w, v.Errors())
		return
	}

	// Call service to send message
	message, err := h.whatsappService.SendTextMessage(r.Context(), sessionID, userID.(string), isAdmin, service.SendTextMessageInput{
		To:      req.To,
		Content: req.Content,
	})

	if err != nil {
		handleWhatsAppError(w, err)
		return
	}

	// Prepare response
	messageResponse := response.MessageSentResponse{
		MessageID: message.MessageID,
		Status:    string(message.Status),
		Timestamp: message.Timestamp,
		Message:   toWhatsAppMessageResponse(message),
	}

	// Return response
	respondSuccess(w, http.StatusOK, "Message sent successfully", messageResponse)
}

// SendMediaMessage handles sending a media message through WhatsApp
// @Summary Send media message
// @Description Send a media message (image, video, audio, document) through WhatsApp
// @Tags WhatsApp
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Session ID"
// @Param to formData string true "Recipient phone number (E.164 format)"
// @Param caption formData string false "Optional caption for media"
// @Param filename formData string false "Filename (required for documents)"
// @Param file formData file true "Media file"
// @Success 200 {object} response.MessageSentResponse "Message sent successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid input"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - requires whatsapp:write permission"
// @Failure 404 {object} response.ErrorResponse "Session not found"
// @Failure 413 {object} response.ErrorResponse "File too large"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /whatsapp/sessions/{id}/media [post]
func (h *WhatsAppHandler) SendMediaMessage(w http.ResponseWriter, r *http.Request) {
	// Get session ID from URL
	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		respondError(w, http.StatusBadRequest, response.ErrCodeValidation, "Session ID is required", nil)
		return
	}

	// Get user ID from context
	userID := r.Context().Value("user_id")
	if userID == nil {
		respondError(w, http.StatusUnauthorized, response.ErrCodeUnauthorized, "User ID not found in context", nil)
		return
	}

	// Check if user is admin
	isAdmin := false
	if adminVal := r.Context().Value("is_admin"); adminVal != nil {
		isAdmin = adminVal.(bool)
	}

	// Parse multipart form (max 32MB in memory)
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		respondError(w, http.StatusBadRequest, response.ErrCodeValidation, "Failed to parse multipart form", nil)
		return
	}

	// Get form values
	to := r.FormValue("to")
	caption := r.FormValue("caption")
	fileName := r.FormValue("filename")

	// Validate required fields
	v := validator.New()
	v.ValidateRequired("to", to)

	if v.HasErrors() {
		respondValidationError(w, v.Errors())
		return
	}

	// Get uploaded file
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, response.ErrCodeValidation, "File is required", nil)
		return
	}
	defer file.Close()

	// Read file data
	fileData := make([]byte, fileHeader.Size)
	_, err = file.Read(fileData)
	if err != nil {
		respondError(w, http.StatusInternalServerError, response.ErrCodeInternal, "Failed to read file", nil)
		return
	}

	// Get MIME type from file header
	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType == "" {
		respondError(w, http.StatusBadRequest, response.ErrCodeValidation, "Could not determine file MIME type", nil)
		return
	}

	// Determine media type from MIME type
	mediaType := determineMediaType(mimeType)
	if mediaType == "" {
		respondError(w, http.StatusBadRequest, response.ErrCodeValidation, "Unsupported media type", nil)
		return
	}

	// If filename not provided, use the uploaded filename
	if fileName == "" {
		fileName = fileHeader.Filename
	}

	// Validate file size (default 16MB, configurable)
	maxSizeMB := 16
	if fileHeader.Size > int64(maxSizeMB*1024*1024) {
		respondError(w, http.StatusRequestEntityTooLarge, response.ErrCodeValidation, 
			fmt.Sprintf("File size exceeds maximum allowed size of %dMB", maxSizeMB), nil)
		return
	}

	// Call service to send media message
	message, err := h.whatsappService.SendMediaMessage(r.Context(), sessionID, userID.(string), isAdmin, service.SendMediaMessageInput{
		To: to,
		Media: service.MediaMessage{
			Type:     service.MediaType(mediaType),
			Data:     fileData,
			MimeType: mimeType,
			Caption:  caption,
			FileName: fileName,
		},
		MaxSizeMB: maxSizeMB,
	})

	if err != nil {
		handleWhatsAppError(w, err)
		return
	}

	// Prepare response
	messageResponse := response.MessageSentResponse{
		MessageID: message.MessageID,
		Status:    string(message.Status),
		Timestamp: message.Timestamp,
		Message:   toWhatsAppMessageResponse(message),
	}

	// Return response
	respondSuccess(w, http.StatusOK, "Media message sent successfully", messageResponse)
}

// GetMessages handles retrieving message history for a session
// @Summary Get message history
// @Description Retrieve paginated message history for a WhatsApp session
// @Tags WhatsApp
// @Accept json
// @Produce json
// @Param id path string true "Session ID"
// @Param direction query string false "Filter by direction (inbound, outbound)"
// @Param start_date query string false "Start date (ISO 8601 format)"
// @Param end_date query string false "End date (ISO 8601 format)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page (max 100)" default(50)
// @Success 200 {object} response.PaginatedResponse "Message history"
// @Failure 400 {object} response.ErrorResponse "Invalid parameters"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - requires whatsapp:read permission"
// @Failure 404 {object} response.ErrorResponse "Session not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /whatsapp/sessions/{id}/messages [get]
func (h *WhatsAppHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	// Get session ID from URL
	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		respondError(w, http.StatusBadRequest, response.ErrCodeValidation, "Session ID is required", nil)
		return
	}

	// Get user ID from context
	userID := r.Context().Value("user_id")
	if userID == nil {
		respondError(w, http.StatusUnauthorized, response.ErrCodeUnauthorized, "User ID not found in context", nil)
		return
	}

	// Check if user is admin
	isAdmin := false
	if adminVal := r.Context().Value("is_admin"); adminVal != nil {
		isAdmin = adminVal.(bool)
	}

	// Parse query parameters
	directionStr := r.URL.Query().Get("direction")
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	// Parse page and limit
	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Parse direction
	var direction entity.MessageDirection
	if directionStr != "" {
		direction = entity.MessageDirection(directionStr)
		// Validate direction
		if direction != entity.MessageDirectionInbound && direction != entity.MessageDirectionOutbound && direction != "" {
			respondError(w, http.StatusBadRequest, response.ErrCodeValidation, "Invalid direction. Must be 'inbound' or 'outbound'", nil)
			return
		}
	}

	// Parse dates
	var startDate, endDate *time.Time
	if startDateStr != "" {
		parsedDate, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, response.ErrCodeValidation, "Invalid start_date format. Use ISO 8601 format", nil)
			return
		}
		startDate = &parsedDate
	}

	if endDateStr != "" {
		parsedDate, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			respondError(w, http.StatusBadRequest, response.ErrCodeValidation, "Invalid end_date format. Use ISO 8601 format", nil)
			return
		}
		endDate = &parsedDate
	}

	// Call service to get messages
	result, err := h.whatsappService.GetMessages(r.Context(), sessionID, userID.(string), isAdmin, service.GetMessagesInput{
		Direction: direction,
		StartDate: startDate,
		EndDate:   endDate,
		Page:      page,
		Limit:     limit,
	})

	if err != nil {
		handleWhatsAppError(w, err)
		return
	}

	// Convert messages to response format
	messageResponses := make([]response.WhatsAppMessageResponse, len(result.Messages))
	for i, message := range result.Messages {
		messageResponses[i] = toWhatsAppMessageResponse(message)
	}

	// Return paginated response
	respondJSON(w, http.StatusOK, response.NewSuccessPaginatedResponse(
		"Messages retrieved successfully",
		messageResponses,
		response.Pagination{
			Total:      result.Total,
			Page:       result.Page,
			Limit:      result.Limit,
			TotalPages: result.TotalPages,
		},
	))
}

// determineMediaType determines the media type from MIME type
func determineMediaType(mimeType string) string {
	switch {
	case contains(mimeType, "image/"):
		return "image"
	case contains(mimeType, "video/"):
		return "video"
	case contains(mimeType, "audio/"):
		return "audio"
	case contains(mimeType, "application/") || contains(mimeType, "text/"):
		return "document"
	default:
		return ""
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}

// toWhatsAppSessionResponse converts a WhatsAppSession entity to a response
func toWhatsAppSessionResponse(session *entity.WhatsAppSession) response.WhatsAppSessionResponse {
	deviceInfo := ""
	if session.DeviceInfo != nil {
		deviceInfo = *session.DeviceInfo
	}
	
	return response.WhatsAppSessionResponse{
		ID:           session.ID,
		UserID:       session.UserID,
		SessionName:  session.SessionName,
		PhoneNumber:  session.PhoneNumber,
		Status:       string(session.Status),
		QRCode:       safeString(session.QRCode),
		QRCodeExpiry: session.QRCodeExpiry,
		DeviceInfo:   deviceInfo,
		LastActivity: session.LastActivity,
		WebhookURL:   session.WebhookURL,
		IsActive:     session.IsActive,
		CreatedAt:    session.CreatedAt,
		UpdatedAt:    session.UpdatedAt,
	}
}

// toWhatsAppMessageResponse converts a WhatsAppMessage entity to a response
func toWhatsAppMessageResponse(message *entity.WhatsAppMessage) response.WhatsAppMessageResponse {
	return response.WhatsAppMessageResponse{
		ID:               message.ID,
		SessionID:        message.SessionID,
		MessageID:        message.MessageID,
		Direction:        string(message.Direction),
		Status:           string(message.Status),
		From:             message.From,
		To:               message.To,
		Content:          message.Content,
		MessageType:      message.MessageType,
		Timestamp:        message.Timestamp,
		WebhookDelivered: message.WebhookDelivered,
		WebhookAttempts:  message.WebhookAttempts,
		ErrorMessage:     message.ErrorMessage,
		CreatedAt:        message.CreatedAt,
		UpdatedAt:        message.UpdatedAt,
	}
}

// ReconnectSession handles reconnecting a disconnected WhatsApp session
// @Summary Reconnect WhatsApp session
// @Description Attempt to reconnect a disconnected WhatsApp session
// @Tags WhatsApp
// @Accept json
// @Produce json
// @Param id path string true "Session ID"
// @Success 200 {object} response.BaseResponse "Reconnection initiated successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid session ID or session already connected"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - requires whatsapp:write permission or session ownership"
// @Failure 404 {object} response.ErrorResponse "Session not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /whatsapp/sessions/{id}/reconnect [post]
func (h *WhatsAppHandler) ReconnectSession(w http.ResponseWriter, r *http.Request) {
	// Get session ID from URL
	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		respondError(w, http.StatusBadRequest, response.ErrCodeValidation, "Session ID is required", nil)
		return
	}

	// Get user ID from context
	userID := r.Context().Value("user_id")
	if userID == nil {
		respondError(w, http.StatusUnauthorized, response.ErrCodeUnauthorized, "User ID not found in context", nil)
		return
	}

	// Check if user is admin
	isAdmin := false
	if adminVal := r.Context().Value("is_admin"); adminVal != nil {
		isAdmin = adminVal.(bool)
	}

	// Call service to reconnect session
	err := h.whatsappService.ReconnectSession(r.Context(), sessionID, userID.(string), isAdmin)
	if err != nil {
		handleWhatsAppError(w, err)
		return
	}

	// Return response
	respondSuccess(w, http.StatusOK, "Session reconnection initiated successfully", nil)
}

// GetSessionStatus handles retrieving the current status of a WhatsApp session
// @Summary Get session status
// @Description Get the current connection and authentication status of a WhatsApp session
// @Tags WhatsApp
// @Accept json
// @Produce json
// @Param id path string true "Session ID"
// @Success 200 {object} response.SessionStatusResponse "Session status retrieved successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid session ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - requires whatsapp:read permission or session ownership"
// @Failure 404 {object} response.ErrorResponse "Session not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /whatsapp/sessions/{id}/status [get]
func (h *WhatsAppHandler) GetSessionStatus(w http.ResponseWriter, r *http.Request) {
	// Get session ID from URL
	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		respondError(w, http.StatusBadRequest, response.ErrCodeValidation, "Session ID is required", nil)
		return
	}

	// Get user ID from context
	userID := r.Context().Value("user_id")
	if userID == nil {
		respondError(w, http.StatusUnauthorized, response.ErrCodeUnauthorized, "User ID not found in context", nil)
		return
	}

	// Check if user is admin
	isAdmin := false
	if adminVal := r.Context().Value("is_admin"); adminVal != nil {
		isAdmin = adminVal.(bool)
	}

	// Call service to get session status
	statusInfo, err := h.whatsappService.GetSessionStatus(r.Context(), sessionID, userID.(string), isAdmin)
	if err != nil {
		handleWhatsAppError(w, err)
		return
	}

	// Convert to response format
	statusResponse := response.SessionStatusResponse{
		SessionID:      statusInfo.SessionID,
		Status:         string(statusInfo.Status),
		PhoneNumber:    statusInfo.PhoneNumber,
		ConnectedPhone: statusInfo.ConnectedPhone,
		LastActivity:   statusInfo.LastActivity,
		IsActive:       statusInfo.IsActive,
		IsConnected:    statusInfo.IsConnected,
		IsLoggedIn:     statusInfo.IsLoggedIn,
	}

	// Return response
	respondSuccess(w, http.StatusOK, "Session status retrieved successfully", statusResponse)
}
