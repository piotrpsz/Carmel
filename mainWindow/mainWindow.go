package mainWindow

import (
	"Carmel/rsakeys"
	"Carmel/shared"
	"Carmel/shared/tr"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"io/ioutil"
	"net/http"
)

const (
	licence = `BSD 2-Clause License

Copyright (c) 2019, Piotr Pszczółkowski
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this
list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
this list of conditions and the following disclaimer in the documentation
and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.`

	ipFormat = "<span font_desc='8' foreground='#999999'>IP: </span>" +
		"<span font_desc='10' foreground='#FFFFFF'> %s</span>"
	userFormat = "<span font_desc='8' foreground='#999999'>User: </span>" +
		"<span font_desc='10' foreground='#FFFFFF'> %s</span>"
	unknownUserFormat = "<span font_desc='8' foreground='#999999'>User: </span>" +
		"<span font_desc='10' foreground='#FF9966'> unknown</span>"
)

type MainWindow struct {
	app             *gtk.Application
	win             *gtk.ApplicationWindow
	user            *gtk.Label
	ipAddr          *gtk.Label
	connectToAction *glib.SimpleAction
	rsaAction       *glib.SimpleAction
}

func New(app *gtk.Application) *MainWindow {
	if win, err := gtk.ApplicationWindowNew(app); tr.IsOK(err) {
		w := &MainWindow{app: app, win: win}
		if headerBar := w.SetupHeaderBar(); headerBar != nil {
			win.SetTitlebar(headerBar)
			if w.SetupMenu(headerBar) {
				w.user, _ = gtk.LabelNew("")
				headerBar.PackStart(w.user)

				w.ipAddr, _ = gtk.LabelNew("")
				headerBar.PackEnd(w.ipAddr)

				go w.updateIP()
				go w.updateUser()

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
		return headerBar
	}
	return nil
}

func (mw *MainWindow) SetupMenu(headerBar *gtk.HeaderBar) bool {
	if menuButton, err := gtk.MenuButtonNew(); tr.IsOK(err) {
		if menu := glib.MenuNew(); menu != nil {

			menu.Append("Parameters for connection...", "custom.connection_parameters")
			menu.Append("Connect to...", "custom.connect_to")
			menu.Append("Generate RSA keys...", "custom.rsa_keys")
			menu.Append("Settings...", "custom.settings")
			menu.Append("About...", "custom.about")
			menu.Append("Quit", "app.quit")

			//=======================================================
			customGroup := glib.SimpleActionGroupNew()
			//.......................................................
			aboutAction := glib.SimpleActionNew("about", nil)
			aboutAction.Connect("activate", func() {
				mw.aboutActionHandler()
			})
			//.......................................................
			mw.connectToAction = glib.SimpleActionNew("connect_to", nil)
			mw.connectToAction.Connect("activate", func() {
				mw.connectToActionHandler()
			})
			//.......................................................
			mw.rsaAction = glib.SimpleActionNew("rsa_keys", nil)
			mw.rsaAction.Connect("activate", func() {
				mw.generatingRSAKeys()
			})
			//.......................................................
			connectionParametersAction := glib.SimpleActionNew("connection_parameters", nil)
			connectionParametersAction.Connect("activate", func() {
				fmt.Println("Clipboard")
				clipboard, _ := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD)
				clipboard.SetText("IDs: 34d8df, PIN: sjdh4")
			})
			//-------------------------------------------------------
			customGroup.AddAction(connectionParametersAction)
			customGroup.AddAction(aboutAction)
			customGroup.AddAction(mw.connectToAction)
			customGroup.AddAction(mw.rsaAction)

			mw.win.InsertActionGroup("custom", customGroup)
			//=======================================================

			menuButton.SetMenuModel(&menu.MenuModel)

			headerBar.PackEnd(menuButton)
			return true
		}
	}
	return false
}

func (mw *MainWindow) aboutActionHandler() {
	if dialog, err := gtk.AboutDialogNew(); tr.IsOK(err) {
		defer dialog.Destroy()

		dialog.SetTransientFor(mw.app.GetActiveWindow())
		dialog.SetProgramName(shared.AppName)
		dialog.SetVersion(shared.AppVersion)
		dialog.SetCopyright("Copyright (c) 2019, Beesoft Software")
		dialog.SetAuthors([]string{"Piotr Pszczółkowski (piotr@beesoft.pl)"})
		dialog.SetWebsite("http://www.beesoft.pl/carmel")
		dialog.SetWebsiteLabel("Carmel home page")
		dialog.SetLicense(licence)
		dialog.SetLogo(nil)
		dialog.Run()
	}
}

func (mw *MainWindow) connectToActionHandler() {
	msg :=
		"You are an undefined user.\n" +
			"You cannot currently connect to or receive calls from other Carmel users." +
			"This is due to the fact that no private key was found in the program directory.\n\n" +
			"First, generate your RSA keys (private and public)."

	if dialog := gtk.MessageDialogNew(mw.app.GetActiveWindow(), gtk.DIALOG_MODAL, gtk.MESSAGE_INFO, gtk.BUTTONS_OK, msg); dialog != nil {
		defer dialog.Destroy()

		dialog.SetTitle(fmt.Sprintf("%s - unknown user", shared.AppName))
		dialog.Run()
	}
}

func (mw *MainWindow) updateIP() {
	if response, err := http.Get("https://api.ipify.org/?format=json"); tr.IsOK(err) {
		defer response.Body.Close()
		if content, err := ioutil.ReadAll(response.Body); tr.IsOK(err) {
			data := make(map[string]interface{})
			if err := json.Unmarshal(content, &data); tr.IsOK(err) {
				if text, ok := data["ip"].(string); ok {
					shared.MyIPAddr = text
					markup := fmt.Sprintf(ipFormat, text)
					glib.IdleAdd(mw.ipAddr.SetMarkup, markup)
				}
			}
		}
	}
}

func (mw *MainWindow) updateUser() {
	if name := rsakeys.New().MyUserName(); name != "" {
		glib.IdleAdd(mw.user.SetMarkup, fmt.Sprintf(userFormat, name))
		return
	}
	glib.IdleAdd(mw.user.SetMarkup, unknownUserFormat)
	/*
	 */
}

func (mw *MainWindow) generatingRSAKeys() {
	fmt.Println("generateRSAKeys")
	manager := rsakeys.New()
	manager.CreateKeysForUser("piotr")

	privateKey := manager.PrivateKeyFromFileForUser("piotr")
	publicKey := manager.PublicKeyFromFileForUser("piotr")

	text := "Piotr, Artur, Błażej Pszczółkowscy"
	if cipher, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, []byte(text)); tr.IsOK(err) {
		if plain, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, cipher); tr.IsOK(err) {
			fmt.Println(string(plain))
		}
	}
}
