package media

import (
	"golang-chi-hexagonal-base/internal/pkg/media"
)

// ManagerAdapter adapts the media.Manager to the service layer interface
type ManagerAdapter struct {
	manager *media.Manager
}

// NewManagerAdapter creates a new ManagerAdapter instance
func NewManagerAdapter(manager *media.Manager) *ManagerAdapter {
	return &ManagerAdapter{
		manager: manager,
	}
}

// DeleteSessionMedia deletes all media files for a session
func (a *ManagerAdapter) DeleteSessionMedia(sessionID string) error {
	return a.manager.DeleteSessionMedia(sessionID)
}
