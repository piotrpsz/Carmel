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
	dialogTitle         = "wait for connection"
	descriptionFormat   = "<span style='italic' font_desc='9' foreground='#AAA555'>%s</span>"
	promptFormat        = "<span font_desc='8' foreground='#999999'>%s:</span>"
	enabledValueFormat  = "<span font_desc='11' foreground='#FFFFFF'>%s</span>"
	disabledValueFormat = "<span font_desc='11' foreground='#999999'>%s</span>"
	clipboardDataFormat = "IP: %s\nPort: %s\nName: %s\nPIN: %s\n"
	description         = "The following data should be sent securely\nto your partner so that he can connect with you."

	// button titles
	startBtnTitle  = "start"
	cancelBtnTitle = "cancel"
	pinBtnTitle    = "pin"
	copyBtnTtile   = "copy"

	// tooltips
	pinTooltip    = "generate new random PIN number"
	cancelTooltip = "break action and return"
	copyTooltip   = "copy data to the clipboard"
	startTooltip  = "start waiting for connection"
)

type Dialog struct {
	self              *gtk.Dialog
	ipLabel           *gtk.Label
	portEntry         *gtk.Entry
	nameLabel         *gtk.Label
	pinLabel          *gtk.Label
	spinner           *gtk.Spinner
	startBtn          *gtk.Button
	pinBtn            *gtk.Button
	copyBtn           *gtk.Button
	connectionAttempt bool
}

func New(app *gtk.Application) *Dialog {
	if dialog, err := gtk.DialogNew(); tr.IsOK(err) {
		dialog.SetTransientFor(app.GetActiveWindow())
		dialog.SetTitle(dialogTitle)

		instance := &Dialog{self: dialog}
		if contentGrid := instance.createContent(); contentGrid != nil {
			if buttonsBox := instance.createButtons(); buttonsBox != nil {
				if descriptionLabel, err := gtk.LabelNew(""); tr.IsOK(err) {
					if separator, err := gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL); tr.IsOK(err) {
						if box, err := dialog.GetContentArea(); tr.IsOK(err) {
							descriptionLabel.SetMarkup(fmt.Sprintf(descriptionFormat, description))

							box.SetBorderWidth(6)
							box.SetSpacing(4)

							box.PackStart(descriptionLabel, false, false, 0)
							box.PackStart(contentGrid, true, true, 0)
							box.PackStart(separator, true, true, 0)
							box.PackStart(buttonsBox, false, false, 0)
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
	d.self.SetResizable(false)
}

func (d *Dialog) Run() gtk.ResponseType {
	return d.self.Run()
}

func (d *Dialog) Destroy() {
	d.self.Destroy()
}

func (d *Dialog) createButtons() *gtk.Box {
	var err error

	if d.startBtn, err = gtk.ButtonNewWithLabel(startBtnTitle); tr.IsOK(err) {
		if cancelBtn, err := gtk.ButtonNewWithLabel(cancelBtnTitle); tr.IsOK(err) {
			if d.copyBtn, err = gtk.ButtonNewWithLabel(copyBtnTtile); tr.IsOK(err) {
				if d.pinBtn, err = gtk.ButtonNewWithLabel(pinBtnTitle); tr.IsOK(err) {
					if box, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 1); tr.IsOK(err) {
						// tooltips
						d.startBtn.SetTooltipText(startTooltip)
						d.copyBtn.SetTooltipText(copyTooltip)
						d.pinBtn.SetTooltipText(pinTooltip)
						cancelBtn.SetTooltipText(cancelTooltip)

						// pack widgets
						box.PackStart(d.startBtn, true, true, 2)
						box.PackStart(d.pinBtn, true, true, 2)
						box.PackStart(d.copyBtn, true, true, 2)
						box.PackStart(cancelBtn, true, true, 2)

						// handle button events
						d.startBtn.Connect("clicked", func() {
							d.enableDisable(false)
							d.connectionAttempt = true
							d.spinner.Start()
						})

						cancelBtn.Connect("clicked", func() {
							if d.connectionAttempt {
								d.connectionAttempt = false
								d.enableDisable(true)
								d.spinner.Stop()
								d.portEntry.GrabFocusWithoutSelecting()
								return
							}
							d.self.Response(gtk.RESPONSE_CANCEL)
						})
						d.copyBtn.Connect("clicked", func() {
							if clipboard, err := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD); tr.IsOK(err) {
								ip, _ := d.ipLabel.GetText()
								port, _ := d.portEntry.GetText()
								name, _ := d.nameLabel.GetText()
								pin, _ := d.pinLabel.GetText()

								text := fmt.Sprintf(clipboardDataFormat, ip, port, name, pin)
								clipboard.SetText(text)
							}
						})
						d.pinBtn.Connect("clicked", func() {
							if pin := createPIN(); pin != "" {
								glib.IdleAdd(d.pinLabel.SetMarkup, fmt.Sprintf(enabledValueFormat, pin))
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
						if spinner, err := gtk.SpinnerNew(); tr.IsOK(err) {
							d.ipLabel = ipLabel
							d.portEntry = portEntry
							d.nameLabel = nameLabel
							d.pinLabel = pinLabel
							d.spinner = spinner

							y := 0
							grid.Attach(spinner, 0, y, 2, 1)
							y++
							grid.Attach(ipPrompt, 0, y, 1, 1)
							grid.Attach(ipLabel, 1, y, 1, 1)
							y++
							grid.Attach(portPrompt, 0, y, 1, 1)
							grid.Attach(portEntry, 1, y, 1, 1)
							y++
							grid.Attach(namePrompt, 0, y, 1, 1)
							grid.Attach(nameLabel, 1, y, 1, 1)
							y++
							grid.Attach(pinPrompt, 0, y, 1, 1)
							grid.Attach(pinLabel, 1, y, 1, 1)
							return grid
						}
					}
				}
			}
		}
	}
	return nil
}

func (d *Dialog) enableDisable(state bool) {
	format := disabledValueFormat
	if state {
		format = enabledValueFormat
	}
	text, _ := d.ipLabel.GetText()
	d.ipLabel.SetMarkup(fmt.Sprintf(format, text))
	text, _ = d.nameLabel.GetText()
	d.nameLabel.SetMarkup(fmt.Sprintf(format, text))
	text, _ = d.pinLabel.GetText()
	d.pinLabel.SetMarkup(fmt.Sprintf(format, text))

	d.portEntry.SetSensitive(state)
	d.startBtn.SetSensitive(state)
	d.copyBtn.SetSensitive(state)
	d.pinBtn.SetSensitive(state)
}

func createIPWidgets() (*gtk.Label, *gtk.Label) {
	if ipPrompt, err := gtk.LabelNew(""); tr.IsOK(err) {
		if ipLabel, err := gtk.LabelNew(""); tr.IsOK(err) {
			ipPrompt.SetHAlign(gtk.ALIGN_END)
			ipLabel.SetHAlign(gtk.ALIGN_START)
			ipPrompt.SetMarkup(fmt.Sprintf(promptFormat, "IP"))
			ipLabel.SetMarkup(fmt.Sprintf(enabledValueFormat, shared.MyIPAddr))

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
			namePrompt.SetMarkup(fmt.Sprintf(promptFormat, "Name"))
			nameLabel.SetMarkup(fmt.Sprintf(enabledValueFormat, shared.MyUserName))
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
				pinLabel.SetMarkup(fmt.Sprintf(enabledValueFormat, pin))
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