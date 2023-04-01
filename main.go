package main

import (
	"fmt"
	"log"
	"os"
	"io"
	"strings"
	"time"
	"context"
	"encoding/json"
	"io/ioutil"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/config"
	"github.com/fsnotify/fsnotify"
	"github.com/manifoldco/promptui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const settingsFile = "settings.json"

type logWriter struct {
	logWindow *logWindowWidget
}

type Settings struct {
	ProjectPath  string `json:"project_path"`
	GitRepoURL   string `json:"git_repo_url"`
	GitUsername  string `json:"git_username"`
	GitPassword  string `json:"git_password"`
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


////////////////////////// GOOD STUFF //////////////////////////

func loadSettings() (*Settings, error) {
	data, err := ioutil.ReadFile(settingsFile)
	if err != nil {
		return nil, err
	}

	var settings Settings
	err = json.Unmarshal(data, &settings)
	if err != nil {
		return nil, err
	}

	return &settings, nil
}

func saveSettings(settings *Settings) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(settingsFile, data, 0644)
}


func promptForConfig() (string, string, string, string) {
	promptProjectPath := promptui.Prompt{
		Label: "Ableton Project Path",
		Validate: func(input string) error {
			_, err := os.Stat(input)
			if err != nil {
				return fmt.Errorf("Invalid path")
			}
			return nil
		},
	}
	projectPath, _ := promptProjectPath.Run()

	promptGitRepoURL := promptui.Prompt{
		Label: "Git Repository URL",
	}
	gitRepoURL, _ := promptGitRepoURL.Run()

	promptGitUsername := promptui.Prompt{
		Label: "Git Username",
	}
	gitUsername, _ := promptGitUsername.Run()

	promptGitPassword := promptui.Prompt{
		Label: "Git Password",
		Mask:  '*',
	}
	gitPassword, _ := promptGitPassword.Run()

	return projectPath, gitRepoURL, gitUsername, gitPassword
}

func watchAbletonProject(ctx context.Context, projectPath, gitRepoURL, gitUsername, gitPassword string) {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println("Error creating watcher:", err)
		return
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-watcher.Events:
				if !strings.HasSuffix(event.Name, ".als") || strings.HasPrefix(event.Name, ".git") {
					continue
				}

				log.Println("Detected change:", event)
				r, err := git.PlainOpen(projectPath)
				if err != nil {
					log.Println("Error opening repository:", err)
					return
				}

				w, err := r.Worktree()
				if err != nil {
					log.Println(err)
				}

				_, err = w.Add(".")
				if err != nil {
					log.Println(err)
				}

				commit, err := w.Commit("Ableton project file changed", &git.CommitOptions{
					Author: &object.Signature{
						Name:  gitUsername,
						Email: fmt.Sprintf("%s@example.com", gitUsername),
						When:  time.Now(),
					},
				})
				if err != nil {
					log.Println(err)
				}
				obj, err := r.CommitObject(commit)
				if err != nil {
					log.Println(err)
				}
				log.Println("Created new commit:", obj)

				// Periodically push changes
				go func() {
					ticker := time.NewTicker(1 * time.Hour)
					for {
						select {
						case <-ctx.Done():
							ticker.Stop()
							return
						case <-ticker.C:
							err := r.Push(&git.PushOptions{
								RemoteName: "origin",
								RefSpecs: []config.RefSpec{
									config.RefSpec("refs/heads/*:refs/heads/*"),
								},
								Auth: &http.BasicAuth{
									Username: gitUsername,
									Password: gitPassword,
								},
							})
							if err != nil {
								log.Println("Failed to push changes:", err)
							} else {
								log.Println("Pushed changes to remote repository")
							}
						}
					}
				}()

			case err := <-watcher.Errors:
				log.Println("Error:", err)
			}
		}
	}()

	err = watcher.Add(projectPath)
	if err != nil {
		log.Println(err)
	}
	<-done
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
		gitPasswordEntry.SetText(settings.GitPassword)
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
			GitPassword:  gitPassword,
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
				return
			}
		}

		go watchAbletonProject(ctx, projectPath, gitRepoURL, gitUsername, gitPassword)
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

		err = r.Push(&git.PushOptions{
			RemoteName: "origin",
			RefSpecs: []config.RefSpec{
				config.RefSpec("refs/heads/*:refs/heads/*"),
			},
			Auth: &http.BasicAuth{
				Username: gitUsername,
				Password: gitPassword,
			},
		})
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

func main() {
	ctx, cancel := createGUI()
	defer cancel()
	<-ctx.Done()
}