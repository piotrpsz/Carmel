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
	MyNameTag        = "my_name"
	MyMessageTag     = "my_message"
	OtherNameTag     = "other_name"
	OtherMessageTag  = "other_message"
	subtitleFormat   = "IP: %s"
	canConnectFormat = "Would you like to chat with %s?"
	connectionClosed = "Connection with %s is closed"
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
	headerBar       *gtk.HeaderBar
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
	stopAction      *glib.SimpleAction
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	buddyNewsChan   chan news.News
	connectionInUse bool
	mutex           sync.Mutex
}

func New(app *gtk.Application, role vtc.RoleType, buddyName string, ssn *session.Session) *Window {
	if finalInit(app, role, buddyName, ssn) {
		if win, err := gtk.ApplicationWindowNew(app); tr.IsOK(err) {
			w := &Window{app: app, win: win, buddyName: buddyName, ssn: ssn, connectionInUse: true}
			if w.headerBar = w.createHeaderBar(); w.headerBar != nil {
				if menuButton := w.createMenu(); menuButton != nil {
					w.headerBar.PackEnd(menuButton)
					if content := w.createContent(); content != nil {
						win.Add(content)
						win.SetTitlebar(w.headerBar)
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

func (w *Window) ShowAll() {
	w.buddyNewsChan = make(chan news.News)
	w.wg.Add(1)

	go w.browserLoop(w.buddyNewsChan, &w.wg)
	go w.netLoop(w.buddyNewsChan, &w.wg)

	w.win.ShowAll()
	w.entry.GrabFocus()
}

// Akcja wywołana ponieważ użytkownik wybrał 'Quit' w menu okna,
func (w *Window) closeWindowAction() {
	if w.connectionInUse {
		w.stopConnectionAction()
	}
	w.win.Close()
}

// Akcja wywołana ponieważ użytkownik wybrał 'Stop' w menu okna.
func (w *Window) stopConnectionAction() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.connectionInUse {
		w.sendDisconnectRequest()
		w.dialogConnectionClosed()
		w.cancel()
		w.ssn.Close()
		close(w.buddyNewsChan)
		w.disableWidget()
		w.connectionInUse = false
		w.stopAction.SetEnabled(false)
	}
}

// Funkcja wywoływana po stwierdzeniu, że kolega zamknął połączenie.
func (w *Window) buddyClosedConnection() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.connectionInUse {
		w.dialogConnectionClosed()
		w.cancel()
		w.ssn.Close()
		close(w.buddyNewsChan)
		w.disableWidget()
		w.connectionInUse = false
		w.stopAction.SetEnabled(false)
	}
}

func (w *Window) disableWidget() {
	glib.IdleAdd(func() {
		w.browser.SetSensitive(false)
		w.entry.SetSensitive(false)
		title := w.headerBar.GetTitle()
		w.headerBar.SetTitle(title + " (closed)")
		w.headerBar.SetSubtitle("")
	})
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
			menu.Append("Stop", "win.stop")
			menu.Append("Quit", "win.close")

			stopAction := glib.SimpleActionNew("stop", nil)
			stopAction.Connect("activate", w.stopConnectionAction)
			w.win.AddAction(stopAction)
			w.stopAction = stopAction

			closeAction := glib.SimpleActionNew("close", nil)
			closeAction.Connect("activate", w.closeWindowAction)
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
			tr.Info("%v", w.ctx.Err())
			return
		case msg := <-inChan:
			w.appendTextToBrowser(msg)
		}
	}
}

func (w *Window) netLoop(inChan chan<- news.News, wg *sync.WaitGroup) {
	defer func() {
		w.ssn.Close()
		//wg.Done()
	}()

	for {
		select {
		case <-w.ctx.Done():
			tr.Info("%v", w.ctx.Err())
			return
		default:
			if request := w.ssn.In.Requester.Read(); request != nil {
				switch request.Id {
				case vtc.Message:
					if answer := w.ssn.In.Responder.Send(vtc.Ok, request, nil, nil); answer != nil {
						if msg := news.New(w.buddyName, string(request.Data), false); msg.Valid() {
							inChan <- msg
							continue
						}
					}
				case vtc.Logout:
					if answer := w.ssn.In.Responder.Send(vtc.Ok, request, nil, nil); answer != nil {
					}
				}
			}
			w.buddyClosedConnection()
		}
	}
}
