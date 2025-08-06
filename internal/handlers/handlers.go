// Package handlers provides the business logic layer for record management.
// It defines interfaces for storage and cache operations, implements event handlers
// for all record operations, and manages the interaction between UI, storage, and cache layers.
// This package follows the clean architecture principles with dependency inversion.
package handlers

import (
	"errors"
	"fmt"
	"time"

	"github.com/DanilNaum/secret-storage-client/internal/models"
)

// StorageReader defines the interface for reading storage data.
// This interface is used by handlers to access stored records without
// the ability to modify them, following the principle of least privilege.
type StorageReader interface {
	// GetLocalRecords returns all local records from storage.
	GetLocalRecords() []*models.Record
	// GetServerRecords returns all server record metadata from storage.
	GetServerRecords() []*models.ServerRecord
}

// StorageWriter defines the interface for writing storage data.
// This interface provides all the write operations needed by handlers
// to manage both local records and server record metadata.
type StorageWriter interface {
	// AddLocalRecord adds a new record to local storage.
	AddLocalRecord(record *models.Record)
	// RemoveLocalRecord removes a local record by index and returns it.
	RemoveLocalRecord(index int) *models.Record
	// AddServerRecord adds server record metadata to storage.
	AddServerRecord(serverRecord *models.ServerRecord)
	// RemoveServerRecord removes server record metadata by index and returns it.
	RemoveServerRecord(index int) *models.ServerRecord
	// CopyToLocal creates a local copy of a server record.
	CopyToLocal(serverRecord *models.Record)
	// UpdateLocalRecordServerID updates the server ID of a local record.
	UpdateLocalRecordServerID(localID, serverID string) bool
	// FindLocalRecordByServerID finds a local record by its server ID.
	FindLocalRecordByServerID(serverID string) *models.Record
	// FindServerRecordByID finds server record metadata by server ID.
	FindServerRecordByID(serverID string) *models.ServerRecord
}

// CacheManager defines the interface for cache operations.
// This interface provides caching functionality for server records
// to improve performance and reduce server load.
// It's designed to work with the generic cache implementation.
type CacheManager interface {
	// Get retrieves a record from cache with metadata about when it was cached.
	Get(id string) (*models.Record, bool, time.Time)
	// Set stores a record in cache with default TTL.
	Set(id string, record *models.Record)
	// Remove removes a record from cache.
	Remove(id string)
	// IsExpired checks if a cached record has expired.
	IsExpired(id string) bool
}

// EventHandlers defines all event handlers for the application.
// This struct contains the business logic for all record operations
// and manages the interaction between storage, cache, and external systems.
// It implements the event-driven architecture pattern.
type EventHandlers struct {
	// storage provides read access to stored data.
	storage StorageReader
	// writer provides write access to stored data.
	writer StorageWriter
	// cache manages server record caching using the generic cache.
	cache CacheManager

	// Local record event handlers

	// OnCreateRecord handles the creation of new local records.
	OnCreateRecord func(record *models.Record) error
	// OnDeleteLocalRecord handles the deletion of local records.
	OnDeleteLocalRecord func(record *models.Record) error
	// OnLocalRecordChange handles updates to existing local records.
	OnLocalRecordChange func(record *models.Record) error
	// OnLocalRecordOpen handles opening/loading local records by ID.
	OnLocalRecordOpen func(id string) (*models.Record, error)
	// OnLocalRecordMovedToServer handles uploading local records to server.
	OnLocalRecordMovedToServer func(record *models.Record) error

	// Server record event handlers

	// OnDeleteServerRecord handles the deletion of server records.
	OnDeleteServerRecord func(record *models.Record) error
	// OnServerRecordChange handles updates to existing server records.
	OnServerRecordChange func(record *models.Record) error
	// OnServerRecordOpen handles loading server records by ID.
	OnServerRecordOpen func(id string) (*models.Record, error)
	// OnServerRecordMovedToLocal handles downloading server records to local storage.
	OnServerRecordMovedToLocal func(record *models.Record) error
	// OnRefreshServerRecord handles forced refresh of server records from server.
	OnRefreshServerRecord func(id string) (*models.Record, error)

	// System event handlers

	// OnSync handles full synchronization between local and server storage.
	OnSync func() error
}

// StorageAdapter combines reader and writer interfaces.
// This interface is used for storage implementations that provide
// both read and write capabilities.
type StorageAdapter interface {
	StorageReader
	StorageWriter
}

// CacheAdapter defines the cache interface.
// This interface is used for cache implementations that provide
// the required caching functionality for server records.
// It's compatible with the generic cache implementation.
type CacheAdapter interface {
	CacheManager
}

// NewDefaultHandlers creates default handlers with storage and cache adapters.
// This function sets up the complete event handling system with default implementations
// for all record operations. The handlers implement the business logic for:
// - Local record management (create, read, update, delete)
// - Server record management with caching
// - Synchronization between local and server storage
// - Cache management and TTL handling
//
// The cache parameter should be a generic cache instance like cache.NewCache[models.Record]().
func NewDefaultHandlers(storage StorageAdapter, cache CacheAdapter) *EventHandlers {
	return &EventHandlers{
		storage: storage,
		writer:  storage,
		cache:   cache,

		// Local handlers - implement actual logic with storage
		OnCreateRecord: func(record *models.Record) error {
			storage.AddLocalRecord(record)
			return nil
		},

		OnDeleteLocalRecord: func(record *models.Record) error {
			// Find and remove the record by ID
			localRecords := storage.GetLocalRecords()
			for i, r := range localRecords {
				if r.ID == record.ID {
					storage.RemoveLocalRecord(i)
					return nil
				}
			}
			return errors.New("local record not found")
		},

		OnLocalRecordChange: func(record *models.Record) error {
			// Find and update the record by ID
			localRecords := storage.GetLocalRecords()
			for _, r := range localRecords {
				if r.ID == record.ID {
					// Update the record in place
					*r = *record
					return nil
				}
			}
			return errors.New("local record not found")
		},

		OnLocalRecordOpen: func(id string) (*models.Record, error) {
			localRecords := storage.GetLocalRecords()
			for _, r := range localRecords {
				if r.ID == id {
					return r, nil
				}
			}
			return nil, errors.New("local record not found")
		},

		OnLocalRecordMovedToServer: func(record *models.Record) error {
			// Simulate server upload and get server ID
			serverID := fmt.Sprintf("server_%s", record.ID)

			// Update local record with server ID
			storage.UpdateLocalRecordServerID(record.ID, serverID)

			// Add server record metadata
			serverRecord := &models.ServerRecord{
				ID:   serverID,
				Name: record.Name,
				Type: record.Type,
			}
			storage.AddServerRecord(serverRecord)

			return nil
		},

		// Server handlers
		OnDeleteServerRecord: func(record *models.Record) error {
			// Find and remove server record by server ID
			serverRecords := storage.GetServerRecords()
			for i, r := range serverRecords {
				if r.ID == record.ServerID {
					storage.RemoveServerRecord(i)
					// Remove from cache as well
					cache.Remove(record.ServerID)

					// Clear server ID from local record if exists
					if localRecord := storage.FindLocalRecordByServerID(record.ServerID); localRecord != nil {
						localRecord.ServerID = ""
					}

					return nil
				}
			}
			return errors.New("server record not found")
		},

		OnServerRecordChange: func(record *models.Record) error {
			// Update cache with new data
			cache.Set(record.ServerID, record)

			// Update local record if exists
			if localRecord := storage.FindLocalRecordByServerID(record.ServerID); localRecord != nil {
				localRecord.Name = record.Name
				localRecord.Type = record.Type
				localRecord.Username = record.Username
				localRecord.Password = record.Password
				localRecord.TextContent = record.TextContent
				localRecord.FilePath = record.FilePath
			}

			return errors.New("server record modification not implemented")
		},

		OnServerRecordOpen: func(serverID string) (*models.Record, error) {
			// Try to get from cache first
			if cachedRecord, found, _ := cache.Get(serverID); found {
				return cachedRecord, nil
			}

			// If not in cache or expired, return error (server not implemented)
			return nil, errors.New("server record loading not implemented")
		},

		OnRefreshServerRecord: func(serverID string) (*models.Record, error) {
			// Force refresh from server (bypass cache)
			// This would normally make a server request
			// For now, return error as server is not implemented
			return nil, errors.New("server record refresh not implemented")
		},

		OnServerRecordMovedToLocal: func(record *models.Record) error {
			// Check if local record with this server ID already exists
			if existingRecord := storage.FindLocalRecordByServerID(record.ServerID); existingRecord != nil {
				// Update existing local record
				existingRecord.Name = record.Name
				existingRecord.Type = record.Type
				existingRecord.Username = record.Username
				existingRecord.Password = record.Password
				existingRecord.TextContent = record.TextContent
				existingRecord.FilePath = record.FilePath
			} else {
				// Create new local record
				newRecord := record.Clone()
				newRecord.ID = fmt.Sprintf("local_%d", len(storage.GetLocalRecords())+1)
				newRecord.ServerID = record.ServerID
				newRecord.IsServer = false
				storage.AddLocalRecord(newRecord)
			}

			return nil
		},

		// Sync handler
		OnSync: func() error {
			// Sync all local records to server
			localRecords := storage.GetLocalRecords()
			for _, record := range localRecords {
				if !record.HasServerID() {
					// Simulate server upload
					serverID := fmt.Sprintf("server_%s", record.ID)
					storage.UpdateLocalRecordServerID(record.ID, serverID)

					// Add server record metadata
					serverRecord := &models.ServerRecord{
						ID:   serverID,
						Name: record.Name,
						Type: record.Type,
					}
					storage.AddServerRecord(serverRecord)
				}
			}

			return nil
		},
	}
}

// GetCachedRecord retrieves a record from cache with metadata.
// Returns the cached record, whether it was found, when it was cached,
// and whether it has expired. This method provides complete cache status
// information for UI display purposes.
func (h *EventHandlers) GetCachedRecord(id string) (*models.Record, bool, time.Time, bool) {
	record, found, cachedAt := h.cache.Get(id)
	expired := h.cache.IsExpired(id)
	return record, found, cachedAt, expired
}

// SetCachedRecord stores a record in cache.
// This method provides a way for external components to cache records,
// typically used after loading records from the server.
func (h *EventHandlers) SetCachedRecord(id string, record *models.Record) {
	h.cache.Set(id, record)
}

// RemoveCachedRecord removes a record from cache.
// This method is used for cache invalidation when records are deleted
// or when manual cache cleanup is needed.
func (h *EventHandlers) RemoveCachedRecord(id string) {
	h.cache.Remove(id)
}

// IsCacheExpired checks if a cached record is expired.
// Returns true if the record is not in cache or has exceeded its TTL.
// This method is used by the UI to display appropriate cache status indicators.
func (h *EventHandlers) IsCacheExpired(id string) bool {
	return h.cache.IsExpired(id)
}

// Method wrappers to implement UI interface
// These methods provide a clean interface for the UI layer by wrapping
// the function-based event handlers with proper error handling.

// CreateRecord creates a new local record.
// This method wraps the OnCreateRecord handler and provides
// a consistent interface for the UI layer.
func (h *EventHandlers) CreateRecord(record *models.Record) error {
	if h.OnCreateRecord != nil {
		return h.OnCreateRecord(record)
	}
	return errors.New("OnCreateRecord handler not set")
}

// DeleteLocalRecord deletes a local record.
// This method wraps the OnDeleteLocalRecord handler and provides
// a consistent interface for the UI layer.
func (h *EventHandlers) DeleteLocalRecord(record *models.Record) error {
	if h.OnDeleteLocalRecord != nil {
		return h.OnDeleteLocalRecord(record)
	}
	return errors.New("OnDeleteLocalRecord handler not set")
}

// LocalRecordChange updates an existing local record.
// This method wraps the OnLocalRecordChange handler and provides
// a consistent interface for the UI layer.
func (h *EventHandlers) LocalRecordChange(record *models.Record) error {
	if h.OnLocalRecordChange != nil {
		return h.OnLocalRecordChange(record)
	}
	return errors.New("OnLocalRecordChange handler not set")
}

// LocalRecordOpen loads a local record by ID.
// This method wraps the OnLocalRecordOpen handler and provides
// a consistent interface for the UI layer.
func (h *EventHandlers) LocalRecordOpen(id string) (*models.Record, error) {
	if h.OnLocalRecordOpen != nil {
		return h.OnLocalRecordOpen(id)
	}
	return nil, errors.New("OnLocalRecordOpen handler not set")
}

// LocalRecordMovedToServer uploads a local record to the server.
// This method wraps the OnLocalRecordMovedToServer handler and provides
// a consistent interface for the UI layer.
func (h *EventHandlers) LocalRecordMovedToServer(record *models.Record) error {
	if h.OnLocalRecordMovedToServer != nil {
		return h.OnLocalRecordMovedToServer(record)
	}
	return errors.New("OnLocalRecordMovedToServer handler not set")
}

// DeleteServerRecord deletes a server record.
// This method wraps the OnDeleteServerRecord handler and provides
// a consistent interface for the UI layer.
func (h *EventHandlers) DeleteServerRecord(record *models.Record) error {
	if h.OnDeleteServerRecord != nil {
		return h.OnDeleteServerRecord(record)
	}
	return errors.New("OnDeleteServerRecord handler not set")
}

// ServerRecordChange updates an existing server record.
// This method wraps the OnServerRecordChange handler and provides
// a consistent interface for the UI layer.
func (h *EventHandlers) ServerRecordChange(record *models.Record) error {
	if h.OnServerRecordChange != nil {
		return h.OnServerRecordChange(record)
	}
	return errors.New("OnServerRecordChange handler not set")
}

// ServerRecordOpen loads a server record by ID.
// This method wraps the OnServerRecordOpen handler and provides
// a consistent interface for the UI layer.
func (h *EventHandlers) ServerRecordOpen(id string) (*models.Record, error) {
	if h.OnServerRecordOpen != nil {
		return h.OnServerRecordOpen(id)
	}
	return nil, errors.New("OnServerRecordOpen handler not set")
}

// ServerRecordMovedToLocal downloads a server record to local storage.
// This method wraps the OnServerRecordMovedToLocal handler and provides
// a consistent interface for the UI layer.
func (h *EventHandlers) ServerRecordMovedToLocal(record *models.Record) error {
	if h.OnServerRecordMovedToLocal != nil {
		return h.OnServerRecordMovedToLocal(record)
	}
	return errors.New("OnServerRecordMovedToLocal handler not set")
}

// RefreshServerRecord forces a refresh of a server record from the server.
// This method wraps the OnRefreshServerRecord handler and provides
// a consistent interface for the UI layer.
func (h *EventHandlers) RefreshServerRecord(id string) (*models.Record, error) {
	if h.OnRefreshServerRecord != nil {
		return h.OnRefreshServerRecord(id)
	}
	return nil, errors.New("OnRefreshServerRecord handler not set")
}

// Sync performs full synchronization between local and server storage.
// This method wraps the OnSync handler and provides
// a consistent interface for the UI layer.
func (h *EventHandlers) Sync() error {
	if h.OnSync != nil {
		return h.OnSync()
	}
	return errors.New("OnSync handler not set")
}
