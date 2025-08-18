package localhandlers

import (
	"errors"

	"github.com/DanilNaum/secret-storage-client/internal/models"
)

// StorageReader defines the interface for reading storage data.
type StorageReader interface {
	GetLocalRecords() []*models.Record
}

// StorageWriter defines the interface for writing storage data.
type StorageWriter interface {
	AddLocalRecord(record *models.Record)
	RemoveLocalRecord(index int) *models.Record
	FindLocalRecordByServerID(serverID string) *models.Record
	UpdateLocalRecordServerID(localID, serverID string) bool
}

// StorageAdapter combines reader and writer interfaces.
type StorageAdapter interface {
	StorageReader
	StorageWriter
}

// LocalHandlers defines all event handlers for local record operations.
type LocalHandlers struct {
	storage StorageReader
	writer  StorageWriter
}

// NewLocalHandlers creates local handlers with storage adapter.
func NewLocalHandlers(storage StorageAdapter) *LocalHandlers {
	return &LocalHandlers{
		storage: storage,
		writer:  storage,
	}
}

// CreateRecord creates a new local record.
func (h *LocalHandlers) CreateRecord(record *models.Record) error {
	
	h.writer.AddLocalRecord(record)
	return nil
}

// DeleteLocalRecord deletes a local record.
func (h *LocalHandlers) DeleteLocalRecord(record *models.Record) error {
	localRecords := h.storage.GetLocalRecords()
	for i, r := range localRecords {
		if r.ID == record.ID {
			h.writer.RemoveLocalRecord(i)
			return nil
		}
	}
	return errors.New("local record not found")
}

// LocalRecordChange updates an existing local record.
func (h *LocalHandlers) LocalRecordChange(record *models.Record) error {
	localRecords := h.storage.GetLocalRecords()
	for _, r := range localRecords {
		if r.ID == record.ID {
			*r = *record
			return nil
		}
	}
	return errors.New("local record not found")
}

// LocalRecordOpen loads a local record by ID.
func (h *LocalHandlers) LocalRecordOpen(id string) (*models.Record, error) {
	localRecords := h.storage.GetLocalRecords()
	for _, r := range localRecords {
		if r.ID == id {
			return r, nil
		}
	}
	return nil, errors.New("local record not found")
}

// LocalRecordMovedToServer uploads a local record to the server.
func (h *LocalHandlers) LocalRecordMovedToServer(record *models.Record, serverID string) error {
	h.writer.UpdateLocalRecordServerID(record.ID, serverID)
	return nil
}

// ServerRecordMovedToLocal downloads a server record to local storage.
func (h *LocalHandlers) ServerRecordMovedToLocal(record *models.Record) error {
	if existingRecord := h.writer.FindLocalRecordByServerID(record.ServerID); existingRecord != nil {
		existingRecord.Name = record.Name
		existingRecord.Type = record.Type
		existingRecord.Username = record.Username
		existingRecord.Password = record.Password
		existingRecord.TextContent = record.TextContent
		existingRecord.FilePath = record.FilePath
	} else {
		newRecord := record.Clone()
		newRecord.ServerID = record.ServerID
		record.ID = ""
		newRecord.IsServer = true
		h.writer.AddLocalRecord(newRecord)
	}

	return nil
}