package storage

import (
	"fmt"

	"github.com/DanilNaum/secret-storage-client/internal/models"
)

// Storage manages local and server records in memory.
type Storage struct {
	localRecords  []*models.Record
	serverRecords []*models.ServerRecord
}

// NewStorage creates a new storage instance with predefined test data.
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
func (s *Storage) GetLocalRecords() []*models.Record {
	return s.localRecords
}

// GetServerRecords returns all server record metadata.
func (s *Storage) GetServerRecords() []*models.ServerRecord {
	return s.serverRecords
}

// AddLocalRecord adds a new record to local storage.
func (s *Storage) AddLocalRecord(record *models.Record) {
	if record.ID == "" {
		record.ID = fmt.Sprintf("local_%d", len(s.localRecords)+1)
	}
	record.IsServer = false
	s.localRecords = append(s.localRecords, record)
}

// AddServerRecord adds server record metadata to storage.
func (s *Storage) AddServerRecord(serverRecord *models.ServerRecord) {
	s.serverRecords = append(s.serverRecords, serverRecord)
}

// RemoveLocalRecord removes a local record by index.
func (s *Storage) RemoveLocalRecord(index int) *models.Record {
	if index < 0 || index >= len(s.localRecords) {
		return nil
	}
	record := s.localRecords[index]
	s.localRecords = append(s.localRecords[:index], s.localRecords[index+1:]...)
	return record
}

// RemoveServerRecord removes a server record metadata by index.
func (s *Storage) RemoveServerRecord(index int) *models.ServerRecord {
	if index < 0 || index >= len(s.serverRecords) {
		return nil
	}
	record := s.serverRecords[index]
	s.serverRecords = append(s.serverRecords[:index], s.serverRecords[index+1:]...)
	return record
}

// CopyToServer creates a copy of a local record for server upload.
func (s *Storage) CopyToServer(index int) *models.Record {
	if index < 0 || index >= len(s.localRecords) {
		return nil
	}
	record := s.localRecords[index]
	return record.Clone()
}

// CopyToLocal adds a server record to local storage.
func (s *Storage) CopyToLocal(serverRecord *models.Record) {
	copy := serverRecord.Clone()
	copy.ID = fmt.Sprintf("local_%d", len(s.localRecords)+1)
	copy.IsServer = false
	s.localRecords = append(s.localRecords, copy)
}

// UpdateLocalRecordServerID updates the server ID of a local record.
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
func (s *Storage) FindLocalRecordByServerID(serverID string) *models.Record {
	for _, record := range s.localRecords {
		if record.ServerID == serverID {
			return record
		}
	}
	return nil
}

// FindServerRecordByID finds server record metadata by server ID.
func (s *Storage) FindServerRecordByID(serverID string) *models.ServerRecord {
	for _, record := range s.serverRecords {
		if record.ID == serverID {
			return record
		}
	}
	return nil
}

// SyncRecords replaces all records with the provided data.
func (s *Storage) SyncRecords(localRecords []*models.Record, serverRecords []*models.ServerRecord) {
	s.localRecords = localRecords
	s.serverRecords = serverRecords
}