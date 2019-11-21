package mainWindow

import (
	"Carmel/shared"
	"Carmel/shared/tr"
	"github.com/gotk3/gotk3/gtk"
)

type MainWindow struct {
	app *gtk.Application
	win *gtk.ApplicationWindow
}

func New(app *gtk.Application) *MainWindow{
	if win, err := gtk.ApplicationWindowNew(app); tr.IsOK(err) {
		w := &MainWindow{app: app, win:win}
		if w.Setup() {
			win.SetPosition(gtk.WIN_POS_CENTER)
			return w
		}
	}
	return nil
}

func (mw *MainWindow) ShowAll() {
	mw.win.ShowAll()
}

func (mw *MainWindow) Setup() bool {
	if headerBar, err := gtk.HeaderBarNew(); tr.IsOK(err) {
		headerBar.SetShowCloseButton(true)
		headerBar.SetTitle(shared.AppNameAndVersion())
		headerBar.SetSubtitle(shared.AppSubname)

		mw.win.SetTitlebar(headerBar)
		return true
	}
	return false
}
