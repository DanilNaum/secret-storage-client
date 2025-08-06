// Package ui provides the user interface layer for the secret storage client.
// It implements a terminal-based UI using the tview library and manages
// the presentation of records, user interactions, and navigation between different views.
// This package follows the clean architecture principles by depending only on interfaces.
package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/rivo/tview"

	"github.com/DanilNaum/secret-storage-client/internal/constants"
	"github.com/DanilNaum/secret-storage-client/internal/models"
)

// StorageReader defines the interface for reading storage data used by UI.
// This interface provides read-only access to stored records,
// following the principle of least privilege for the UI layer.
type StorageReader interface {
	// GetLocalRecords returns all local records from storage.
	GetLocalRecords() []*models.Record
	// GetServerRecords returns all server record metadata from storage.
	GetServerRecords() []*models.ServerRecord
}

// EventHandlers defines the interface for event handlers used by UI.
// This interface provides all the operations that the UI can trigger,
// abstracting away the business logic implementation details.
type EventHandlers interface {
	// Local record handlers

	// CreateRecord creates a new local record.
	CreateRecord(record *models.Record) error
	// DeleteLocalRecord deletes an existing local record.
	DeleteLocalRecord(record *models.Record) error
	// LocalRecordChange updates an existing local record.
	LocalRecordChange(record *models.Record) error
	// LocalRecordOpen loads a local record by ID.
	LocalRecordOpen(id string) (*models.Record, error)
	// LocalRecordMovedToServer uploads a local record to server.
	LocalRecordMovedToServer(record *models.Record) error

	// Server record handlers

	// DeleteServerRecord deletes a server record.
	DeleteServerRecord(record *models.Record) error
	// ServerRecordChange updates a server record.
	ServerRecordChange(record *models.Record) error
	// ServerRecordOpen loads a server record by ID.
	ServerRecordOpen(id string) (*models.Record, error)
	// ServerRecordMovedToLocal downloads a server record to local storage.
	ServerRecordMovedToLocal(record *models.Record) error
	// RefreshServerRecord forces refresh of a server record from server.
	RefreshServerRecord(id string) (*models.Record, error)

	// System handlers

	// Sync performs full synchronization between local and server storage.
	Sync() error

	// Cache operations

	// GetCachedRecord retrieves cache information for a record.
	GetCachedRecord(id string) (*models.Record, bool, time.Time, bool)
	// SetCachedRecord stores a record in cache.
	SetCachedRecord(id string, record *models.Record)
	// RemoveCachedRecord removes a record from cache.
	RemoveCachedRecord(id string)
	// IsCacheExpired checks if a cached record has expired.
	IsCacheExpired(id string) bool
}

// App represents the main application UI controller.
// It manages the overall application state, UI components, and user interactions.
// The App follows the MVC pattern where it acts as both the view and controller.
type App struct {
	// Core application components

	// app is the main tview application instance.
	app *tview.Application
	// pages manages different UI pages (main, forms, dialogs).
	pages *tview.Pages
	// flex is the main layout container.
	flex *tview.Flex

	// Dependencies (injected via constructor)

	// storage provides read access to stored data.
	storage StorageReader
	// handlers provides business logic operations.
	handlers EventHandlers

	// UI components

	// localList displays the list of local records.
	localList *tview.List
	// serverList displays the list of server records.
	serverList *tview.List
	// detailsPanel shows detailed information about selected records.
	detailsPanel *tview.TextView
	// detailsContainer holds the details panel and action buttons.
	detailsContainer *tview.Flex

	// Application state

	// activeList tracks which list currently has focus.
	activeList *tview.List
	// currentRecord holds the currently selected record.
	currentRecord *models.Record
	// currentIsServer indicates if the current record is from server.
	currentIsServer bool
	// currentRecordID stores the current record ID for server records.
	currentRecordID string
}

// NewApp creates a new application instance with the provided dependencies.
// This constructor follows the dependency injection pattern, accepting
// interfaces rather than concrete implementations for better testability and flexibility.
//
// Parameters:
//   - storage: provides read access to stored records
//   - handlers: provides business logic operations
//
// Returns a configured App instance ready to run.
func NewApp(storage StorageReader, handlers EventHandlers) *App {
	return &App{
		app:      tview.NewApplication(),
		pages:    tview.NewPages(),
		storage:  storage,
		handlers: handlers,
	}
}

// SetHandlers updates the event handlers for the application.
// This method allows for runtime replacement of handlers,
// useful for testing or dynamic behavior changes.
func (a *App) SetHandlers(h EventHandlers) {
	a.handlers = h
}

// Run starts the application and enters the main event loop.
// This method initializes the UI, sets up event handling,
// and blocks until the application is terminated.
//
// Returns an error if the application fails to start or run.
func (a *App) Run() error {
	a.createMainUI()
	return a.app.SetRoot(a.pages, true).EnableMouse(true).Run()
}

// createMainUI creates and configures the main user interface.
// This method sets up the layout with three main panels:
// local records, server records, and details panel.
// It also configures the initial focus and navigation.
func (a *App) createMainUI() {
	a.flex = tview.NewFlex()

	// Create local records panel
	localPanel := a.createLocalPanel()

	// Create server records panel
	serverPanel := a.createServerPanel()

	// Create details panel
	a.createDetailsPanel()

	// Add panels to main flex with proportional sizing
	a.flex.AddItem(localPanel, 0, constants.ListColumnProportion, true)
	a.flex.AddItem(serverPanel, 0, constants.ListColumnProportion, false)
	a.flex.AddItem(a.detailsContainer, 0, constants.DetailColumnProportion, false)

	// Set local list as active by default
	a.activeList = a.localList
	a.updateListTitles()

	a.pages.AddPage(constants.MainPageName, a.flex, true, true)

	// Set up input handling for keyboard shortcuts
	a.setupInputHandling()

	// Set focus on active list
	a.app.SetFocus(a.activeList)
}

// createLocalPanel creates the local records panel with list and buttons.
// This panel displays all local records and provides actions for creating new records.
// Returns a configured flex container with the local records list and action buttons.
func (a *App) createLocalPanel() *tview.Flex {
	a.localList = tview.NewList().
		SetSelectedFunc(func(index int, name string, secondName string, shortcut rune) {
			a.clearOtherSelection(a.localList)
			a.showLocalRecordDetails(index)
		}).
		SetWrapAround(false).
		SetHighlightFullLine(true)

	a.localList.SetBorder(true).SetTitle(constants.LocalRecordsTitle)
	a.updateLocalList()

	// Create buttons for local record actions
	localButtons := tview.NewFlex().SetDirection(tview.FlexColumn)
	newRecordBtn := tview.NewButton(constants.NewRecordButton).SetSelectedFunc(func() {
		a.showRecordForm(nil, true)
	})
	localButtons.AddItem(newRecordBtn, 0, 1, false)

	// Create panel with list and buttons
	localPanel := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.localList, 0, 1, true).
		AddItem(localButtons, constants.ButtonRowHeight, 1, false)

	return localPanel
}

// createServerPanel creates the server records panel with list and buttons.
// This panel displays all server record metadata and provides actions for synchronization.
// Returns a configured flex container with the server records list and action buttons.
func (a *App) createServerPanel() *tview.Flex {
	a.serverList = tview.NewList().
		SetSelectedFunc(func(index int, name string, secondName string, shortcut rune) {
			a.clearOtherSelection(a.serverList)
			a.showServerRecordDetails(index)
		}).
		SetWrapAround(false).
		SetHighlightFullLine(true)

	a.serverList.SetBorder(true).SetTitle(constants.ServerRecordsTitle)
	a.updateServerList()

	// Create buttons for server record actions
	serverButtons := tview.NewFlex().SetDirection(tview.FlexColumn)
	syncBtn := tview.NewButton(constants.SyncButton).SetSelectedFunc(func() {
		a.syncRecords()
	})
	serverButtons.AddItem(syncBtn, 0, 1, false)

	// Create panel with list and buttons
	serverPanel := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.serverList, 0, 1, true).
		AddItem(serverButtons, constants.ButtonRowHeight, 1, false)

	return serverPanel
}

// createDetailsPanel creates the details panel for displaying record information.
// This panel shows detailed information about the selected record
// and provides action buttons for record operations.
func (a *App) createDetailsPanel() {
	a.detailsPanel = tview.NewTextView()
	a.detailsPanel.SetDynamicColors(true).
		SetScrollable(true).
		SetBorder(true).
		SetTitle(constants.DetailsTitle)

	a.detailsPanel.SetText(constants.SelectRecordMessage)

	a.detailsContainer = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.detailsPanel, 0, 1, false)
}

// updateLocalList refreshes the local records list display.
// This method updates the list with current local records and shows
// sync status indicators for records that have been uploaded to server.
func (a *App) updateLocalList() {
	a.localList.Clear()
	for _, record := range a.storage.GetLocalRecords() {
		displayName := record.Name
		if record.HasServerID() {
			displayName += " [green](synced)[-]"
		}
		a.localList.AddItem(displayName, fmt.Sprintf(constants.TypeFieldLabel, record.Type.String()), 0, nil)
	}

	title := fmt.Sprintf(constants.RecordCountFormat, constants.LocalRecordsTitle, len(a.storage.GetLocalRecords()), constants.ScrollHint)
	a.localList.SetTitle(title)
}

// updateServerList refreshes the server records list display.
// This method updates the list with current server record metadata and shows
// cache status indicators (cached, expired, or metadata-only).
func (a *App) updateServerList() {
	a.serverList.Clear()
	for _, serverRecord := range a.storage.GetServerRecords() {
		// Check cache status through handlers
		cachedRecord, found, _, expired := a.handlers.GetCachedRecord(serverRecord.ID)

		var displayName, displayType string
		if found && !expired {
			// Show cached record with indicator
			displayName = cachedRecord.Name + " " + constants.CachedIndicator
			displayType = fmt.Sprintf(constants.TypeFieldLabel, cachedRecord.Type.String())
		} else if found && expired {
			// Show expired cached record
			displayName = cachedRecord.Name + " " + constants.ExpiredIndicator
			displayType = fmt.Sprintf(constants.TypeFieldLabel, cachedRecord.Type.String())
		} else {
			// Show server record metadata (name and type from server)
			displayName = serverRecord.Name
			displayType = fmt.Sprintf(constants.TypeFieldLabel, serverRecord.Type.String())
		}

		a.serverList.AddItem(displayName, displayType, 0, nil)
	}

	title := fmt.Sprintf(constants.RecordCountFormat, constants.ServerRecordsTitle, len(a.storage.GetServerRecords()), constants.ScrollHint)
	a.serverList.SetTitle(title)
}

// updateListTitles updates list titles to show which list is currently active.
// This method provides visual feedback about which list has focus
// by highlighting the active list title.
func (a *App) updateListTitles() {
	localTitle := fmt.Sprintf(constants.RecordCountFormat, constants.LocalRecordsTitle, len(a.storage.GetLocalRecords()), constants.ScrollHint)
	serverTitle := fmt.Sprintf(constants.RecordCountFormat, constants.ServerRecordsTitle, len(a.storage.GetServerRecords()), constants.ScrollHint)

	if a.activeList == a.localList {
		localTitle = fmt.Sprintf(constants.ActiveTitleFormat, localTitle, constants.ActiveIndicator)
	} else {
		serverTitle = fmt.Sprintf(constants.ActiveTitleFormat, serverTitle, constants.ActiveIndicator)
	}

	a.localList.SetTitle(localTitle)
	a.serverList.SetTitle(serverTitle)
}

// showLocalRecordDetails displays details of a selected local record.
// This method loads and displays the full details of a local record
// in the details panel with appropriate action buttons.
func (a *App) showLocalRecordDetails(index int) {
	records := a.storage.GetLocalRecords()
	if index < 0 || index >= len(records) {
		return
	}

	record := records[index]
	a.currentRecord = record
	a.currentIsServer = false
	a.currentRecordID = record.ID

	a.displayRecordDetails(record, false, time.Time{}, false)
}

// showServerRecordDetails displays details of a selected server record.
// This method handles loading server record content from cache or server,
// and displays the details with appropriate cache status information.
func (a *App) showServerRecordDetails(index int) {
	serverRecords := a.storage.GetServerRecords()
	if index < 0 || index >= len(serverRecords) {
		return
	}

	serverRecord := serverRecords[index]
	a.currentRecordID = serverRecord.ID

	// Check cache status through handlers
	cachedRecord, found, cachedAt, expired := a.handlers.GetCachedRecord(serverRecord.ID)

	if found {
		// Show cached record (even if expired)
		a.currentRecord = cachedRecord
		a.currentIsServer = true
		a.displayRecordDetails(cachedRecord, true, cachedAt, expired)
		return
	}

	// Try to load from server through handler
	record, err := a.handlers.ServerRecordOpen(serverRecord.ID)
	if err != nil {
		a.showError(err.Error())
		return
	}

	if record != nil {
		record.IsServer = true
		record.ServerID = serverRecord.ID
		// Cache the record through handlers
		a.handlers.SetCachedRecord(serverRecord.ID, record)
		a.currentRecord = record
		a.currentIsServer = true
		a.displayRecordDetails(record, true, time.Now(), false)
		// Update the server list to show cached status
		a.updateServerList()
	}
}

// displayRecordDetails displays record details in the details panel.
// This method formats and displays record information with cache status
// and provides appropriate action buttons based on record type and source.
//
// Parameters:
//   - record: the record to display
//   - isServer: whether this is a server record
//   - cachedAt: when the record was cached (for server records)
//   - expired: whether the cached record has expired
func (a *App) displayRecordDetails(record *models.Record, isServer bool, cachedAt time.Time, expired bool) {
	var content strings.Builder

	// Add cache information for server records
	if isServer && !cachedAt.IsZero() {
		if expired {
			content.WriteString("[red]CACHE EXPIRED[-]\n")
		} else {
			content.WriteString("[blue]CACHED[-]\n")
		}
		content.WriteString(fmt.Sprintf(constants.CacheTimeFormat, cachedAt.Format("15:04:05")))
		content.WriteString("\n")
	}

	content.WriteString(fmt.Sprintf(constants.DetailNameLabel, record.Name))
	content.WriteString(fmt.Sprintf(constants.DetailTypeLabel, record.Type.String()))

	// Show sync status for local records
	if !isServer && record.HasServerID() {
		content.WriteString(fmt.Sprintf("[green]Server ID:[-] %s\n", record.ServerID))
	}

	// Display type-specific content
	switch record.Type {
	case models.Credentials:
		content.WriteString(fmt.Sprintf(constants.DetailUsernameLabel, record.Username))
		content.WriteString(fmt.Sprintf(constants.DetailPasswordLabel, strings.Repeat("*", len(record.Password))))
	case models.TextData:
		content.WriteString(fmt.Sprintf(constants.DetailContentLabel, record.TextContent))
	case models.File:
		content.WriteString(fmt.Sprintf(constants.DetailFileLabel, record.FilePath))
	}

	a.detailsPanel.SetText(content.String())

	// Create action buttons based on record type and source
	buttons := tview.NewFlex().SetDirection(tview.FlexColumn)

	editBtn := tview.NewButton(constants.EditButton).SetSelectedFunc(func() {
		a.showRecordForm(record, false)
	})
	buttons.AddItem(editBtn, 0, 1, false)

	var copyBtn *tview.Button
	if isServer {
		copyBtn = tview.NewButton(constants.CopyToLocalButton).SetSelectedFunc(func() {
			a.copySelectedRecord()
		})
	} else {
		copyBtn = tview.NewButton(constants.CopyToServerButton).SetSelectedFunc(func() {
			a.copySelectedRecord()
		})
	}
	buttons.AddItem(copyBtn, 0, 1, false)

	deleteBtn := tview.NewButton(constants.DeleteButton).SetSelectedFunc(func() {
		a.deleteSelectedRecord()
	})
	buttons.AddItem(deleteBtn, 0, 1, false)

	// Add refresh button for server records
	if isServer {
		refreshBtn := tview.NewButton(constants.RefreshButton).SetSelectedFunc(func() {
			a.refreshServerRecord()
		})
		buttons.AddItem(refreshBtn, 0, 1, false)
	}

	// Update details container with content and buttons
	a.detailsContainer.Clear()
	a.detailsContainer.AddItem(a.detailsPanel, 0, 1, false)
	a.detailsContainer.AddItem(buttons, constants.ButtonRowHeight, 1, false)
}

// refreshServerRecord refreshes the current server record from server.
// This method forces a reload of the server record, bypassing the cache,
// and updates the display with the fresh data.
func (a *App) refreshServerRecord() {
	if !a.currentIsServer || a.currentRecordID == "" {
		return
	}

	record, err := a.handlers.RefreshServerRecord(a.currentRecordID)
	if err != nil {
		a.showError(err.Error())
		return
	}

	if record != nil {
		record.IsServer = true
		record.ServerID = a.currentRecordID
		// Update cache through handlers
		a.handlers.SetCachedRecord(a.currentRecordID, record)
		a.currentRecord = record
		a.displayRecordDetails(record, true, time.Now(), false)
		// Update the server list to show refreshed status
		a.updateServerList()
	}
}

// clearDetails clears the details panel and resets the current record state.
// This method is called when no record is selected or when switching contexts.
func (a *App) clearDetails() {
	a.currentRecord = nil
	a.currentIsServer = false
	a.currentRecordID = ""

	a.detailsPanel.SetText(constants.SelectRecordMessage)

	a.detailsContainer.Clear()
	a.detailsContainer.AddItem(a.detailsPanel, 0, 1, false)

	a.app.SetFocus(a.activeList)
}

// switchActiveList switches focus between local and server lists.
// This method handles keyboard navigation between the two main lists
// and updates the visual indicators accordingly.
func (a *App) switchActiveList() {
	if a.activeList == a.localList {
		a.activeList = a.serverList
	} else {
		a.activeList = a.localList
	}
	a.updateListTitles()
	a.app.SetFocus(a.activeList)
}

// clearOtherSelection clears selection in other lists and sets the active list.
// This method ensures that only one list shows selection at a time
// and updates the focus and visual indicators.
func (a *App) clearOtherSelection(currentList *tview.List) {
	a.activeList = currentList
	a.updateListTitles()
	a.app.SetFocus(a.activeList)
}

// copySelectedRecord copies the selected record to the other storage location.
// This method handles copying records between local and server storage,
// triggering the appropriate business logic through handlers.
func (a *App) copySelectedRecord() {
	if a.currentRecord == nil {
		return
	}

	var err error

	if a.currentIsServer {
		// Copy from server to local
		err = a.handlers.ServerRecordMovedToLocal(a.currentRecord)
	} else {
		// Copy from local to server
		err = a.handlers.LocalRecordMovedToServer(a.currentRecord)
	}

	if err != nil {
		a.showError(err.Error())
		return
	}

	// Refresh lists to show the changes
	a.updateLocalList()
	a.updateServerList()
	a.updateListTitles()
}

// deleteSelectedRecord deletes the currently selected record.
// This method shows a confirmation dialog before performing the deletion
// to prevent accidental data loss.
func (a *App) deleteSelectedRecord() {
	if a.currentRecord == nil {
		return
	}

	// Show confirmation dialog
	modal := tview.NewModal().
		SetText(constants.DeleteConfirmText).
		AddButtons([]string{constants.YesButton, constants.NoButton}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == constants.YesButton {
				a.performDelete()
			}
			a.pages.SwitchToPage(constants.MainPageName)
		})

	a.pages.AddAndSwitchToPage("deleteConfirm", modal, true)
}

// performDelete performs the actual deletion of the selected record.
// This method is called after user confirmation and handles the deletion
// through the appropriate business logic handlers.
func (a *App) performDelete() {
	if a.currentRecord == nil {
		return
	}

	var err error

	if a.currentIsServer {
		// Delete server record (this will also remove from cache in handler)
		err = a.handlers.DeleteServerRecord(a.currentRecord)
	} else {
		// Delete local record
		err = a.handlers.DeleteLocalRecord(a.currentRecord)
	}

	if err != nil {
		a.showError(err.Error())
	} else {
		a.clearDetails()
		// Refresh lists to show the changes
		a.updateLocalList()
		a.updateServerList()
		a.updateListTitles()
	}
}

// syncRecords synchronizes all records between local and server storage.
// This method triggers a full synchronization operation through the handlers
// and updates the display to reflect the changes.
func (a *App) syncRecords() {
	err := a.handlers.Sync()
	if err != nil {
		a.showError(err.Error())
		return
	}

	// Refresh both lists to show synchronization results
	a.updateLocalList()
	a.updateServerList()
	a.updateListTitles()
}

// showError displays an error message to the user.
// This method creates a modal dialog to show error messages
// and provides a way for users to acknowledge and dismiss the error.
func (a *App) showError(message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{constants.OkButton}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.SwitchToPage(constants.MainPageName)
		})

	a.pages.AddAndSwitchToPage(constants.ErrorPageName, modal, true)
}
