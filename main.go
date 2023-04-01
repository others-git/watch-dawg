package main

import (
	"fmt"
	"log"
	"strings"
	"time"
	"context"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/fsnotify/fsnotify"

	"fyne.io/fyne/v2"
)

func watchAbletonProject(ctx context.Context, projectPath, gitRepoURL, gitUsername, gitPassword string, w fyne.Window) {

    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Println("Error creating watcher:", err)
        return
    }
    defer watcher.Close()

    done := make(chan bool)

    auth := &http.BasicAuth{
        Username: gitUsername,
        Password: gitPassword,
    }

    go func() {
        for {
            select {
            case <-ctx.Done():
                return
            case event := <-watcher.Events:
				if !strings.HasSuffix(event.Name, ".als") || strings.HasPrefix(event.Name, ".git") || event.Op != fsnotify.Write {
                    continue
                }

                log.Println("Detected change:", event)
                r, err := openRepository(projectPath)
                if err != nil {
                    log.Println("Error opening repository:", err)
                    return
                }

                err = addToWorktreeAndCommit(r, gitUsername)
                if err != nil {
                    log.Println("Error committing changes:", err)
                }

                // Periodically push changes
                go func() {
                    ticker := time.NewTicker(1 * time.Hour)
                    for {
                        select {
                        case <-ctx.Done():
                            ticker.Stop()
                            return
                        case <-ticker.C:
                            err := pushChanges(r, auth)
                            if err != nil {
                                log.Println("Failed to push changes:", err)
                            }
                        }
                    }
                }()
            case err := <-watcher.Errors:
                log.Println("Error:", err)
                showError(w, fmt.Sprintf("Error watching project directory: %s", err.Error()))
            }
        }
    }()

    err = watcher.Add(projectPath)
    if err != nil {
        log.Println(err)
    }
    <-done
}

func main() {
	ctx, cancel := createGUI()
	defer cancel()
	<-ctx.Done()
}