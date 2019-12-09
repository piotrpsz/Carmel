/*
 * BSD 2-Clause License
 *
 *	Copyright (c) 2019, Piotr Pszczółkowski
 *	All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 * 1. Redistributions of source code must retain the above copyright notice, this
 * list of conditions and the following disclaimer.
 *
 * 2. Redistributions in binary form must reproduce the above copyright notice,
 * this list of conditions and the following disclaimer in the documentation
 * and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
 * FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
 * DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
 * SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
 * CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
 * OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
 * OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

package chat

import (
	"Carmel/chat/news"
	"Carmel/connector/message"
	"Carmel/connector/session"
	"Carmel/shared"
	"Carmel/shared/tr"
	"Carmel/shared/vtc"
	"context"
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/pango"
	"strings"
	"sync"
)

const (
	MyNameTag            = "my_name"
	MyMessageTag         = "my_message"
	OtherNameTag         = "other_name"
	OtherMessageTag      = "other_message"
	subtitleFormat       = "IP: %s"
	canConnectFormat     = "Would you like to chat with %s?"
	rsaKeyNotFoundFormat = "RSA public key for %s not found"
)

var (
	ownNameTagData = map[string]interface{}{
		"foreground": "#FFFAAA",
		"style":      pango.STYLE_ITALIC,
		"weight":     pango.WEIGHT_BOLD,
		"font":       "Italic 10", //"Sans Italic 12"
	}
	ownMessageData = map[string]interface{}{
		"foreground": "#999999",
	}

	otherNameTagData = map[string]interface{}{
		"foreground":    "#AAAFFF",
		"style":         pango.STYLE_ITALIC,
		"weight":        pango.WEIGHT_BOLD,
		"font":          "Italic 10", //"Sans Italic 12"
		"justification": gtk.JUSTIFY_RIGHT,
	}
	otherMessageData = map[string]interface{}{
		"foreground":    "#FFFFFF",
		"justification": gtk.JUSTIFY_RIGHT,
	}
)

type Window struct {
	app             *gtk.Application
	win             *gtk.ApplicationWindow
	buddyName       string
	ssn             *session.Session
	browser         *gtk.TextView
	browserBuffer   *gtk.TextBuffer
	entry           *gtk.TextView
	entryBuffer     *gtk.TextBuffer
	myNameTag       *gtk.TextTag
	myMessageTag    *gtk.TextTag
	otherNameTag    *gtk.TextTag
	otherMessageTag *gtk.TextTag
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	buddyNewsChan   chan news.News
}

func New(app *gtk.Application, role vtc.RoleType, buddyName string, ssn *session.Session) *Window {
	if finalInit(app, role, buddyName, ssn) {
		if win, err := gtk.ApplicationWindowNew(app); tr.IsOK(err) {
			w := &Window{app: app, win: win, buddyName: buddyName, ssn: ssn}
			if headerBar := w.createHeaderBar(); headerBar != nil {
				if menuButton := w.createMenu(); menuButton != nil {
					headerBar.PackEnd(menuButton)
					if content := w.createContent(); content != nil {
						win.Add(content)
						win.SetTitlebar(headerBar)
						win.SetDefaultSize(400, 400)

						w.ctx, w.cancel = context.WithCancel(context.Background())
						return w
					}
				}
			}
		}
	}
	return nil
}

func finalInit(app *gtk.Application, role vtc.RoleType, buddyName string, ssn *session.Session) bool {
	// Wszystko do tej pory poszło dobrze, ale może się okazać że
	// nie mamy publicznego klucza RSA dla wskazanej osoby.
	// Jeśli tak by było to dupa.
	if ssn.In.Enigma.SetBuddyRSAPublicKey(buddyName) {
		// Możemy kontynuuować komunikację, ale czy na pewno chcemy?
		if dialogCanConnectWith(app, buddyName) {
			switch role {
			case vtc.Server:
				if !sendAcceptance(ssn) {
					return false
				}
				if !ssn.SendKeys() {
					return false
				}
				if !ssn.ExchangeBlockIdentifiersAsServer() {
					return false
				}
			case vtc.Client:
				if !ssn.ReadKeys() {
					return false
				}
				if !ssn.ExchangeBlockIdentifiersAsClient() {
					return false
				}
			}
			return true
		}
	}
	return false
}

func sendAcceptance(ssn *session.Session) bool {
	msg := message.NewWithType(vtc.Answer)
	msg.Id = vtc.Login
	msg.Status = vtc.Accepted
	msg.Tstamp = shared.Now()
	if data := msg.ToJsonSnapped(); data != nil {
		if cipher := ssn.In.Enigma.EncryptRSA(data); cipher != nil {
			return ssn.In.Requester.SendRawMessage(cipher)
		}
	}
	return false
}

func dialogCanConnectWith(app *gtk.Application, buddyName string) bool {
	if dialog := gtk.MessageDialogNew(app.GetActiveWindow(), gtk.DIALOG_MODAL, gtk.MESSAGE_QUESTION, gtk.BUTTONS_YES_NO, ""); dialog != nil {
		defer dialog.Destroy()
		dialog.FormatSecondaryText(fmt.Sprintf(canConnectFormat, buddyName))
		if dialog.Run() == gtk.RESPONSE_YES {
			return true
		}
	}
	return false
}

func dialogUnknownRSAKey(app *gtk.Application, buddyName string) {
	if dialog := gtk.MessageDialogNew(app.GetActiveWindow(), gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_CLOSE, ""); dialog != nil {
		defer dialog.Destroy()
		dialog.FormatSecondaryText(fmt.Sprintf(rsaKeyNotFoundFormat, buddyName))
		dialog.Run()
	}
}

func (w *Window) ShowAll() {
	w.buddyNewsChan = make(chan news.News)
	w.wg.Add(2)

	go w.browserLoop(w.buddyNewsChan, &w.wg)
	go w.netLoop(w.buddyNewsChan, &w.wg)

	w.win.ShowAll()
	w.entry.GrabFocus()
}

func (w *Window) Close() {
	w.cancel()
	w.wg.Wait()
	close(w.buddyNewsChan)
	w.win.Close()
}

func (w *Window) createHeaderBar() *gtk.HeaderBar {
	if bar, err := gtk.HeaderBarNew(); tr.IsOK(err) {
		bar.SetShowCloseButton(false)
		bar.SetTitle(w.buddyName)
		address := strings.Split(w.ssn.In.RemoteAddr, ":")
		bar.SetSubtitle(fmt.Sprintf(subtitleFormat, address[0]))
		return bar
	}
	return nil
}

func (w *Window) createMenu() *gtk.MenuButton {
	if btn, err := gtk.MenuButtonNew(); tr.IsOK(err) {
		if menu := glib.MenuNew(); menu != nil {
			menu.Append("Quit", "win.close")

			closeAction := glib.SimpleActionNew("close", nil)
			closeAction.Connect("activate", w.Close)
			w.win.AddAction(closeAction)

			btn.SetMenuModel(&menu.MenuModel)
			return btn
		}
	}
	return nil
}

func (w *Window) createContent() *gtk.Box {
	if box, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0); tr.IsOK(err) {
		if separator, err := gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL); tr.IsOK(err) {
			if browserWindow := w.createBrowser(); browserWindow != nil {
				if entryWindow := w.createEntry(); entryWindow != nil {
					box.PackStart(browserWindow, true, true, 1)
					box.PackStart(separator, false, true, 1)
					box.PackStart(entryWindow, false, true, 1)
					return box
				}
			}
		}
	}
	return nil
}

/********************************************************************
*                                                                   *
*                           E N T R Y                               *
*                                                                   *
********************************************************************/

func (w *Window) createEntry() *gtk.ScrolledWindow {
	if scrolledWindow, err := gtk.ScrolledWindowNew(nil, nil); tr.IsOK(err) {
		if entry, err := gtk.TextViewNew(); tr.IsOK(err) {
			if buffer, err := entry.GetBuffer(); tr.IsOK(err) {
				entry.SetLeftMargin(5)
				entry.SetRightMargin(5)
				entry.SetWrapMode(gtk.WRAP_CHAR)
				entry.Connect("key-press-event", w.entryHandler)
				scrolledWindow.Add(entry)

				w.entry = entry
				w.entryBuffer = buffer
				return scrolledWindow
			}
		}
	}
	return nil
}

func (w *Window) entryHandler(_, e interface{}) {
	if event, ok := e.(*gdk.Event); ok {
		keyEvent := gdk.EventKeyNewFromEvent(event)
		if keyEvent.KeyVal() == gdk.KEY_Return {
			if (keyEvent.State() & uint(gdk.GDK_SHIFT_MASK)) == 0 {
				startIter := w.entryBuffer.GetStartIter()
				endIter := w.entryBuffer.GetEndIter()
				if text, err := w.entryBuffer.GetText(startIter, endIter, true); tr.IsOK(err) {

					w.entryBuffer.Delete(w.entryBuffer.GetStartIter(), w.entryBuffer.GetEndIter())
					w.entryBuffer.PlaceCursor(w.entryBuffer.GetIterAtLine(0))

					if msg := news.New(shared.MyUserName, text, true); msg.Valid() {
						if request := w.ssn.Out.Requester.Send(vtc.Message, []byte(text), nil); request != nil {
							if answer := w.ssn.Out.Responder.Read(request); answer != nil {
								w.buddyNewsChan <- msg
							}
						}
					}
				}
			}
		}
	}
}

/********************************************************************
*                                                                   *
*                         B R O W S E R                             *
*                                                                   *
********************************************************************/

func (w *Window) createBrowser() *gtk.ScrolledWindow {
	if scrolledWindow, err := gtk.ScrolledWindowNew(nil, nil); tr.IsOK(err) {
		if browser, err := gtk.TextViewNew(); tr.IsOK(err) {
			if buffer, err := browser.GetBuffer(); tr.IsOK(err) {
				browser.SetEditable(false)
				browser.SetCursorVisible(false)
				browser.SetLeftMargin(5)
				browser.SetRightMargin(5)
				scrolledWindow.Add(browser)

				w.myNameTag = buffer.CreateTag(MyNameTag, ownNameTagData)
				w.myMessageTag = buffer.CreateTag(MyMessageTag, ownMessageData)
				w.otherNameTag = buffer.CreateTag(OtherNameTag, otherNameTagData)
				w.otherMessageTag = buffer.CreateTag(OtherMessageTag, otherMessageData)

				w.browser = browser
				w.browserBuffer = buffer
				return scrolledWindow
			}
		}
	}
	return nil
}

func (w *Window) appendTextToBrowser(msg news.News) {
	nameTag := w.otherNameTag
	messageTag := w.otherMessageTag
	if msg.Own {
		nameTag = w.myNameTag
		messageTag = w.myMessageTag
	}

	w.browserBuffer.InsertWithTag(w.browserBuffer.GetEndIter(), msg.Name+"\n", nameTag)
	w.browserBuffer.InsertWithTag(w.browserBuffer.GetEndIter(), msg.Text+"\n\n", messageTag)

	mark := w.browserBuffer.GetMark("insert")
	w.browser.ScrollToMark(mark, 0.0, true, 0.0, 1.0)
}

func (w *Window) browserLoop(inChan <-chan news.News, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-w.ctx.Done():
			fmt.Println(w.ctx.Err())
			return
		case msg := <-inChan:
			w.appendTextToBrowser(msg)
		}
	}
}

func (w *Window) netLoop(inChan chan<- news.News, wg *sync.WaitGroup) {
	defer w.ssn.Close()

	for {
		select {
		case <-w.ctx.Done():
			fmt.Println(w.ctx.Err())
			return
		default:
			if request := w.ssn.In.Requester.Read(); request != nil {
				if request.Id == vtc.Message {
					if answer := w.ssn.In.Responder.Send(vtc.Ok, request, nil, nil); answer != nil {
						if msg := news.New(w.buddyName, string(request.Data), false); msg.Valid() {
							inChan <- msg
						}
					}
				}
			}
		}
	}
}
