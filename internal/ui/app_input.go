// Package ui provides keyboard input handling for the secret storage client.
// This file contains the input handling methods of the App struct that manage
// keyboard shortcuts and navigation throughout the application.
package ui

import (
	"github.com/DanilNaum/secret-storage-client/internal/constants"
	"github.com/gdamore/tcell/v2"
)

// setupInputHandling configures global keyboard input handling for the application.
// This method sets up keyboard shortcuts and navigation that work across different
// pages and contexts. It handles both global shortcuts and context-specific input.
//
// The input handling supports:
// - Function key shortcuts (F1-F6) for common operations
// - Tab navigation between lists and form fields
// - Context-aware input processing based on current page
func (a *App) setupInputHandling() {
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Get the name of the currently displayed page
		currentPageName, _ := a.pages.GetFrontPage()

		// Allow forms to handle their own input (especially Tab for field navigation)
		if currentPageName == constants.FormPageName || currentPageName == constants.TypeSelectPageName {
			// Pass the event through to the form for processing
			return event
		}

		// Handle global shortcuts only on the main page
		if currentPageName == constants.MainPageName {
			switch event.Key() {
			case tcell.KeyTab:
				// Switch focus between local and server lists
				a.switchActiveList()
				return nil

			case tcell.KeyF1:
				// Create new record
				a.showRecordForm(nil, true)
				return nil

			case tcell.KeyF2:
				// Synchronize all records
				a.syncRecords()
				return nil

			case tcell.KeyF3:
				// Edit current record (if any is selected)
				if a.currentRecord != nil {
					a.showRecordForm(a.currentRecord, false)
				}
				return nil

			case tcell.KeyF4:
				// Copy selected record between local and server
				a.copySelectedRecord()
				return nil

			case tcell.KeyF5:
				// Delete selected record
				a.deleteSelectedRecord()
				return nil

			case tcell.KeyF6:
				// Refresh server record (only available for server records)
				if a.currentIsServer {
					a.refreshServerRecord()
				}
				return nil
			}
		}

		// Pass through any unhandled events
		return event
	})
}
