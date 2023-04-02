package main

import (
	"context"
	"errors"
	"io"
	"log"
	"os"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Define other GUI components like logWindowWidget, logWriter, etc.

type logWriter struct {
	logWindow *logWindowWidget
}

type logWindowWidget struct {
	*widget.Entry
}

func newLogWindowWidget() *logWindowWidget {
	w := &logWindowWidget{
		Entry: widget.NewMultiLineEntry(),
	}
	w.Resize(fyne.NewSize(400, 300))
	return w
}

func (lw *logWindowWidget) AppendText(text string) {
	lw.Entry.SetText(lw.Entry.Text + text)
}

func (lw *logWindowWidget) CreateRenderer() fyne.WidgetRenderer {
	return lw.Entry.CreateRenderer()
}

func (lw *logWindowWidget) ApplyTheme() {
	// do nothing
}

func (lw *logWriter) Write(p []byte) (n int, err error) {
	lw.logWindow.AppendText(string(p))
	return len(p), nil
}

func showError(w fyne.Window, message string) {
	dialog.ShowError(errors.New(message), w)
}

func createGUI() (context.Context, context.CancelFunc) {
	guiApp := app.New()
	w := guiApp.NewWindow("Ableton Git Backup")

	// New "main" tab
	statusLabel := widget.NewLabel("Status: Not running")
	mainTab := container.NewVBox(
		widget.NewLabel("Ableton Git Sync"),
		statusLabel,
	)

	// Settings tab
	settings, err := loadSettings()
	projectPathEntry := widget.NewEntry()
	gitRepoURLEntry := widget.NewEntry()
	gitUsernameEntry := widget.NewEntry()
	gitPasswordEntry := widget.NewPasswordEntry()

	if err == nil {
		projectPathEntry.SetText(settings.ProjectPath)
		gitRepoURLEntry.SetText(settings.GitRepoURL)
		gitUsernameEntry.SetText(settings.GitUsername)
	}

	ctx, cancel := context.WithCancel(context.Background())

	startButton := widget.NewButton("Start", func() {
		projectPath := projectPathEntry.Text
		gitRepoURL := gitRepoURLEntry.Text
		gitUsername := gitUsernameEntry.Text
		gitPassword := gitPasswordEntry.Text

		// Save settings
		settings := &Settings{
			ProjectPath: projectPath,
			GitRepoURL:  gitRepoURL,
			GitUsername: gitUsername,
			GitPassword: gitPassword,
		}
		err := saveSettings(settings)
		if err != nil {
			log.Println("Failed to save settings:", err)
			return
		}

		go func() {
			err := settings.Validate()
			if err == nil {
				statusLabel.SetText("Status: Running")
				watchAbletonProject(ctx, settings.ProjectPath, settings.GitRepoURL, settings.GitUsername, settings.GitPassword, w)
			} else {
				showError(w, "Please complete all settings fields.")
			}
		}()

		go watchAbletonProject(ctx, projectPath, gitRepoURL, gitUsername, gitPassword, w)
	})

	stopButton := widget.NewButton("Stop", func() {
		cancel()
		statusLabel.SetText("Status: Not running")
	})

	pushButton := widget.NewButton("Push", func() {
		projectPath := projectPathEntry.Text
		gitUsername := gitUsernameEntry.Text
		gitPassword := gitPasswordEntry.Text

		r, err := git.PlainOpen(projectPath)
		if err != nil {
			log.Println("Error opening repository:", err)
			return
		}

		auth := &http.BasicAuth{
			Username: gitUsername,
			Password: gitPassword,
		}

		err = pushChanges(r, auth)

		if err != nil {
			log.Println("Failed to push changes:", err)
			return
		}

		log.Println("Pushed changes to remote repository")
	})

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Ableton Project Path", Widget: projectPathEntry},
			{Text: "Git Repository URL", Widget: gitRepoURLEntry},
			{Text: "Git Username", Widget: gitUsernameEntry},
			{Text: "Git Token", Widget: gitPasswordEntry},
		},
	}

	buttons := container.NewHBox(startButton, stopButton, pushButton)

	// Create the tab container
	// Create a log window
	logWindow := newLogWindowWidget()

	// Capture log output and display it in the log window
	log.SetOutput(io.MultiWriter(os.Stderr, &logWriter{logWindow}))

	tabs := container.NewAppTabs(
		container.NewTabItem("Main", mainTab),
		container.NewTabItem("Settings", form),
		container.NewTabItem("Log", logWindow),
	)
	tabs.SetTabLocation(container.TabLocationLeading)

	w.SetContent(container.NewVBox(tabs, buttons))
	w.Resize(fyne.NewSize(600, 200))
	w.ShowAndRun()

	return ctx, cancel
}
