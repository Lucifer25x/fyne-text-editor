package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

var currentPath string = ""
var undoText []string = []string{}
var redoText []string = []string{}

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

//TODO: Improve redo/undo
func undo(entry *widget.Entry) {
	if len(undoText) > 0 {
		if len(redoText) < 6 {
			redoText = append(redoText, entry.Text)
		} else {
			redoText = redoText[1:]
			redoText = append(redoText, entry.Text)
		}
		entry.SetText(undoText[len(undoText)-1])
		undoText = undoText[:len(undoText)-1]
	}
}

func redo(entry *widget.Entry) {
	if len(redoText) > 0 {
		undoText = append(undoText, entry.Text)
		entry.SetText(redoText[len(redoText)-1])
		redoText = redoText[:len(redoText)-1]
	}
}

func find(entry *widget.Entry, w fyne.Window) {
	dialog.ShowEntryDialog("Find", "Search for text", func(s string) {
		index := strings.Index(entry.Text, s)
		if index != -1 {
			row := strings.Count(entry.Text[:index], "\n")
			foundedLine := strings.Split(entry.Text[:index], "\n")[row]
			col := len(foundedLine)
			w.Canvas().Focus(entry)
			entry.CursorRow = row
			entry.CursorColumn = col
			entry.Refresh()
			dialog.ShowInformation("Found", fmt.Sprintf("Found '%s' at %d (row: %d, col: %d)", s, index, row, col), w)
		} else {
			dialog.ShowInformation("Not found", fmt.Sprintf("Could not find '%s'", s), w)
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
		} else {
			dialog.ShowInformation("Not found", fmt.Sprintf("Could not find '%s'", s), w)
		}
	}, w)
}

func new(entry *widget.Entry, w fyne.Window) {
	currentPath = ""
	redoText = []string{}

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

type Char struct {
	Char string
	Row  int
	Col  int
}

func main() {
	a := app.New()
	w := a.NewWindow("Text Editor")
	w.Resize(fyne.NewSize(800, 500))
	entry := widget.NewMultiLineEntry()
	entry.SetPlaceHolder("Type here")

	entry.OnChanged = func(s string) {
		if len(s) > 0 {
			lastChar := s[len(s)-1:]
			switch lastChar {
			case "{":
				entry.SetText(s + "}")
			case "(":
				entry.SetText(s + ")")
			case "[":
				entry.SetText(s + "]")
			}

			if len(undoText) < 6 {
				undoText = append(undoText, s)
			} else {
				undoText = append(undoText[1:], s)
			}
		}
	}

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
			w.SetTitle(path.Base(file.URI().Path()) + " - Text Editor")
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
			w.SetTitle(path.Base(file.URI().Path()) + " - Text Editor")
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

	pathEl := fyne.NewMenuItem("Path", func() {
		if len(currentPath) > 0 {
			dialog.ShowInformation("Path", currentPath, w)
		} else {
			dialog.ShowError(errors.New("No path found"), w)
		}
	})

	seperator := fyne.NewMenuItemSeparator()

	w.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("File", openfile, savefile, newFile, seperator, pathEl, quit),
		fyne.NewMenu("Edit", undoEl, redoEl, seperator, findEl, replaceEl),
	))

	w.SetContent(entry)
	w.ShowAndRun()
}
