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

package waitForConnection

import (
	"Carmel/chat"
	"Carmel/connector/message"
	"Carmel/connector/session"
	"Carmel/secret"
	"Carmel/shared"
	"Carmel/shared/tr"
	"Carmel/shared/vtc"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"os"
	"strconv"
	"strings"
	"sync"
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

	connectionCanceled  = "Canceled"
	connectionError     = "Unknown error"
	connectionSecurity  = "Security breach"
	connectionMsgFormat = "Connection failed on port:  %d"
)

type Dialog struct {
	self              *gtk.Dialog
	app               *gtk.Application
	ipLabel           *gtk.Label
	portEntry         *gtk.Entry
	nameLabel         *gtk.Label
	pinLabel          *gtk.Label
	spinner           *gtk.Spinner
	startBtn          *gtk.Button
	pinBtn            *gtk.Button
	copyBtn           *gtk.Button
	cancelBtn         *gtk.Button
	connectionAttempt bool
	ctx               context.Context
	cancel            context.CancelFunc
}

func New(app *gtk.Application) *Dialog {
	if dialog, err := gtk.DialogNew(); tr.IsOK(err) {
		dialog.SetTransientFor(app.GetActiveWindow())
		dialog.SetTitle(dialogTitle)

		instance := &Dialog{self: dialog, app: app}

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
		if d.cancelBtn, err = gtk.ButtonNewWithLabel(cancelBtnTitle); tr.IsOK(err) {
			if d.copyBtn, err = gtk.ButtonNewWithLabel(copyBtnTtile); tr.IsOK(err) {
				if d.pinBtn, err = gtk.ButtonNewWithLabel(pinBtnTitle); tr.IsOK(err) {
					if box, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 1); tr.IsOK(err) {
						// tooltips
						d.startBtn.SetTooltipText(startTooltip)
						d.copyBtn.SetTooltipText(copyTooltip)
						d.pinBtn.SetTooltipText(pinTooltip)
						d.cancelBtn.SetTooltipText(cancelTooltip)

						// pack widgets
						box.PackStart(d.startBtn, true, true, 2)
						box.PackStart(d.pinBtn, true, true, 2)
						box.PackStart(d.copyBtn, true, true, 2)
						box.PackStart(d.cancelBtn, true, true, 2)

						// handle button events
						d.startBtn.Connect("clicked", d.start)
						d.cancelBtn.Connect("clicked", d.stop)
						d.copyBtn.Connect("clicked", d.copy)

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

// Sprawdzenie poprawności danych w polu 'port'.
func (d *Dialog) validData() (bool, int, string) {
	if portAsString, err := d.portEntry.GetText(); tr.IsOK(err) && shared.OnlyDigits(portAsString) {
		if port, err := strconv.Atoi(portAsString); tr.IsOK(err) {
			if pin, err := d.pinLabel.GetText(); tr.IsOK(err) && shared.OnlyHexDigits(pin) {
				return true, port, pin
			}
		}
	}

	return false, 0, ""
}

// Serwer rozpoczyna nasłuchiwanie nadchodzących połączeń od klienta.
func (d *Dialog) start() {
	ok, port, pin := d.validData()
	if !ok {
		return
	}
	d.connectionAttempt = true
	d.spinner.Start()
	d.enableDisable(false)

	if ssn := session.ServerNew(port); ssn != nil {
		d.ctx, d.cancel = context.WithCancel(context.Background())
		go func() {
			var failureReason string
			var faildePort int
			var wg sync.WaitGroup
			var stateIn, stateOut vtc.OperationStatusType
			state := vtc.Ok

			wg.Add(2)
			go func() {
				stateIn = ssn.In.Run(d.ctx, &wg)
			}()
			go func() {
				stateOut = ssn.Out.Run(d.ctx, &wg)
			}()
			wg.Wait()

			if stateIn == vtc.Ok && stateOut == vtc.Ok {
				if buddyName := d.initConnection(ssn, pin); buddyName != "" {
					glib.IdleAdd(func() {
						d.self.Destroy()
						if chatter := chat.New(d.app, vtc.Server, buddyName, ssn); chatter != nil {
							chatter.ShowAll()
						}
					})
					return
				}
				state = vtc.SecurityBreach
				ssn.Close()
				ssn = nil
			}

			if state == vtc.Ok {
				if stateIn != vtc.Ok {
					faildePort = ssn.In.ServerPort
					state = stateIn
				} else {
					faildePort = ssn.Out.ServerPort
					state = stateOut
				}
			}

			switch state {
			case vtc.Cancel:
				failureReason = connectionCanceled
			case vtc.SecurityBreach:
				failureReason = connectionSecurity
			default:
				failureReason = connectionError
			}

			glib.IdleAdd(func() {
				d.spinner.Stop()
				if errDialog := gtk.MessageDialogNew(d.self, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, failureReason); errDialog != nil {
					defer func() {
						errDialog.Destroy()
						d.continueEdition()
					}()
					if state != vtc.SecurityBreach {
						errDialog.FormatSecondaryText(fmt.Sprintf(connectionMsgFormat, faildePort))
					}
					errDialog.Run()
					if state == vtc.SecurityBreach {
						os.Exit(1)
					}
				}
			})
		}()
		return
	}

	d.stop()
}

// Odczyt od klienta żądania inicjacyjnego.
// Operacja przesyłu danych szyfrowana jest w całości kluczem RSA.
func (d *Dialog) initConnection(ssn *session.Session, pin string) string {
	if data := ssn.In.Requester.ReadRawMessage(); data != nil {
		if plain := ssn.In.Enigma.DecryptRsa(data); plain != nil {
			if msg := message.NewFromJson(plain); msg != nil {
				if msg.Id == vtc.Login {
					if items := strings.Split(string(msg.Data), "|"); len(items) == 2 {
						if items[1] == shared.MyUserName && pin == string(msg.Extra) {
							return items[0]
						}
					}
				}
			}
		}
	}
	return ""
}

func (d *Dialog) continueEdition() {
	d.connectionAttempt = false
	d.enableDisable(true)
}

// Przerywa nasłuchiwanie serwera.
// Użytkownik nacisnął kliwisz 'Cancel'.
func (d *Dialog) stop() {
	if d.connectionAttempt {
		glib.IdleAdd(func() {
			d.spinner.Stop()
			d.cancelBtn.SetSensitive(false)
		})
		d.cancel()
		return
	}
	d.self.Response(gtk.RESPONSE_CANCEL)
}

// Kopiuje zawartość pół edycyjnych do schowka w określonym formacie.
// Ze schowka użytkownik może skopiować te dane np. do e-maila i je przesłać.
func (d *Dialog) copy() {
	if clipboard, err := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD); tr.IsOK(err) {
		if ip, err := d.ipLabel.GetText(); tr.IsOK(err) {
			if port, err := d.portEntry.GetText(); tr.IsOK(err) {
				if name, err := d.nameLabel.GetText(); tr.IsOK(err) {
					if pin, err := d.pinLabel.GetText(); tr.IsOK(err) {
						text := fmt.Sprintf(clipboardDataFormat, ip, port, name, pin)
						clipboard.SetText(text)
					}
				}
			}
		}
	}
}

// Włączenie/wyłączenie możliwości edycji.
func (d *Dialog) enableDisable(state bool) {
	glib.IdleAdd(func() {
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
		d.cancelBtn.SetSensitive(true)
		d.portEntry.GrabFocusWithoutSelecting()
	})
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
			portEntry.SetText("40404")
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
