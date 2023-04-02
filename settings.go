package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

const settingsFile = "settings.json"

type Settings struct {
	ProjectPath string `json:"project_path"`
	GitRepoURL  string `json:"git_repo_url"`
	GitUsername string `json:"git_username"`
	GitPassword string `json:"git_password"`
}

type SettingsWithoutPassword struct {
	ProjectPath string `json:"project_path"`
	GitRepoURL  string `json:"git_repo_url"`
	GitUsername string `json:"git_username"`
}

func (s *Settings) Validate() error {
	if s.ProjectPath == "" {
		return errors.New("project path is required")
	}
	if s.GitRepoURL == "" {
		return errors.New("Git repository URL is required")
	}
	if s.GitUsername == "" {
		return errors.New("Git username is required")
	}
	if s.GitPassword == "" {
		return errors.New("Git password is required")
	}
	return nil
}

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
	settingsWithoutPassword := &SettingsWithoutPassword{
		ProjectPath: settings.ProjectPath,
		GitRepoURL:  settings.GitRepoURL,
		GitUsername: settings.GitUsername,
	}

	data, err := json.MarshalIndent(settingsWithoutPassword, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(settingsFile, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

/*
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
*/
