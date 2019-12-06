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

package dialogWithOneField

import (
	"Carmel/shared"
	"Carmel/shared/tr"
	"fmt"
	"github.com/gotk3/gotk3/gtk"
	"strings"
)

type ValidationResult uint8

const (
	Ok ValidationResult = iota << 1
	EmptyStringNotAllowed
	SpacesNotAllowed
	DigitAtStartNotAllowed
)

const (
	FailedValidation     = "Entered text is incorrect"
	ContentWithSpaces    = "The text cannot contain spaces."
	EmptyContennt        = "The text can't be empty."
	FirstDigitNotAllowed = "First character can't be a digit"
)

type DialogWithOneField struct {
	self             *gtk.Dialog
	descriptionLabel *gtk.Label
	promptLabel      *gtk.Label
	entry            *gtk.Entry
}

const (
	descriptionFormat = "<span style='italic' font_desc='9' foreground='#AAA555'>%s</span>"
	promptFormat      = "<span font_desc='10' foreground='#999999'>%s</span>"
)

func New(app *gtk.Application, validateFn func(string) ValidationResult) *DialogWithOneField {
	if dialog, err := gtk.DialogNew(); tr.IsOK(err) {
		dialog.SetTransientFor(app.GetActiveWindow())
		dialog.SetTitle(shared.AppName)

		instance := &DialogWithOneField{self: dialog}

		if contentArea := instance.setContent(); contentArea != nil {
			if buttonArea := instance.buttonArea(validateFn); buttonArea != nil {
				if box, err := dialog.GetContentArea(); tr.IsOK(err) {
					box.PackStart(contentArea, true, true, 0)
					box.PackEnd(buttonArea, false, false, 0)
					return instance
				}
			}
		}
	}
	return nil
}

func (dialog *DialogWithOneField) ShowAll() {
	dialog.self.ShowAll()
}

func (dialog *DialogWithOneField) Run() gtk.ResponseType {
	return dialog.self.Run()
}

func (dialog *DialogWithOneField) Destroy() {
	dialog.self.Destroy()
}

func (dialog *DialogWithOneField) buttonArea(validateFn func(string) ValidationResult) *gtk.Box {
	if acceptButton, err := gtk.ButtonNewWithLabel("Accept"); tr.IsOK(err) {
		if cancelButton, err := gtk.ButtonNewWithLabel("Cancel"); tr.IsOK(err) {
			if box, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 2); tr.IsOK(err) {

				cancelButton.Connect("clicked", func() {
					dialog.self.Response(gtk.RESPONSE_CANCEL)
				})
				acceptButton.Connect("clicked", func() {
					text, _ := dialog.entry.GetText()
					result := validateFn(text)
					if result == Ok {
						dialog.self.Response(gtk.RESPONSE_ACCEPT)
						return
					}
					if dialog := gtk.MessageDialogNew(dialog.self, gtk.DIALOG_MODAL, gtk.MESSAGE_WARNING, gtk.BUTTONS_OK, FailedValidation); dialog != nil {
						defer dialog.Destroy()

						switch result {
						case EmptyStringNotAllowed:
							dialog.FormatSecondaryText(EmptyContennt)
						case SpacesNotAllowed:
							dialog.FormatSecondaryText(ContentWithSpaces)
						case DigitAtStartNotAllowed:
							dialog.FormatSecondaryText(FirstDigitNotAllowed)
						}
						dialog.Run()
					}
					dialog.entry.GrabFocus()
				})

				box.PackEnd(cancelButton, false, false, 2)
				box.PackEnd(acceptButton, false, false, 2)
				return box
			}
		}
	}
	return nil
}

func (dialog *DialogWithOneField) setContent() *gtk.Grid {
	if grid, err := gtk.GridNew(); tr.IsOK(err) {
		if promptLabel, err := gtk.LabelNew(""); tr.IsOK(err) {
			if entry, err := gtk.EntryNew(); tr.IsOK(err) {
				entry.SetWidthChars(10)
				if descriptionLabel, err := gtk.LabelNew(""); tr.IsOK(err) {
					if separator, err := gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL); tr.IsOK(err) {
						dialog.promptLabel = promptLabel
						dialog.entry = entry
						dialog.descriptionLabel = descriptionLabel

						grid.SetBorderWidth(8)
						grid.SetRowSpacing(8)
						grid.SetColumnSpacing(8)

						grid.Attach(promptLabel, 0, 0, 1, 1)
						grid.Attach(entry, 1, 0, 1, 1)
						grid.Attach(descriptionLabel, 1, 1, 1, 1)
						grid.Attach(separator, 0, 2, 2, 1)

						return grid
					}
				}
			}
		}
	}
	return nil
}

func (dialog *DialogWithOneField) SetPrompt(text string) {
	markup := fmt.Sprintf(promptFormat, text)
	dialog.promptLabel.SetMarkup(markup)
}

func (dialog *DialogWithOneField) SetDescription(text string) {
	markup := fmt.Sprintf(descriptionFormat, text)
	dialog.descriptionLabel.SetMarkup(markup)
}

func (dialog *DialogWithOneField) SetValue(text string) {
	dialog.entry.SetText(text)
}

func (dialog *DialogWithOneField) GetValue() string {
	if text, err := dialog.entry.GetText(); tr.IsOK(err) {
		return strings.ToLower(text)
	}
	return ""
}
