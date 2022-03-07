package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

var currentPath string = ""
var redoChars []string = []string{}

func save(w fyne.Window, entry *widget.Entry) error {
	txt := entry.Text
	txtbyte := []byte(txt)

	if len(currentPath) > 0 {
		newerr := ioutil.WriteFile(currentPath, txtbyte, 0644)
		if newerr != nil {
			log.Fatal(newerr)
			return newerr
		}
	} else {
		dialog.ShowFileSave(func(uc fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if uc != nil {
				newerr := ioutil.WriteFile(uc.URI().Path(), txtbyte, 0644)
				if newerr != nil {
					log.Fatal(newerr)
					return
				}
				currentPath = uc.URI().Path()
			}
		}, w)
	}

	return nil
}

func open(entry *widget.Entry, file fyne.URIReadCloser) error {
	if file != nil {
		currentPath = file.URI().Path()
		txt, err := ioutil.ReadAll(file)
		if err != nil {
			return err
		}
		entry.SetText(string(txt))
	}

	return nil
}

func undo(entry *widget.Entry) {
	if len(entry.Text) > 0 {
		var lastChar string = entry.Text[len(entry.Text)-1:]
		if len(redoChars) <= 5 {
			redoChars = append(redoChars, lastChar)
		} else {
			redoChars = redoChars[1:]
			redoChars = append(redoChars, lastChar)
		}
		entry.SetText(entry.Text[:len(entry.Text)-1])
	}
}

func redo(entry *widget.Entry) {
	if len(redoChars) > 0 {
		entry.SetText(entry.Text + redoChars[len(redoChars)-1])
		redoChars = redoChars[:len(redoChars)-1]
	}
}

func find(entry *widget.Entry, w fyne.Window) {
	dialog.ShowEntryDialog("Find", "Search for text", func(s string) {
		index := strings.Index(entry.Text, s)
		if index != -1 {
			dialog.ShowInformation("Found", fmt.Sprintf("Found %s at %d", s, index), w)
		}
	}, w)
}

func replace(entry *widget.Entry, w fyne.Window) {
	dialog.ShowEntryDialog("Replace", "Which word do you want to replace?", func(s string) {
		index := strings.Index(entry.Text, s)
		if index != -1 {
			dialog.ShowEntryDialog("Replace", "With what?", func(s2 string) {
				entry.SetText(strings.Replace(entry.Text, s, s2, 1))
			}, w)
		}
	}, w)
}

func new(entry *widget.Entry, w fyne.Window) {
	currentPath = ""
	redoChars = []string{}

	dialog.ShowFileSave(func(uc fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		if uc != nil {
			currentPath = uc.URI().Path()
		}
	}, w)
}

func main() {
	a := app.New()
	w := a.NewWindow("Text Editor")
	w.Resize(fyne.NewSize(800, 500))

	entry := widget.NewMultiLineEntry()
	entry.SetPlaceHolder("Type here")

	// Ctrl + S to save
	ctrlS := desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: desktop.ControlModifier}
	w.Canvas().AddShortcut(&ctrlS, func(shortcut fyne.Shortcut) {
		save(w, entry)
	})

	// Ctrl + N to new file
	ctrlN := desktop.CustomShortcut{KeyName: fyne.KeyN, Modifier: desktop.ControlModifier}
	w.Canvas().AddShortcut(&ctrlN, func(shortcut fyne.Shortcut) {
		entry.SetText("")
		new(entry, w)
	})

	// Ctrl + Z to undo
	ctrlZ := desktop.CustomShortcut{KeyName: fyne.KeyZ, Modifier: desktop.ControlModifier}
	w.Canvas().AddShortcut(&ctrlZ, func(shortcut fyne.Shortcut) {
		undo(entry)
	})

	// Ctrl + Y to redo
	ctrlY := desktop.CustomShortcut{KeyName: fyne.KeyY, Modifier: desktop.ControlModifier}
	w.Canvas().AddShortcut(&ctrlY, func(shortcut fyne.Shortcut) {
		redo(entry)
	})

	// Ctrl + F to find
	ctrlF := desktop.CustomShortcut{KeyName: fyne.KeyF, Modifier: desktop.ControlModifier}
	w.Canvas().AddShortcut(&ctrlF, func(shortcut fyne.Shortcut) {
		find(entry, w)
	})

	// Ctrl + H to replace
	ctrlH := desktop.CustomShortcut{KeyName: fyne.KeyH, Modifier: desktop.ControlModifier}
	w.Canvas().AddShortcut(&ctrlH, func(shortcut fyne.Shortcut) {
		replace(entry, w)
	})

	// Ctrl + O to open
	ctrlO := desktop.CustomShortcut{KeyName: fyne.KeyO, Modifier: desktop.ControlModifier}
	w.Canvas().AddShortcut(&ctrlO, func(shortcut fyne.Shortcut) {
		dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
			openErr := open(entry, file)
			if openErr != nil {
				dialog.ShowError(err, w)
			}
		}, w).Show()
	})

	// Ctrl + Q to quit
	ctrlQ := desktop.CustomShortcut{KeyName: fyne.KeyQ, Modifier: desktop.ControlModifier}
	w.Canvas().AddShortcut(&ctrlQ, func(shortcut fyne.Shortcut) {
		if len(currentPath) > 0 {
			dialog.ShowConfirm("Quit", "Are you sure want to quit?", func(b bool) {
				if b {
					a.Quit()
				} else {
					return
				}
			}, w)
		} else {
			a.Quit()
		}
	})

	openfile := fyne.NewMenuItem("Open File", func() {
		dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
			openErr := open(entry, file)
			if openErr != nil {
				dialog.ShowError(err, w)
			}
		}, w).Show()
	})

	savefile := fyne.NewMenuItem("Save File", func() {
		save(w, entry)
	})

	quit := fyne.NewMenuItem("Quit", func() {
		if len(currentPath) > 0 {
			dialog.ShowConfirm("Quit", "Are you sure want to quit?", func(b bool) {
				if b {
					a.Quit()
				} else {
					return
				}
			}, w)
		} else {
			a.Quit()
		}
	})

	newFile := fyne.NewMenuItem("New File", func() {
		entry.SetText("")
		new(entry, w)
	})

	redoEl := fyne.NewMenuItem("Redo", func() {
		redo(entry)
	})

	undoEl := fyne.NewMenuItem("Undo", func() {
		undo(entry)
	})

	findEl := fyne.NewMenuItem("Find", func() {
		find(entry, w)
	})

	replaceEl := fyne.NewMenuItem("Replace", func() {
		replace(entry, w)
	})

	seperator := fyne.NewMenuItemSeparator()

	w.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("File", openfile, savefile, newFile, seperator, quit),
		fyne.NewMenu("Edit", undoEl, redoEl, seperator, findEl, replaceEl),
	))

	w.SetContent(entry)
	w.ShowAndRun()
}
