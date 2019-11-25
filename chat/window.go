package chat

import (
	"Carmel/shared/tr"
	"fmt"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

type Window struct {
	app     *gtk.Application
	win     *gtk.ApplicationWindow
	browser *gtk.TextView
	entry   *gtk.TextView
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
					//win.SetResizable(false)
					return w
				}
			}
		}
	}
	return nil
}

func (w *Window) ShowAll() {
	w.win.ShowAll()
	w.entry.GrabFocus()
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
			closeAction.Connect("activate", func() {
				fmt.Println("Close window")
				w.win.Close()
			})
			w.win.AddAction(closeAction)

			btn.SetMenuModel(&menu.MenuModel)
			return btn
		}
	}
	return nil
}

func (w *Window) createContent() *gtk.Box {
	if box, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0); tr.IsOK(err) {
		if browserScroll, err := gtk.ScrolledWindowNew(nil, nil); tr.IsOK(err) {
			if browser, err := gtk.TextViewNew(); tr.IsOK(err) {
				browserScroll.Add(browser)
				if entryScroll, err := gtk.ScrolledWindowNew(nil, nil); tr.IsOK(err) {
					if entry, err := gtk.TextViewNew(); tr.IsOK(err) {
						entryScroll.Add(entry)
						if separator, err := gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL); tr.IsOK(err) {
							browser.SetEditable(false)
							browser.SetCursorVisible(false)

							entry.SetLeftMargin(5)
							entry.SetRightMargin(5)

							box.PackStart(browserScroll, true, true, 1)
							box.PackStart(separator, false, true, 1)
							box.PackStart(entryScroll, false, true, 1)

							w.browser = browser
							w.entry = entry

							return box
						}
					}
				}
			}
		}
	}
	return nil
}