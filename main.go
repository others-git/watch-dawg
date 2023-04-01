package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/fsnotify/fsnotify"
	"github.com/manifoldco/promptui"
)


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

func watchAbletonProject(projectPath, gitRepoURL, gitUsername, gitPassword string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if !strings.HasSuffix(event.Name, ".als") {
					continue
				}

				log.Println("Detected change:", event)
				r, err := git.PlainOpen(projectPath)
				if err != nil {
					log.Fatal(err)
				}

				w, err := r.Worktree()
				if err != nil {
					log.Fatal(err)
				}

				_, err = w.Add(".")
				if err != nil {
					log.Fatal(err)
				}

				commit, err := w.Commit("Ableton project file changed", &git.CommitOptions{
					Author: &object.Signature{
						Name:  gitUsername,
						Email: fmt.Sprintf("%s@example.com", gitUsername),
						When:  time.Now(),
					},
				})
				if err != nil {
					log.Fatal(err)
				}
				obj, err := r.CommitObject(commit)
				if err != nil {
					log.Fatal(err)
				}
				log.Println("Created new commit:", obj)

				// Periodically push changes
				go func() {
					ticker := time.NewTicker(24 * time.Hour)
					for {
						<-ticker.C
						err := r.Push(&git.PushOptions{
							RemoteName: "origin",
							RefSpecs: []plumbing.RefSpec{
								"refs/heads/*:refs/heads/*",
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
				}()

			case err := <-watcher.Errors:
				log.Println("Error:", err)
			}
		}
	}()

func main() {
	projectPath, gitRepoURL, gitUsername, gitPassword := promptForConfig()

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
			log.Fatal("Failed to clone repository:", err)
		}
	}

	watchAbletonProject(projectPath, gitRepoURL, gitUsername, gitPassword)
}
