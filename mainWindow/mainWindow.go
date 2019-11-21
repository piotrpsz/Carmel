package mainWindow

import (
	"Carmel/shared"
	"Carmel/shared/tr"
	"encoding/json"
	"fmt"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"io/ioutil"
	"net/http"
)

const (
	ipFormat = "<span font_desc='8' foreground='#999999'>IP: </span>" +
		"<span font_desc='10' foreground='#AAA555'> %s</span>"
)

type MainWindow struct {
	app *gtk.Application
	win *gtk.ApplicationWindow
	ip  *gtk.Label
}

func New(app *gtk.Application) *MainWindow {
	if win, err := gtk.ApplicationWindowNew(app); tr.IsOK(err) {
		w := &MainWindow{app: app, win: win}
		if headerBar := w.SetupHeaderBar(); headerBar != nil {
			win.SetTitlebar(headerBar)
			if w.SetupMenu(headerBar) {
				win.SetPosition(gtk.WIN_POS_CENTER)
				win.SetDefaultSize(400, 200)
				return w
			}
		}
	}
	return nil
}

func (mw *MainWindow) ShowAll() {
	mw.win.ShowAll()
}

func (mw *MainWindow) SetupHeaderBar() *gtk.HeaderBar {
	if headerBar, err := gtk.HeaderBarNew(); tr.IsOK(err) {
		headerBar.SetShowCloseButton(false)
		headerBar.SetTitle(shared.AppNameAndVersion())
		headerBar.SetSubtitle(shared.AppSubname)

		mw.ip, _ = gtk.LabelNew("")
		headerBar.PackStart(mw.ip)
		go mw.updateIP()
		return headerBar
	}
	return nil
}

func (mw *MainWindow) SetupMenu(headerBar *gtk.HeaderBar) bool {
	if menuButton, err := gtk.MenuButtonNew(); tr.IsOK(err) {
		if menu := glib.MenuNew(); menu != nil {
			menu.Append("About...", "custom.about")
			menu.Append("Connect to ...", "custom.connect_to")
			menu.Append("Generate RSA keys ...", "custom.rsa_keys")
			menu.Append("Quit", "app.quit")

			//=======================================================
			customGroup := glib.SimpleActionGroupNew()
			//.......................................................
			aboutAction := glib.SimpleActionNew("about", nil)
			aboutAction.Connect("activate", func() {
				fmt.Println("About ...")
			})
			//.......................................................
			connectToAction := glib.SimpleActionNew("connect_to", nil)
			connectToAction.Connect("activate", func() {
				fmt.Println("Connect to ...")
			})
			rsaAction := glib.SimpleActionNew("rsa_keys", nil)
			rsaAction.Connect("activate", func() {
				fmt.Println("Generate RSA keys ...")
			})
			//-------------------------------------------------------
			customGroup.AddAction(aboutAction)
			customGroup.AddAction(connectToAction)
			customGroup.AddAction(rsaAction)
			mw.win.InsertActionGroup("custom", customGroup)
			//=======================================================

			menuButton.SetMenuModel(&menu.MenuModel)

			headerBar.PackEnd(menuButton)
			return true
		}
	}
	return false
}

func (mw *MainWindow) updateIP() {
	if response, err := http.Get("https://api.ipify.org/?format=json"); tr.IsOK(err) {
		defer response.Body.Close()
		if content, err := ioutil.ReadAll(response.Body); tr.IsOK(err) {
			data := make(map[string]interface{})
			if err := json.Unmarshal(content, &data); tr.IsOK(err) {
				if text, ok := data["ip"].(string); ok {
					markup := fmt.Sprintf(ipFormat, text)
					glib.IdleAdd(mw.ip.SetMarkup, markup)
				}
			}
		}
	}
}
