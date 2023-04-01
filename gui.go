package main

import (
	"fmt"
	"os"
	"io"
	"log"
	"context"
	"errors"

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

	projectPathEntry := widget.NewEntry()
	gitRepoURLEntry := widget.NewEntry()
	gitUsernameEntry := widget.NewEntry()
	gitPasswordEntry := widget.NewPasswordEntry()

	ctx, cancel := context.WithCancel(context.Background())

	settings, err := loadSettings()
	if err == nil {
		projectPathEntry.SetText(settings.ProjectPath)
		gitRepoURLEntry.SetText(settings.GitRepoURL)
		gitUsernameEntry.SetText(settings.GitUsername)
	}


	startButton := widget.NewButton("Start", func() {
		projectPath := projectPathEntry.Text
		gitRepoURL := gitRepoURLEntry.Text
		gitUsername := gitUsernameEntry.Text
		gitPassword := gitPasswordEntry.Text

		// Save settings
		settings := &Settings{
			ProjectPath:  projectPath,
			GitRepoURL:   gitRepoURL,
			GitUsername:  gitUsername,
		}
		err := saveSettings(settings)
		if err != nil {
			log.Println("Failed to save settings:", err)
			return
		}

		// Clone the Git repository if it doesn't exist locally
		if _, err := os.Stat(projectPath); os.IsNotExist(err) {
			_, err := git.PlainClone(projectPath, false, &git.CloneOptions{
				URL: gitRepoURL,
				Auth: &http.BasicAuth{
					Username: gitUsername,
					Password: gitPassword,
				},
			})
			if err != nil {
				log.Println("Failed to clone repository:", err)
				showError(w, fmt.Sprintf("Failed to clone repository: %s", err.Error()))
				return
			}
		}

		go watchAbletonProject(ctx, projectPath, gitRepoURL, gitUsername, gitPassword, w)
	})

	stopButton := widget.NewButton("Stop", func() {
		cancel()
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
			{Text: "Git Password", Widget: gitPasswordEntry},
		},
	}

	buttons := container.NewHBox(startButton, stopButton, pushButton)

	// Create the tab container
	tabs := createTabContainer(form)
	
	w.SetContent(container.NewVBox(tabs, buttons))
	w.Resize(fyne.NewSize(600, 200))
	w.ShowAndRun()

	return ctx, cancel
}

func createTabContainer(form *widget.Form) *container.AppTabs {
	// Create a log window
	logWindow := newLogWindowWidget()

	// Capture log output and display it in the log window
	log.SetOutput(io.MultiWriter(os.Stderr, &logWriter{logWindow}))

	// Create the tab container with the form and the log window
	tabs := container.NewAppTabs(
		container.NewTabItem("Configuration", form),
		container.NewTabItem("Log", logWindow),
	)
	tabs.SetTabLocation(container.TabLocationLeading)

	return tabs
}