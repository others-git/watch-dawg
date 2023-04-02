package main

import (
	"log"
	"fmt"
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
