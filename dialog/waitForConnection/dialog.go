package waitForConnection

import (
	"Carmel/secret"
	"Carmel/shared"
	"Carmel/shared/tr"
	"encoding/hex"
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const (
	descriptionFormat   = "<span style='italic' font_desc='9' foreground='#AAA555'>%s</span>"
	promptFormat        = "<span font_desc='8' foreground='#999999'>%s:</span>"
	valueFormat         = "<span font_desc='11' foreground='#FFFFFF'>%s</span>"
	clipboardDataFormat = "IP: %s\nPort: %s\nName: %s\nPIN: %s\n"
	description         = "The following data should be sent securely\nto your partner so that he can connect with you.\n "
)

type Dialog struct {
	self      *gtk.Dialog
	ipLabel   *gtk.Label
	entryPort *gtk.Entry
	nameLabel *gtk.Label
	pinLabel  *gtk.Label
}

func New(app *gtk.Application) *Dialog {
	if dialog, err := gtk.DialogNew(); tr.IsOK(err) {
		dialog.SetTransientFor(app.GetActiveWindow())
		dialog.SetTitle(shared.AppName)

		instance := &Dialog{self: dialog}
		if contentGrid := instance.createContent(); contentGrid != nil {
			if buttonsBox := instance.createButtons(); buttonsBox != nil {
				if descriptionLabel, err := gtk.LabelNew(""); tr.IsOK(err) {
					if separator, err := gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL); tr.IsOK(err) {
						if box, err := dialog.GetContentArea(); tr.IsOK(err) {
							descriptionLabel.SetMarkup(fmt.Sprintf(descriptionFormat, description))

							box.PackStart(descriptionLabel, false, false, 0)
							box.PackStart(contentGrid, true, true, 0)
							box.PackEnd(separator, true, true, 0)
							box.PackEnd(buttonsBox, false, false, 0)
							return instance
						}
					}
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

						cancelBtn.Connect("clicked", func() {
							d.self.Response(gtk.RESPONSE_CANCEL)
						})
						copyBtn.Connect("clicked", func() {
							if clipboard, err := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD); tr.IsOK(err) {
								ip, _ := d.ipLabel.GetText()
								port, _ := d.entryPort.GetText()
								name, _ := d.nameLabel.GetText()
								pin, _ := d.pinLabel.GetText()

								text := fmt.Sprintf(clipboardDataFormat, ip, port, name, pin)
								clipboard.SetText(text)
							}
						})
						pinBtn.Connect("clicked", func() {
							if pin := createPIN(); pin != "" {
								glib.IdleAdd(d.pinLabel.SetMarkup, fmt.Sprintf(valueFormat, pin))
							}
						})

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
						d.ipLabel = ipLabel
						d.entryPort = portEntry
						d.nameLabel = nameLabel
						d.pinLabel = pinLabel

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
			portEntry.SetText("30303")
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
			pinLabel.SetHAlign(gtk.ALIGN_START)
			pinPrompt.SetMarkup(fmt.Sprintf(promptFormat, "PIN"))
			if pin := createPIN(); pin != "" {
				pinLabel.SetMarkup(fmt.Sprintf(valueFormat, pin))
			}
			return pinPrompt, pinLabel
		}
	}
	return nil, nil
}

func createPIN() string {
	if data := secret.RandomBytes(5); data != nil {
		ndigits := hex.EncodedLen(len(data))
		buffer := make([]byte, ndigits)
		hex.Encode(buffer, data)
		return string(buffer)
	}
	return ""
}
