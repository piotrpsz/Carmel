package chat

import (
	"Carmel/chat/message"
	"Carmel/shared/tr"
	"context"
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/pango"
	"sync"
)

const (
	MyNameTag       = "my_name"
	MyMessageTag    = "my_message"
	OtherNameTag    = "other_name"
	OtherMessageTag = "other_message"
)

var (
	ownNameTagData = map[string]interface{} {
		"foreground": "#FFFAAA",
		"style"     : pango.STYLE_ITALIC,
		"weight"    : pango.WEIGHT_BOLD,
		"font"      : "Italic 10",		//"Sans Italic 12"
	}
	ownMessageData = map[string]interface{} {
		"foreground": "#999999",
	}

	otherNameTagData = map[string]interface{} {
		"foreground"   : "#AAAFFF",
		"style"        : pango.STYLE_ITALIC,
		"weight"       : pango.WEIGHT_BOLD,
		"font"         : "Italic 10",		//"Sans Italic 12"
		"justification": gtk.JUSTIFY_RIGHT,
	}
	otherMessageData = map[string]interface{} {
		"foreground"   : "#FFFFFF",
		"justification": gtk.JUSTIFY_RIGHT,
	}

)

type Window struct {
	app           *gtk.Application
	win           *gtk.ApplicationWindow
	browser       *gtk.TextView
	browserBuffer *gtk.TextBuffer
	entry         *gtk.TextView
	entryBuffer   *gtk.TextBuffer

	// browser tags
	myNameTag       *gtk.TextTag
	myMessageTag    *gtk.TextTag
	otherNameTag    *gtk.TextTag
	otherMessageTag *gtk.TextTag

	// golang
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	browserIn chan message.ChatMessage
}

func New(app *gtk.Application) *Window {
	if win, err := gtk.ApplicationWindowNew(app); tr.IsOK(err) {
		w := &Window{app:app, win:win}
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
	return nil
}

func (w *Window) ShowAll() {
	w.browserIn = make(chan message.ChatMessage)
	w.wg.Add(1)
	go w.browserLoop(w.browserIn)

	w.win.ShowAll()
	w.entry.GrabFocus()
}

func (w *Window) Close() {
	w.cancel()
	w.wg.Wait()
	close(w.browserIn)
	w.win.Close()
}

func (w *Window) createHeaderBar() *gtk.HeaderBar {
	if bar, err := gtk.HeaderBarNew(); tr.IsOK(err) {
		bar.SetShowCloseButton(false)
		bar.SetTitle("Chat with John")
		bar.SetSubtitle("IP 124.35.3.11")
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

					if msg := message.New("john", text, true); msg.Valid() {
						w.browserIn <- msg
						// TODO: send message via network to my partner
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

func (w *Window) appendTextToBrowser(msg message.ChatMessage) {
	nameTag := w.otherNameTag
	messageTag := w.otherMessageTag
	if msg.Own {
		nameTag = w.myNameTag
		messageTag = w.myMessageTag
	}

	w.browserBuffer.InsertWithTag(w.browserBuffer.GetEndIter(), msg.Name + "\n", nameTag)
	w.browserBuffer.InsertWithTag(w.browserBuffer.GetEndIter(), msg.Text + "\n\n", messageTag)

	mark := w.browserBuffer.GetMark("insert")
	w.browser.ScrollToMark(mark, 0.0, true, 0.0, 1.0)
}

func (w *Window) browserLoop(inChan <-chan message.ChatMessage) {
	defer w.wg.Done()

	for {
		select {
			case <- w.ctx.Done():
				fmt.Println(w.ctx.Err())
				return
			case msg := <-inChan:
				w.appendTextToBrowser(msg)
		}
	}
}
