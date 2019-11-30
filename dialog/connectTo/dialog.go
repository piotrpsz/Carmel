package connectTo

import (
	"Carmel/shared"
	"Carmel/shared/tr"
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"strings"
)

const (
	dialogTitle       = "connect to"
	descriptionFormat = "<span style='italic' font_desc='9' foreground='#AAA555'>%s</span>"
	description       = "Here you should enter (or copy from the clipboard)\nthe data received from the partner."
	promptFormat      = "<span font_desc='8' foreground='#999999'>%s:</span>"

	// button titles
	startBtnTitle  = "start"
	cancelBtnTitle = "cancel"
	copyBtnTtile   = "copy"

	// tooltips
	startSetTooltip = "Connect to the server."
	copyTooltip     = "Copy data from the clipboard"
	cancelTooltip   = "Break action and return"
	ipTooltip       = "IP address of the server"
	portTooltip     = "port number on which the server listens"
	nameTooltip     = "user name to which you would like to connect"
	pinTooltip      = "pin needed to establish connection to the server"
)

type Dialog struct {
	self              *gtk.Dialog
	spinner           *gtk.Spinner
	ipEntry           *gtk.Entry
	portEntry         *gtk.Entry
	nameEntry         *gtk.Entry
	pinEntry          *gtk.Entry
	startBtn          *gtk.Button
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
	if startBtn, err := gtk.ButtonNewWithLabel(startBtnTitle); tr.IsOK(err) {
		if cancelBtn, err := gtk.ButtonNewWithLabel(cancelBtnTitle); tr.IsOK(err) {
			if copyBtn, err := gtk.ButtonNewWithLabel(copyBtnTtile); tr.IsOK(err) {
				if box, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 1); tr.IsOK(err) {
					startBtn.SetTooltipText(startSetTooltip)
					copyBtn.SetTooltipText(copyTooltip)
					cancelBtn.SetTooltipText(cancelTooltip)

					d.startBtn = startBtn
					d.copyBtn = copyBtn

					box.PackStart(startBtn, true, true, 2)
					box.PackStart(copyBtn, true, true, 2)
					box.PackStart(cancelBtn, true, true, 2)

					startBtn.Connect("clicked", func() {
						d.connectionAttempt = true
						d.spinner.Start()
						d.enableDisable(false)
					})
					cancelBtn.Connect("clicked", func() {
						if d.connectionAttempt {
							d.connectionAttempt = false
							d.spinner.Stop()
							d.enableDisable(true)
							return
						}
						d.self.Response(gtk.RESPONSE_CANCEL)
					})
					copyBtn.Connect("clicked", func() {
						if clipboard, err := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD); tr.IsOK(err) {
							if clipboard.WaitIsTextAvailable() {
								if text, err := clipboard.WaitForText(); tr.IsOK(err) {
									d.useDataFromClipboard(strings.TrimSpace(text))
								}
							}
						}
					})

					return box
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

		if ipPrompt, ipEntry := createIPWidgets(); ipPrompt != nil {
			if portPrompt, portEntry := createPortWidgets(); portPrompt != nil {
				if namePrompt, nameEntry := createUsernameWidgets(); namePrompt != nil {
					if pinPrompt, pinEntry := createPINWidgets(); pinPrompt != nil {
						if spinner, err := gtk.SpinnerNew(); tr.IsOK(err) {
							ipEntry.SetTooltipText(ipTooltip)
							portEntry.SetTooltipText(portTooltip)
							nameEntry.SetTooltipText(nameTooltip)
							pinEntry.SetTooltipText(pinTooltip)

							d.ipEntry = ipEntry
							d.portEntry = portEntry
							d.nameEntry = nameEntry
							d.pinEntry = pinEntry
							d.spinner = spinner

							y := 0
							grid.Attach(d.spinner, 0, y, 2, 1)
							y++
							grid.Attach(ipPrompt, 0, y, 1, 1)
							grid.Attach(ipEntry, 1, y, 1, 1)
							y++
							grid.Attach(portPrompt, 0, y, 1, 1)
							grid.Attach(portEntry, 1, y, 1, 1)
							y++
							grid.Attach(namePrompt, 0, y, 1, 1)
							grid.Attach(nameEntry, 1, y, 1, 1)
							y++
							grid.Attach(pinPrompt, 0, y, 1, 1)
							grid.Attach(pinEntry, 1, y, 1, 1)
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

	d.ipEntry.SetSensitive(state)
	d.portEntry.SetSensitive(state)
	d.nameEntry.SetSensitive(state)
	d.pinEntry.SetSensitive(state)

	d.startBtn.SetSensitive(state)
	d.copyBtn.SetSensitive(state)
}

func (d *Dialog) useDataFromClipboard(text string) {
	for _, line := range strings.Split(text, "\n") {
		text := strings.TrimSpace(line)
		if strings.HasPrefix(text, shared.IPClipboardMark) {
			value := text[len(shared.IPClipboardMark):]
			d.ipEntry.SetText(strings.TrimSpace(value))
		}
		if strings.HasPrefix(text, shared.PortClipboardMark) {
			value := text[len(shared.PortClipboardMark):]
			d.portEntry.SetText(strings.TrimSpace(value))
		}
		if strings.HasPrefix(text, shared.IPClipboardMark) {
			value := text[len(shared.IPClipboardMark):]
			d.ipEntry.SetText(strings.TrimSpace(value))
		}
		if strings.HasPrefix(text, shared.NameClipboardMark) {
			value := text[len(shared.NameClipboardMark):]
			d.nameEntry.SetText(strings.TrimSpace(value))
		}
		if strings.HasPrefix(text, shared.PINClipboardMark) {
			value := text[len(shared.PINClipboardMark):]
			d.pinEntry.SetText(strings.TrimSpace(value))
		}
	}
}

func createIPWidgets() (*gtk.Label, *gtk.Entry) {
	if ipPrompt, err := gtk.LabelNew(""); tr.IsOK(err) {
		if ipEntry, err := gtk.EntryNew(); tr.IsOK(err) {
			ipPrompt.SetHAlign(gtk.ALIGN_END)
			ipPrompt.SetMarkup(fmt.Sprintf(promptFormat, "IP"))
			return ipPrompt, ipEntry
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

func createUsernameWidgets() (*gtk.Label, *gtk.Entry) {
	if namePrompt, err := gtk.LabelNew(""); tr.IsOK(err) {
		if nameEntry, err := gtk.EntryNew(); tr.IsOK(err) {
			namePrompt.SetHAlign(gtk.ALIGN_END)
			namePrompt.SetMarkup(fmt.Sprintf(promptFormat, "Name"))
			return namePrompt, nameEntry
		}
	}
	return nil, nil
}

func createPINWidgets() (*gtk.Label, *gtk.Entry) {
	if pinPrompt, err := gtk.LabelNew(""); tr.IsOK(err) {
		if pinEntry, err := gtk.EntryNew(); tr.IsOK(err) {
			pinPrompt.SetHAlign(gtk.ALIGN_END)
			pinPrompt.SetMarkup(fmt.Sprintf(promptFormat, "PIN"))
			return pinPrompt, pinEntry
		}
	}
	return nil, nil
}