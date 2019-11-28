package mainWindow

import (
	"Carmel/dialog/connectTo"
	"Carmel/dialog/dialogWithOneField"
	"Carmel/dialog/waitForConnection"
	"Carmel/rsakeys"
	"Carmel/shared"
	"Carmel/shared/tr"
	"encoding/json"
	"fmt"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"io/ioutil"
	"net/http"
	"strings"
	"unicode"
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

			menu.Append("Wait for connection...", "custom.wait4connection")
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
				mw.generatingRSAKeysHandler()
			})
			//.......................................................
			wait4connectionAction := glib.SimpleActionNew("wait4connection", nil)
			wait4connectionAction.Connect("activate", func() {
				mw.waitForConnection()
				//clipboard, _ := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD)
				//clipboard.SetText("IDs: 34d8df, PIN: sjdh4")
			})
			//-------------------------------------------------------
			customGroup.AddAction(wait4connectionAction)
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

func (mw *MainWindow) notDefinedUserNameInfo() {
	const msg = "You are an undefined user.\n"
	const msgSecondary = "Sorry, you can't currently connect to or receive calls from other Carmel users." +
		"This is due to the fact that no private key was found in the program directory.\n\n" +
		"First, generate your RSA keys (private and public)."

	if dialog := gtk.MessageDialogNew(mw.app.GetActiveWindow(), gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_CANCEL, msg); dialog != nil {
		defer dialog.Destroy()
		dialog.FormatSecondaryText(msgSecondary)
		dialog.Run()
	}
}

/********************************************************************
*                                                                   *
*      W A I T   F O R   C O N N E C T I O N   H A N D L E R        *
*                                                                   *
********************************************************************/

func (mw *MainWindow) waitForConnection() {
	if shared.MyUserName == "" {
		mw.notDefinedUserNameInfo()
		return
	}

	if dialog := waitForConnection.New(mw.app); dialog != nil {
		defer dialog.Destroy()

		dialog.ShowAll()
		dialog.Run()
	}
}

/********************************************************************
*                                                                   *
*            C O N N E C T   T O   H A N D L E R                    *
*                                                                   *
********************************************************************/

func (mw *MainWindow) connectToActionHandler() {
	if shared.MyUserName == "" {
		mw.notDefinedUserNameInfo()
		return
	}

	if dialog := connectTo.New(mw.app); dialog != nil {
		defer dialog.Destroy()

		dialog.ShowAll()
		dialog.Run()
	}

	//chatter := chat.New(mw.app)
	//chatter.ShowAll()
}

/********************************************************************
*                                                                   *
*     G E N E R A T I N G   R S A   K E Y S   H A N D L E R         *
*                                                                   *
********************************************************************/

func (mw *MainWindow) generatingRSAKeysHandler() {
	if userName, ok := mw.getNameFromDialog(); ok {
		mw.createRSAKeysForUserName(userName)
	}
}

// getNameFromDialog
// Displays dialog where user can enter the 'user name'.
// When all is OK returns string with name as first result (second result is true).
// When something was wrong the second parameter is false.
func (mw *MainWindow) getNameFromDialog() (string, bool) {
	validate := func(text string) dialogWithOneField.ValidationResult {
		retv := dialogWithOneField.Ok

		switch {
		case strings.ContainsRune(text, ' '):
			retv = dialogWithOneField.SpacesNotAllowed
		case len(text) == 0:
			retv = dialogWithOneField.EmptyStringNotAllowed
		case unicode.IsDigit([]rune(text)[0]):
			retv = dialogWithOneField.DigitAtStartNotAllowed
		}

		return retv
	}

	if dialog := dialogWithOneField.New(mw.app, validate); dialog != nil {
		defer dialog.Destroy()

		const (
			prompt      = "User name:"
			description = "The username is used in the name of the key\n" +
				"'pem' files (private and public).\n" +
				"After creating the keys, send the public key\n" +
				"to the person you want to talk."
		)

		dialog.SetPrompt(prompt)
		dialog.SetDescription(description)
		if shared.MyUserName != "" {
			dialog.SetValue(shared.MyUserName)
		}

		dialog.ShowAll()

		switch dialog.Run() {
		case gtk.RESPONSE_ACCEPT:
			return dialog.GetValue(), true
		case gtk.RESPONSE_CANCEL:
			// nothing to do
		}
	}
	return "", false
}

func (mw *MainWindow) createRSAKeysForUserName(userName string) bool {
	if rsaManager := rsakeys.New(); rsaManager != nil {
		canCreate := true
		if rsaManager.ExistPrivateKeyFor(userName) || rsaManager.ExistPublicKeyFor(userName) {
			canCreate = false
			if mw.canRecreateKeys(userName) {
				if rsaManager.RemoveKeysFor(userName) {
					canCreate = true
				}
			}
		}
		if canCreate {
			if rsaManager.CreateKeysForUser(userName) {
				shared.MyUserName = userName
				mw.updateUser()
				return true
			}
		}
	}
	return false
}

// canRecreateKeys
// Displays a dialog in which the user should decide
// whether to delete old keys and create new ones.
func (mw *MainWindow) canRecreateKeys(userName string) bool {
	const (
		msgFormat = "Keys for user %s already exists"
		msgSecond = "Would you like to recreate keys?"
	)

	title := fmt.Sprintf(msgFormat, userName)
	if dialog := gtk.MessageDialogNew(mw.app.GetActiveWindow(), gtk.DIALOG_MODAL, gtk.MESSAGE_WARNING, gtk.BUTTONS_OK, title); dialog != nil {
		defer dialog.Destroy()
		if _, err := dialog.AddButton("Cancel", gtk.RESPONSE_CANCEL); tr.IsOK(err) {
			dialog.FormatSecondaryText(msgSecond)
			return dialog.Run() == gtk.RESPONSE_OK
		}
	}
	return false
}

/********************************************************************
*                                                                   *
*                         U P D A T E R S                           *
*                                                                   *
********************************************************************/

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
}

/*

	privateKey := manager.PrivateKeyFromFileForUser("piotr")
	publicKey := manager.PublicKeyFromFileForUser("piotr")

	text := "Piotr, Artur, Błażej Pszczółkowscy"
	if cipher, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, []byte(text)); tr.IsOK(err) {
		if plain, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, cipher); tr.IsOK(err) {
			fmt.Println(string(plain))
		}
	}

*/
