package localhandlers

import (
	"errors"
	"fmt"

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
func (h *LocalHandlers) LocalRecordMovedToServer(record *models.Record) error {
	serverID := fmt.Sprintf("server_%s", record.ID)
	h.writer.UpdateLocalRecordServerID(record.ID, serverID)
	return nil
}