package serverhandlers

import (
	"errors"
	"fmt"
	"time"

	"github.com/DanilNaum/secret-storage-client/internal/models"
)

// StorageReader defines the interface for reading storage data.
type StorageReader interface {
	GetLocalRecords() []*models.Record
	GetServerRecords() []*models.ServerRecord
}

// StorageWriter defines the interface for writing storage data.
type StorageWriter interface {
	AddLocalRecord(record *models.Record)
	AddServerRecord(serverRecord *models.ServerRecord)
	RemoveServerRecord(index int) *models.ServerRecord
	FindLocalRecordByServerID(serverID string) *models.Record
	FindServerRecordByID(serverID string) *models.ServerRecord
}

// CacheManager defines the interface for cache operations.
type CacheManager interface {
	Get(id string) (*models.Record, bool, time.Time)
	Set(id string, record *models.Record)
	Remove(id string)
	IsExpired(id string) bool
}

// StorageAdapter combines reader and writer interfaces.
type StorageAdapter interface {
	StorageReader
	StorageWriter
}

// ServerHandlers defines all event handlers for server record operations.
type ServerHandlers struct {
	storage         StorageReader
	writer          StorageWriter
	cache           CacheManager
	serverURL       string
	authToken       string
	isAuthenticated bool
	masterPassword  string
	salt            string
}

// NewServerHandlers creates server handlers with storage and cache adapters.
func NewServerHandlers(storage StorageAdapter, cache CacheManager, serverURL string) *ServerHandlers {
	return &ServerHandlers{
		storage:         storage,
		writer:          storage,
		cache:           cache,
		serverURL:       serverURL,
		authToken:       "",
		isAuthenticated: false,
		masterPassword:  "",
		salt:            "",
	}
}

// Authenticate performs user authentication with the server.
func (h *ServerHandlers) Authenticate(username, password string) error {
	if username == "admin" && password == "password" {
		h.authToken = "simulated_token_12345"
		h.isAuthenticated = true
		return nil
	}
	return errors.New("invalid credentials")
}

// Register performs user registration with the server.
func (h *ServerHandlers) Register(username, password string) (string, error) {
	if username == "" || password == "" {
		return "", errors.New("username and password cannot be empty")
	}

	salt := fmt.Sprintf("salt_%s_%d", username, time.Now().Unix())
	h.authToken = "simulated_token_12345"
	h.isAuthenticated = true
	h.salt = salt

	return salt, nil
}

// SetMasterPassword sets the master password for encryption.
func (h *ServerHandlers) SetMasterPassword(password string) {
	h.masterPassword = password
}


// Logout performs user logout.
func (h *ServerHandlers) Logout() error {
	h.authToken = ""
	h.isAuthenticated = false
	h.masterPassword = ""
	h.salt = ""
	return nil
}

// IsAuthenticated returns the authentication status.
func (h *ServerHandlers) IsAuthenticated() bool {
	return h.isAuthenticated
}


// GetCachedRecord retrieves a record from cache with metadata.
func (h *ServerHandlers) GetCachedRecord(id string) (*models.Record, bool, time.Time, bool) {
	record, found, cachedAt := h.cache.Get(id)
	expired := h.cache.IsExpired(id)
	return record, found, cachedAt, expired
}

// SetCachedRecord stores a record in cache.
func (h *ServerHandlers) SetCachedRecord(id string, record *models.Record) {
	h.cache.Set(id, record)
}




// DeleteServerRecord deletes a server record.
func (h *ServerHandlers) DeleteServerRecord(record *models.Record) error {
	if !h.isAuthenticated {
		return errors.New("not authenticated")
	}

	serverRecords := h.storage.GetServerRecords()
	for i, r := range serverRecords {
		if r.ID == record.ServerID {
			h.writer.RemoveServerRecord(i)
			h.cache.Remove(record.ServerID)

			if localRecord := h.writer.FindLocalRecordByServerID(record.ServerID); localRecord != nil {
				localRecord.ServerID = ""
			}

			return nil
		}
	}
	return errors.New("server record not found")
}

// ServerRecordChange updates an existing server record.
func (h *ServerHandlers) ServerRecordChange(record *models.Record) error {
	if !h.isAuthenticated {
		return errors.New("not authenticated")
	}

	h.cache.Set(record.ServerID, record)

	if localRecord := h.writer.FindLocalRecordByServerID(record.ServerID); localRecord != nil {
		localRecord.Name = record.Name
		localRecord.Type = record.Type
		localRecord.Username = record.Username
		localRecord.Password = record.Password
		localRecord.TextContent = record.TextContent
		localRecord.FilePath = record.FilePath
	}

	return errors.New("server record modification not implemented")
}

// ServerRecordOpen loads a server record by ID.
func (h *ServerHandlers) ServerRecordOpen(serverID string) (*models.Record, error) {
	if !h.isAuthenticated {
		return nil, errors.New("not authenticated")
	}

	if cachedRecord, found, _ := h.cache.Get(serverID); found {
		return cachedRecord, nil
	}

	return nil, errors.New("server record loading not implemented")
}

// RefreshServerRecord forces a refresh of a server record from the server.
func (h *ServerHandlers) RefreshServerRecord(serverID string) (*models.Record, error) {
	if !h.isAuthenticated {
		return nil, errors.New("not authenticated")
	}

	return nil, errors.New("server record refresh not implemented")
}

// ServerRecordMovedToLocal downloads a server record to local storage.
func (h *ServerHandlers) ServerRecordMovedToLocal(record *models.Record) error {
	if !h.isAuthenticated {
		return errors.New("not authenticated")
	}

	if existingRecord := h.writer.FindLocalRecordByServerID(record.ServerID); existingRecord != nil {
		existingRecord.Name = record.Name
		existingRecord.Type = record.Type
		existingRecord.Username = record.Username
		existingRecord.Password = record.Password
		existingRecord.TextContent = record.TextContent
		existingRecord.FilePath = record.FilePath
	} else {
		newRecord := record.Clone()
		newRecord.ID = fmt.Sprintf("local_%d", len(h.storage.GetLocalRecords())+1)
		newRecord.ServerID = record.ServerID
		newRecord.IsServer = false
		h.writer.AddLocalRecord(newRecord)
	}

	return nil
}

// Sync performs full synchronization between local and server storage.
func (h *ServerHandlers) Sync() error {
	if !h.isAuthenticated {
		return errors.New("not authenticated")
	}

	localRecords := h.storage.GetLocalRecords()
	for _, record := range localRecords {
		if !record.HasServerID() {
			serverID := fmt.Sprintf("server_%s", record.ID)

			serverRecord := &models.ServerRecord{
				ID:   serverID,
				Name: record.Name,
				Type: record.Type,
			}
			h.writer.AddServerRecord(serverRecord)
		}
	}

	return nil
}

// GetServerURL returns the server URL.
func (h *ServerHandlers) GetServerURL() string {
	return h.serverURL
}

// SetServerURL sets the server URL.
func (h *ServerHandlers) SetServerURL(url string) {
	h.serverURL = url
}