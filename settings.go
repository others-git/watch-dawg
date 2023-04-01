package main

import (
	"encoding/json"
	"io/ioutil"
)

const settingsFile = "settings.json"

type Settings struct {
	ProjectPath  string `json:"project_path"`
	GitRepoURL   string `json:"git_repo_url"`
	GitUsername  string `json:"git_username"`
	GitPassword  string `json:"git_password"`
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
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(settingsFile, data, 0644)
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