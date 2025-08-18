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
	// storage         StorageReader
	// writer          StorageWriter
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
		// storage:         storage,
		// writer:          storage,
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
func (h *ServerHandlers) Register(username, password string) error {
	if username == "" || password == "" {
		return errors.New("username and password cannot be empty")
	}

	salt := fmt.Sprintf("salt_%s_%d", username, time.Now().Unix())
	h.authToken = "simulated_token_12345"
	h.isAuthenticated = true
	h.salt = salt

	return nil
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

	return errors.New("not implemented")
}

// ServerRecordChange updates an existing server record.
func (h *ServerHandlers) ServerRecordChange(record *models.Record) error {
	if !h.isAuthenticated {
		return errors.New("not authenticated")
	}

	h.cache.Set(record.ServerID, record)

	
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



// Sync performs full synchronization between local and server storage.
func (h *ServerHandlers) Sync(localRecords []*models.Record) (map[string]string, error) {
	if !h.isAuthenticated {
		return nil, errors.New("not authenticated")
	}

	ids := make(map[string]string, len(localRecords))

	for _, record := range localRecords {
		if !record.HasServerID() {

			// TODO: get serverID and save record
			serverID := fmt.Sprintf("server_%s", record.ID)

		
			ids[record.ID] = serverID
		}
	}

	return  ids,nil
}

// GetServerURL returns the server URL.
func (h *ServerHandlers) GetServerURL() string {
	return h.serverURL
}

// SetServerURL sets the server URL.
func (h *ServerHandlers) SetServerURL(url string) {
	h.serverURL = url
}

func (h *ServerHandlers) LocalRecordMovedToServer(record *models.Record)(string, error){
	return  "", errors.New("server record moved to local not implemented")
}