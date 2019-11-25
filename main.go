package main

import (
	"Carmel/mainWindow"
	"Carmel/shared/tr"
	"fmt"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"os"
)

const (
	appID = "pl.beesoft.gtk3.carmel"
)


func main() {
	if app, err := gtk.ApplicationNew(appID, glib.APPLICATION_FLAGS_NONE); tr.IsOK(err) {
		app.Connect("activate", func() {
			if mw := mainWindow.New(app); mw != nil {
				newWindow := glib.SimpleActionNew("new", nil)
				newWindow.Connect("activate", func() {
					fmt.Println("new chatter window")
				})
				app.AddAction(newWindow)

				quitAction := glib.SimpleActionNew("quit", nil)
				quitAction.Connect("activate", func() {
					app.Quit()
				})
				app.AddAction(quitAction)
				mw.ShowAll()
			}
		})
		retv := app.Run(os.Args)
		os.Exit(retv)
	}
	os.Exit(1)
}
