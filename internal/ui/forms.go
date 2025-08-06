// Package ui provides form handling functionality for record creation and editing.
// This file contains the form-related methods of the App struct that handle
// user input for creating and modifying records of different types.
package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/DanilNaum/secret-storage-client/internal/constants"
	"github.com/DanilNaum/secret-storage-client/internal/models"
)

// showRecordForm displays the record creation or editing form.
// This method handles both creating new records and editing existing ones.
// For new records, it first shows a type selection dialog.
// For existing records, it directly shows the appropriate form.
//
// Parameters:
//   - record: the record to edit (nil for new records)
//   - isNew: true if creating a new record, false if editing existing
func (a *App) showRecordForm(record *models.Record, isNew bool) {
	form := tview.NewForm()
	modal := tview.NewModal()

	if isNew {
		// For new records, first show type selection
		a.showTypeSelectionModal(modal, form)
		return
	}

	// For existing records, show the form directly
	a.createRecordForm(form, record, false)
	a.pages.AddAndSwitchToPage(constants.FormPageName, form, true)
}

// showTypeSelectionModal displays a modal dialog for selecting record type.
// This method is called when creating new records to let users choose
// between different record types (Credentials, Text, File).
//
// Parameters:
//   - modal: the modal dialog to configure
//   - form: the form that will be shown after type selection
func (a *App) showTypeSelectionModal(modal *tview.Modal, form *tview.Form) {
	modal.SetText(constants.SelectRecordTypeMessage).
		AddButtons([]string{constants.CredentialsType, constants.TextType, constants.FileType, constants.CancelButton}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == constants.CancelButton {
				a.pages.SwitchToPage(constants.MainPageName)
				return
			}

			// Determine record type based on user selection
			var recordType models.RecordType
			switch buttonLabel {
			case constants.CredentialsType:
				recordType = models.Credentials
			case constants.TextType:
				recordType = models.TextData
			case constants.FileType:
				recordType = models.File
			}

			// Create and show the form for the selected type
			a.createRecordForm(form, &models.Record{Type: recordType}, true)
			a.pages.AddAndSwitchToPage(constants.FormPageName, form, true)
		})

	a.pages.AddAndSwitchToPage(constants.TypeSelectPageName, modal, true)
}

// createRecordForm creates and configures a form based on the record type.
// This method builds a dynamic form with fields appropriate for the record type.
// It handles keyboard navigation and provides save/cancel functionality.
//
// Parameters:
//   - form: the form widget to configure
//   - record: the record to edit (contains type and existing data)
//   - isNew: true if creating a new record, false if editing existing
func (a *App) createRecordForm(form *tview.Form, record *models.Record, isNew bool) {
	form.Clear(true)

	// Configure form to handle Tab key for field navigation
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Allow Tab for switching between form fields
		return event
	})

	// Add common name field for all record types
	form.AddInputField(constants.NameLabel, record.Name, constants.InputFieldWidth, nil, func(text string) {
		record.Name = text
	})

	// Add type-specific fields based on record type
	switch record.Type {
	case models.Credentials:
		// Add username and password fields for credentials
		form.AddInputField(constants.UsernameLabel, record.Username, constants.InputFieldWidth, nil, func(text string) {
			record.Username = text
		})
		form.AddPasswordField(constants.PasswordLabel, record.Password, constants.InputFieldWidth, constants.PasswordMask, func(text string) {
			record.Password = text
		})

	case models.TextData:
		// Add text area for text content
		form.AddTextArea(constants.ContentLabel, record.TextContent, constants.InputFieldWidth, constants.TextAreaHeight, 0, func(text string) {
			record.TextContent = text
		})

	case models.File:
		// Add file path field for file records
		form.AddInputField(constants.FilePathLabel, record.FilePath, constants.InputFieldWidth, nil, func(text string) {
			record.FilePath = text
		})
	}

	// Add action buttons
	form.AddButton(constants.SaveButton, func() {
		a.saveRecord(record, isNew)
	})

	form.AddButton(constants.CancelButton, func() {
		a.pages.SwitchToPage(constants.MainPageName)
	})
}

// saveRecord saves the record using the appropriate handler.
// This method handles both creating new records and updating existing ones.
// It determines the correct operation based on the isNew parameter and
// the record's IsServer flag.
//
// Parameters:
//   - record: the record to save
//   - isNew: true if creating a new record, false if updating existing
func (a *App) saveRecord(record *models.Record, isNew bool) {
	var err error

	if isNew {
		// Create new record through handler
		err = a.handlers.CreateRecord(record)
	} else {
		// Update existing record through appropriate handler
		if record.IsServer {
			err = a.handlers.ServerRecordChange(record)
		} else {
			err = a.handlers.LocalRecordChange(record)
		}
	}

	if err != nil {
		a.showError(err.Error())
		return
	}

	// Return to main page and refresh lists
	a.pages.SwitchToPage(constants.MainPageName)
	a.updateLocalList()
	a.updateServerList()
}
