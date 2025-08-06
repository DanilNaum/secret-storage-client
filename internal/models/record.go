// Package models defines the core data structures used throughout the application.
// It provides models for records, record types, and server record metadata.
package models

// RecordType represents the type of a record in the storage system.
// It defines the different categories of data that can be stored.
type RecordType int

const (
	// Credentials represents a record containing username/password pairs.
	Credentials RecordType = iota
	// TextData represents a record containing plain text content.
	TextData
	// File represents a record containing file path information.
	File
)

// Record represents a single record in the storage system.
// It contains both local and server identification, along with the actual data.
// Records can exist locally only or be synchronized with a server.
type Record struct {
	// ID is the local identifier for the record, always present and unique within local storage.
	ID string
	// ServerID is the server identifier for the record, empty if not synchronized with server.
	ServerID string
	// Type indicates the category of data stored in this record.
	Type RecordType
	// Name is the display name of the record, used for identification in UI.
	Name string
	// Username stores the username for Credentials type records.
	Username string
	// Password stores the password for Credentials type records.
	Password string
	// TextContent stores the text data for TextData type records.
	TextContent string
	// FilePath stores the file path for File type records.
	FilePath string
	// IsServer indicates whether this record was loaded from server.
	// This affects how the record is displayed and processed in the UI.
	IsServer bool
}

// ServerRecord represents server record metadata without full content.
// This is used for efficient display of server records list without loading
// the complete record data until needed.
type ServerRecord struct {
	// ID is the server identifier for the record.
	ID string
	// Name is the display name of the record from server.
	Name string
	// Type indicates the category of data stored in this server record.
	Type RecordType
}

// String returns the string representation of RecordType.
// This method implements the Stringer interface for better debugging and display.
func (rt RecordType) String() string {
	switch rt {
	case Credentials:
		return "Credentials"
	case TextData:
		return "Text"
	case File:
		return "File"
	default:
		return "Unknown"
	}
}

// Clone creates a deep copy of the record.
// This is useful for creating independent copies when moving records
// between different storage locations or for backup purposes.
func (r *Record) Clone() *Record {
	return &Record{
		ID:          r.ID,
		ServerID:    r.ServerID,
		Type:        r.Type,
		Name:        r.Name,
		Username:    r.Username,
		Password:    r.Password,
		TextContent: r.TextContent,
		FilePath:    r.FilePath,
		IsServer:    r.IsServer,
	}
}

// HasServerID checks if the record is synchronized with server.
// Returns true if the record has been uploaded to server and has a server ID.
// This is used to determine sync status and display appropriate indicators in UI.
func (r *Record) HasServerID() bool {
	return r.ServerID != ""
}