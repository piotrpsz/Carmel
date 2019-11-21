package main

import (
	"Carmel/mainWindow"
	"Carmel/shared/tr"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"os"
)

const (
	appID = "pl.beesoft.gtk3.carmel"

)

var (
	App *gtk.Application
)

func main() {
	if app, err := gtk.ApplicationNew(appID, glib.APPLICATION_FLAGS_NONE); tr.IsOK(err) {
		app.Connect("activate", func() {
			if mw := mainWindow.New(app); mw != nil {
				mw.ShowAll()
			}
		})
		retv := app.Run(os.Args)
		os.Exit(retv)
	}
}

