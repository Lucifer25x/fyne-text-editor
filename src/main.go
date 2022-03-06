package main

import (
	"io/ioutil"
	"log"
	"os/user"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// Save File
func save(filename *widget.Entry, entry *widget.Entry) error {
	myself, err := user.Current()
	if err != nil {
		return err
	}
	homedir := myself.HomeDir
	desktop := homedir + "/Desktop"
	txt := entry.Text
	txtbyte := []byte(txt)
	file := filename.Text
	path := desktop + "/" + file

	newerr := ioutil.WriteFile(file, txtbyte, 0644)

	if newerr != nil {
		log.Fatal(newerr)
		return newerr
	}

	log.Println("File saved to: " + path)

	return nil
}

// Open File
func open(filename *widget.Entry, entry *widget.Entry, file fyne.URIReadCloser) error {
	if file != nil {
		txt, err := ioutil.ReadAll(file)
		filepath := file.URI()
		if err != nil {
			return err
		}
		entry.SetText(string(txt))
		filename.SetText(filepath.Path())
	}

	return nil
}

func main() {
	a := app.New()
	w := a.NewWindow("Hello World")
	w.Resize(fyne.NewSize(800, 500))

	filename := widget.NewEntry()
	entry := widget.NewMultiLineEntry()
	entry.SetPlaceHolder("Type here")
	saveButton := widget.NewButton("Save", func() {
		err := save(filename, entry)
		if err != nil {
			dialog.ShowError(err, w)
		}
	})

	// Ctrl + S shortcut
	ctrlS := desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: desktop.ControlModifier}
	w.Canvas().AddShortcut(&ctrlS, func(shortcut fyne.Shortcut) {
		err := save(filename, entry)
		if err != nil {
			dialog.ShowError(err, w)
		}
	})

	// Ctrl + O shortcut
	ctrlO := desktop.CustomShortcut{KeyName: fyne.KeyO, Modifier: desktop.ControlModifier}
	w.Canvas().AddShortcut(&ctrlO, func(shortcut fyne.Shortcut) {
		dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
			openErr := open(filename, entry, file)
			if openErr != nil {
				dialog.ShowError(err, w)
			}
		}, w).Show()
	})

	vbox := container.NewVBox(filename, entry, saveButton)

	// Menu Elements
	openfile := fyne.NewMenuItem("Open", func() {
		dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
			openErr := open(filename, entry, file)
			if openErr != nil {
				dialog.ShowError(err, w)
			}
		}, w).Show()
	})

	savefile := fyne.NewMenuItem("Save", func() {
		save(filename, entry)
	})

	w.SetMainMenu(fyne.NewMainMenu(fyne.NewMenu("File", openfile, savefile)))

	w.SetContent(vbox)
	w.ShowAndRun()
}
