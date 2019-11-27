package waitForConnection

import (
	"Carmel/shared"
	"Carmel/shared/tr"
	"fmt"
	"github.com/gotk3/gotk3/gtk"
)

const (
	promptFormat = "<span font_desc='8' foreground='#999999'>%s:</span>"
	valueFormat  = "<span font_desc='11' foreground='#FFFFFF'>%s</span>"
)

type Dialog struct {
	self      *gtk.Dialog
	entryPort *gtk.Entry
	pinLabel  *gtk.Label
}

func New(app *gtk.Application) *Dialog {
	if dialog, err := gtk.DialogNew(); tr.IsOK(err) {
		dialog.SetTransientFor(app.GetActiveWindow())
		dialog.SetTitle(shared.AppName)

		instance := &Dialog{self: dialog}
		if contentGrid := instance.createContent(); contentGrid != nil {
			if buttonsBox := instance.createButtons(); buttonsBox != nil {
				if box, err := dialog.GetContentArea(); tr.IsOK(err) {
					box.PackStart(contentGrid, true, true, 0)
					box.PackEnd(buttonsBox, false, false, 0)
					return instance
				}
			}
		}
	}
	return nil
}

func (d *Dialog) ShowAll() {
	d.self.ShowAll()
}

func (d *Dialog) Run() gtk.ResponseType {
	return d.self.Run()
}

func (d *Dialog) Destroy() {
	d.self.Destroy()
}

//shared.MyIPAddr = text

func (d *Dialog) createButtons() *gtk.Box {
	if startBtn, err := gtk.ButtonNewWithLabel("start"); tr.IsOK(err) {
		if cancelBtn, err := gtk.ButtonNewWithLabel("cancel"); tr.IsOK(err) {
			if copyBtn, err := gtk.ButtonNewWithLabel("to clipboard"); tr.IsOK(err) {
				if pinBtn, err := gtk.ButtonNewWithLabel("another PIN"); tr.IsOK(err) {
					if box, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 1); tr.IsOK(err) {
						box.PackStart(pinBtn, false, false, 2)
						box.PackStart(startBtn, false, false, 2)
						box.PackStart(copyBtn, false, false, 2)
						box.PackStart(cancelBtn, false, false, 2)
						return box
					}
				}
			}
		}
	}
	return nil
}

func (d *Dialog) createContent() *gtk.Grid {
	if grid, err := gtk.GridNew(); tr.IsOK(err) {
		grid.SetBorderWidth(8)
		grid.SetRowSpacing(8)
		grid.SetColumnSpacing(8)

		if ipPrompt, ipLabel := createIPWidgets(); ipPrompt != nil {
			if portPrompt, portEntry := createPortWidgets(); portPrompt != nil {
				if namePrompt, nameLabel := createUsernameWidgets(); namePrompt != nil {
					if pinPrompt, pinLabel := createPINWidgets(); pinPrompt != nil {
						grid.Attach(ipPrompt, 0, 0, 1, 1)
						grid.Attach(ipLabel, 1, 0, 1, 1)
						grid.Attach(portPrompt, 0, 2, 1, 1)
						grid.Attach(portEntry, 1, 2, 1, 1)
						grid.Attach(namePrompt, 0, 3, 1, 1)
						grid.Attach(nameLabel, 1, 3, 1, 1)
						grid.Attach(pinPrompt, 0, 4, 1, 1)
						grid.Attach(pinLabel, 1, 4, 1, 1)
						return grid
					}
				}
			}
		}
	}
	return nil
}

func createIPWidgets() (*gtk.Label, *gtk.Label) {
	if ipPrompt, err := gtk.LabelNew(""); tr.IsOK(err) {
		if ipLabel, err := gtk.LabelNew(""); tr.IsOK(err) {
			ipPrompt.SetHAlign(gtk.ALIGN_END)
			ipLabel.SetHAlign(gtk.ALIGN_START)
			ipPrompt.SetMarkup(fmt.Sprintf(promptFormat, "IP"))
			ipLabel.SetMarkup(fmt.Sprintf(valueFormat, shared.MyIPAddr))
			return ipPrompt, ipLabel
		}
	}
	return nil, nil
}

func createPortWidgets() (*gtk.Label, *gtk.Entry) {
	if portPrompt, err := gtk.LabelNew(""); tr.IsOK(err) {
		if portEntry, err := gtk.EntryNew(); tr.IsOK(err) {
			portPrompt.SetHAlign(gtk.ALIGN_END)
			portPrompt.SetMarkup(fmt.Sprintf(promptFormat, "Port"))
			return portPrompt, portEntry
		}
	}
	return nil, nil
}

func createUsernameWidgets() (*gtk.Label, *gtk.Label) {
	if namePrompt, err := gtk.LabelNew(""); tr.IsOK(err) {
		if nameLabel, err := gtk.LabelNew(""); tr.IsOK(err) {
			namePrompt.SetHAlign(gtk.ALIGN_END)
			nameLabel.SetHAlign(gtk.ALIGN_START)
			namePrompt.SetMarkup(fmt.Sprintf(promptFormat, "My name"))
			nameLabel.SetMarkup(fmt.Sprintf(valueFormat, shared.MyUserName))
			return namePrompt, nameLabel
		}
	}
	return nil, nil
}

func createPINWidgets() (*gtk.Label, *gtk.Label) {
	if pinPrompt, err := gtk.LabelNew(""); tr.IsOK(err) {
		if pinLabel, err := gtk.LabelNew(""); tr.IsOK(err) {
			pinPrompt.SetHAlign(gtk.ALIGN_END)
			pinPrompt.SetMarkup(fmt.Sprintf(promptFormat, "PIN"))
			return pinPrompt, pinLabel
		}
	}
	return nil, nil
}
