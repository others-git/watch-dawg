package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func createGUI() (string, string, string, string) {
	guiApp := app.New()
	w := guiApp.NewWindow("Ableton Git Backup")

	projectPathEntry := widget.NewEntry()
	gitRepoURLEntry := widget.NewEntry()
	gitUsernameEntry := widget.NewEntry()
	gitPasswordEntry := widget.NewPasswordEntry()

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Ableton Project Path", Widget: projectPathEntry},
			{Text: "Git Repository URL", Widget: gitRepoURLEntry},
			{Text: "Git Username", Widget: gitUsernameEntry},
			{Text: "Git Password", Widget: gitPasswordEntry},
		},
		OnSubmit: func() {
			w.Close()
		},
	}

	w.SetContent(container.NewVBox(form))
	w.Resize(fyne.NewSize(400, 200))
	w.ShowAndRun()

	return projectPathEntry.Text, gitRepoURLEntry.Text, gitUsernameEntry.Text, gitPasswordEntry.Text
}