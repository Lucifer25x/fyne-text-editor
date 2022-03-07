package main

import (
	"io/ioutil"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

var currentPath string = ""

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

func main() {
	a := app.New()
	w := a.NewWindow("Text Editor")
	w.Resize(fyne.NewSize(800, 500))

	entry := widget.NewMultiLineEntry()
	entry.SetPlaceHolder("Type here")

	ctrlS := desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: desktop.ControlModifier}
	w.Canvas().AddShortcut(&ctrlS, func(shortcut fyne.Shortcut) {
		save(w, entry)
	})

	ctrlO := desktop.CustomShortcut{KeyName: fyne.KeyO, Modifier: desktop.ControlModifier}
	w.Canvas().AddShortcut(&ctrlO, func(shortcut fyne.Shortcut) {
		dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
			openErr := open(entry, file)
			if openErr != nil {
				dialog.ShowError(err, w)
			}
		}, w).Show()
	})

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

	openfile := fyne.NewMenuItem("Open", func() {
		dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
			openErr := open(entry, file)
			if openErr != nil {
				dialog.ShowError(err, w)
			}
		}, w).Show()
	})

	savefile := fyne.NewMenuItem("Save", func() {
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

	w.SetMainMenu(fyne.NewMainMenu(fyne.NewMenu("File", openfile, savefile, quit)))

	w.SetContent(entry)
	w.ShowAndRun()
}
