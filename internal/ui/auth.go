package ui

import (
	"github.com/rivo/tview"

	"github.com/DanilNaum/secret-storage-client/internal/constants"
)

func (a *App) showAuthChoice() {
	modal := tview.NewModal().
		SetText("Welcome to Secret Storage Client!\n\nChoose action:").
		AddButtons([]string{"Login", "Register", "Exit"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			switch buttonLabel {
			case "Login":
				a.showLoginForm()
			case "Register":
				a.showRegistrationForm()
			case "Exit":
				a.app.Stop()
			}
		})

	a.pages.AddAndSwitchToPage("authChoice", modal, true)
}

func (a *App) showLoginForm() {
	form := tview.NewForm()

	usernameField := tview.NewInputField().SetLabel("Username").SetFieldWidth(20)
	passwordField := tview.NewInputField().SetLabel("Password").SetFieldWidth(20).SetMaskCharacter('*')

	form.AddFormItem(usernameField).
		AddFormItem(passwordField).
		AddButton("Login", func() {
			username := usernameField.GetText()
			password := passwordField.GetText()

			if username == "" || password == "" {
				form.SetTitle("Error: Fill all fields")
				return
			}

			err := a.serverHandlers.Authenticate(username, password)
			if err != nil {
				form.SetTitle("Error: " + err.Error())
				return
			}

			a.showMasterPasswordForm()
		}).
		AddButton("Back", func() {
			a.pages.RemovePage("login")
			a.showAuthChoice()
		})

	form.SetBorder(true).SetTitle("Login").SetTitleAlign(tview.AlignLeft)
	a.pages.AddAndSwitchToPage("login", form, true)
}

func (a *App) showRegistrationForm() {
	form := tview.NewForm()

	usernameField := tview.NewInputField().SetLabel("Username").SetFieldWidth(20)
	passwordField := tview.NewInputField().SetLabel("Password").SetFieldWidth(20).SetMaskCharacter('*')
	confirmPasswordField := tview.NewInputField().SetLabel("Confirm Password").SetFieldWidth(20).SetMaskCharacter('*')

	form.AddFormItem(usernameField).
		AddFormItem(passwordField).
		AddFormItem(confirmPasswordField).
		AddButton("Register", func() {
			username := usernameField.GetText()
			password := passwordField.GetText()
			confirmPassword := confirmPasswordField.GetText()

			if username == "" || password == "" || confirmPassword == "" {
				form.SetTitle("Error: Fill all fields")
				return
			}

			if password != confirmPassword {
				form.SetTitle("Error: Passwords do not match")
				return
			}

			if len(password) < 6 {
				form.SetTitle("Error: Password must be at least 6 characters")
				return
			}

			salt, err := a.serverHandlers.Register(username, password)
			if err != nil {
				form.SetTitle("Error: " + err.Error())
				return
			}

			a.showRegistrationSuccess(salt)
		}).
		AddButton("Back", func() {
			a.pages.RemovePage("register")
			a.showAuthChoice()
		})

	form.SetBorder(true).SetTitle("Registration").SetTitleAlign(tview.AlignLeft)
	a.pages.AddAndSwitchToPage("register", form, true)
}

func (a *App) showRegistrationSuccess(salt string) {
	modal := tview.NewModal().
		SetText("Registration successful!\n\nSalt received from server: " + salt + "\n\nPress OK to continue.").
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.RemovePage("registerSuccess")
			a.showMasterPasswordForm()
		})

	a.pages.AddAndSwitchToPage("registerSuccess", modal, true)
}

func (a *App) showMasterPasswordForm() {
	form := tview.NewForm()

	masterPasswordField := tview.NewInputField().SetLabel("Master Password").SetFieldWidth(20).SetMaskCharacter('*')
	confirmMasterPasswordField := tview.NewInputField().SetLabel("Confirm Master Password").SetFieldWidth(20).SetMaskCharacter('*')

	form.AddFormItem(masterPasswordField).
		AddFormItem(confirmMasterPasswordField).
		AddButton("Continue", func() {
			masterPassword := masterPasswordField.GetText()
			confirmMasterPassword := confirmMasterPasswordField.GetText()

			if masterPassword == "" || confirmMasterPassword == "" {
				form.SetTitle("Error: Fill all fields")
				return
			}

			if masterPassword != confirmMasterPassword {
				form.SetTitle("Error: Passwords do not match")
				return
			}

			if len(masterPassword) < 8 {
				form.SetTitle("Error: Master password must be at least 8 characters")
				return
			}

			a.serverHandlers.SetMasterPassword(masterPassword)

			a.pages.RemovePage("login")
			a.pages.RemovePage("register")
			a.pages.RemovePage("authChoice")
			a.pages.RemovePage("registerSuccess")
			a.pages.RemovePage("masterPassword")

			a.createMainUI()
		}).
		AddButton("Back", func() {
			a.pages.RemovePage("masterPassword")
			a.showAuthChoice()
		})

	form.SetBorder(true).SetTitle("Master Password Setup").SetTitleAlign(tview.AlignLeft)
	a.pages.AddAndSwitchToPage("masterPassword", form, true)
}

func (a *App) logout() {
	a.serverHandlers.Logout()

	a.currentRecord = nil
	a.currentIsServer = false
	a.currentRecordID = ""
	a.passwordVisible = false

	a.pages.RemovePage(constants.MainPageName)
	a.showAuthChoice()
}