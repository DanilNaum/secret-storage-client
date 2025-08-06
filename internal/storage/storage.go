// Package storage provides data storage functionality for local and server records.
// It implements the storage interfaces defined by the handlers layer and manages
// the persistence and retrieval of record data.
package storage

import (
	"fmt"

	"github.com/DanilNaum/secret-storage-client/internal/models"
)

// Storage manages local and server records in memory.
// It implements the handlers.StorageAdapter interface, providing both
// read and write operations for local records and server record metadata.
// This implementation uses in-memory storage with test data for demonstration purposes.
type Storage struct {
	// localRecords holds all local records with full data.
	localRecords []*models.Record
	// serverRecords holds server record metadata (ID, name, type only).
	// Full server record content is loaded on-demand and cached separately.
	serverRecords []*models.ServerRecord
}

// NewStorage creates a new storage instance with predefined test data.
// The test data includes sample local records and server record metadata
// to demonstrate the application functionality.
// In a production environment, this would load data from persistent storage.
func NewStorage() *Storage {
	return &Storage{
		localRecords: []*models.Record{
			{ID: "local_1", ServerID: "", Type: models.Credentials, Name: "Рабочий аккаунт", Username: "user1", Password: "pass123", IsServer: false},
			{ID: "local_2", ServerID: "", Type: models.TextData, Name: "Заметка", TextContent: "Не забыть сделать задание", IsServer: false},
			{ID: "local_3", ServerID: "", Type: models.Credentials, Name: "Email аккаунт", Username: "user@email.com", Password: "email123", IsServer: false},
			{ID: "local_4", ServerID: "", Type: models.TextData, Name: "Список покупок", TextContent: "Молоко, хлеб, масло", IsServer: false},
			{ID: "local_5", ServerID: "", Type: models.File, Name: "Документ", FilePath: "/path/to/document.pdf", IsServer: false},
			{ID: "local_6", ServerID: "", Type: models.Credentials, Name: "Банковский аккаунт", Username: "bank_user", Password: "secure123", IsServer: false},
			{ID: "local_7", ServerID: "", Type: models.TextData, Name: "Идеи проекта", TextContent: "Новые функции для приложения", IsServer: false},
			{ID: "local_8", ServerID: "", Type: models.Credentials, Name: "Социальная сеть", Username: "social_user", Password: "social123", IsServer: false},
		},
		serverRecords: []*models.ServerRecord{
			{ID: "server_01", Name: "Корпоративный аккаунт", Type: models.Credentials},
			{ID: "server_02", Name: "Проектная документация", Type: models.TextData},
			{ID: "server_03", Name: "Сертификаты", Type: models.File},
		},
	}
}

// GetLocalRecords returns all local records.
// This method implements the StorageReader interface and provides
// read-only access to the complete list of local records.
func (s *Storage) GetLocalRecords() []*models.Record {
	return s.localRecords
}

// GetServerRecords returns all server record metadata.
// This method implements the StorageReader interface and provides
// access to server record metadata (ID, name, type) without full content.
// Full server record content is loaded separately on-demand.
func (s *Storage) GetServerRecords() []*models.ServerRecord {
	return s.serverRecords
}

// AddLocalRecord adds a new record to local storage.
// This method implements the StorageWriter interface.
// If the record doesn't have an ID, it generates a new local ID.
// The record is marked as not being from server (IsServer = false).
func (s *Storage) AddLocalRecord(record *models.Record) {
	if record.ID == "" {
		record.ID = fmt.Sprintf("local_%d", len(s.localRecords)+1)
	}
	record.IsServer = false
	s.localRecords = append(s.localRecords, record)
}

// AddServerRecord adds server record metadata to storage.
// This method implements the StorageWriter interface.
// It stores only the metadata (ID, name, type) of server records.
// Full server record content is managed separately through caching.
func (s *Storage) AddServerRecord(serverRecord *models.ServerRecord) {
	s.serverRecords = append(s.serverRecords, serverRecord)
}

// RemoveLocalRecord removes a local record by index.
// This method implements the StorageWriter interface.
// Returns the removed record if successful, nil if index is invalid.
// The caller is responsible for ensuring the index is valid.
func (s *Storage) RemoveLocalRecord(index int) *models.Record {
	if index < 0 || index >= len(s.localRecords) {
		return nil
	}
	record := s.localRecords[index]
	s.localRecords = append(s.localRecords[:index], s.localRecords[index+1:]...)
	return record
}

// RemoveServerRecord removes a server record metadata by index.
// This method implements the StorageWriter interface.
// Returns the removed server record metadata if successful, nil if index is invalid.
// This only removes the metadata; cached content is managed separately.
func (s *Storage) RemoveServerRecord(index int) *models.ServerRecord {
	if index < 0 || index >= len(s.serverRecords) {
		return nil
	}
	record := s.serverRecords[index]
	s.serverRecords = append(s.serverRecords[:index], s.serverRecords[index+1:]...)
	return record
}

// CopyToServer creates a copy of a local record for server upload.
// Returns a clone of the local record at the specified index.
// Returns nil if the index is invalid.
// The returned record can be modified without affecting the original.
func (s *Storage) CopyToServer(index int) *models.Record {
	if index < 0 || index >= len(s.localRecords) {
		return nil
	}
	record := s.localRecords[index]
	return record.Clone()
}

// CopyToLocal adds a server record to local storage.
// This method implements the StorageWriter interface.
// Creates a new local record from the server record data,
// assigns a new local ID, and marks it as not from server.
func (s *Storage) CopyToLocal(serverRecord *models.Record) {
	copy := serverRecord.Clone()
	copy.ID = fmt.Sprintf("local_%d", len(s.localRecords)+1)
	copy.IsServer = false
	s.localRecords = append(s.localRecords, copy)
}

// UpdateLocalRecordServerID updates the server ID of a local record.
// This method implements the StorageWriter interface.
// Used when a local record is synchronized with the server and receives a server ID.
// Returns true if the record was found and updated, false otherwise.
func (s *Storage) UpdateLocalRecordServerID(localID, serverID string) bool {
	for _, record := range s.localRecords {
		if record.ID == localID {
			record.ServerID = serverID
			return true
		}
	}
	return false
}

// FindLocalRecordByServerID finds a local record by its server ID.
// This method implements the StorageWriter interface.
// Used to locate local records that have been synchronized with the server.
// Returns the record if found, nil otherwise.
func (s *Storage) FindLocalRecordByServerID(serverID string) *models.Record {
	for _, record := range s.localRecords {
		if record.ServerID == serverID {
			return record
		}
	}
	return nil
}

// FindServerRecordByID finds server record metadata by server ID.
// This method implements the StorageWriter interface.
// Returns the server record metadata if found, nil otherwise.
// This only returns metadata; full content must be loaded separately.
func (s *Storage) FindServerRecordByID(serverID string) *models.ServerRecord {
	for _, record := range s.serverRecords {
		if record.ID == serverID {
			return record
		}
	}
	return nil
}

// SyncRecords replaces all records with the provided data.
// This method is used for bulk synchronization operations.
// It completely replaces both local records and server record metadata.
// Use with caution as it will overwrite all existing data.
func (s *Storage) SyncRecords(localRecords []*models.Record, serverRecords []*models.ServerRecord) {
	s.localRecords = localRecords
	s.serverRecords = serverRecords
}
