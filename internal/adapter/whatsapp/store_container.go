package whatsapp

import (
	"context"
	"sync"
	
	"go.mau.fi/whatsmeow/store"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// StoreContainer implements the store.Container interface for whatsmeow
// This allows whatsmeow to use our PostgreSQL store implementations
type StoreContainer struct {
	store  *PostgresStore
	logger waLog.Logger
	mu     sync.RWMutex
	device *store.Device
}

// GetFirstDevice returns the first device from the container
// For our use case, we create a new device that uses the PostgreSQL store
func (c *StoreContainer) GetFirstDevice() (*store.Device, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.device == nil {
		device := c.store.NewDevice()
		device.Log = c.logger
		device.Container = c
		c.device = device
	}
	return c.device, nil
}

// GetAllDevices returns all devices in the container
// For our use case, we only have one device per container
func (c *StoreContainer) GetAllDevices() ([]*store.Device, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if c.device == nil {
		c.mu.RUnlock()
		_, err := c.GetFirstDevice()
		c.mu.RLock()
		if err != nil {
			return nil, err
		}
	}
	return []*store.Device{c.device}, nil
}

// NewDevice creates a new device in the container
// For our use case, this returns a new device using the PostgreSQL store
func (c *StoreContainer) NewDevice() (*store.Device, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.device == nil {
		device := c.store.NewDevice()
		device.Log = c.logger
		device.Container = c
		c.device = device
	}
	return c.device, nil
}

// PutDevice saves a device to the container
// For our PostgreSQL implementation, device data is automatically saved
// through the store interface methods
func (c *StoreContainer) PutDevice(ctx context.Context, device *store.Device) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Update the cached device
	c.device = device
	
	// Save device to database
	if device != nil && device.ID != nil {
		err := c.store.saveDeviceToDB(ctx, device)
		if err != nil {
			c.logger.Errorf("Failed to save device to database: %v", err)
			return err
		}
	}
	
	return nil
}

// DeleteDevice removes a device from the container
// For our PostgreSQL implementation, we would need to clean up the device data
func (c *StoreContainer) DeleteDevice(ctx context.Context, device *store.Device) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Clear cached device
	c.device = nil
	
	// Clean up would be handled by deleting the session
	// which cascades to all related keys and data
	return nil
}

// DeleteAllDevices removes all devices from the container
func (c *StoreContainer) DeleteAllDevices() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Clear cached device
	c.device = nil
	
	// Would delete all session data
	// Not typically needed in our use case
	return nil
}

// NewStoreContainer creates a new store container with proper initialization
func NewStoreContainer(store *PostgresStore, logger waLog.Logger) *StoreContainer {
	return &StoreContainer{
		store:  store,
		logger: logger,
		device: nil,
	}
}
