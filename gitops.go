package main

import (
	"fmt"
	"log"
	"os"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func openRepository(projectPath string) (*git.Repository, error) {
	r, err := git.PlainOpen(projectPath)
	return r, err
}

func addToWorktreeAndCommit(r *git.Repository, gitUsername string) error {
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	_, err = w.Add(".")
	if err != nil {
		return err
	}

	commit, err := w.Commit("Ableton project file changed", &git.CommitOptions{
		Author: &object.Signature{
			Name:  gitUsername,
			Email: fmt.Sprintf("%s@example.com", gitUsername),
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}

	obj, err := r.CommitObject(commit)
	if err != nil {
		return err
	}

	log.Println("Created new commit:", obj)
	return nil
}

func pushChanges(r *git.Repository, auth *http.BasicAuth) error {
	log.Println(auth)
	err := r.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs: []config.RefSpec{
			config.RefSpec("refs/heads/*:refs/heads/*"),
		},
		Auth: auth,
	})

	if err != nil {
		return err
	}
	return nil
}

// checkoutRepo checks if the repository exists locally, and clones it if it doesn't.
func checkoutRepo(projectPath, gitRepoURL, gitUsername, gitPassword string) error {
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		_, err := git.PlainClone(projectPath, false, &git.CloneOptions{
			URL: gitRepoURL,
			Auth: &http.BasicAuth{
				Username: gitUsername,
				Password: gitPassword,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to clone repository: %w", err)
		}
	}
	return nil
}
