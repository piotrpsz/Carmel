package dialogWithOneField

import (
	"Carmel/shared"
	"Carmel/shared/tr"
	"fmt"
	"github.com/gotk3/gotk3/gtk"
)

type DialogWithOneField struct {
	self             *gtk.Dialog
	descriptionLabel *gtk.Label
	promptLabel      *gtk.Label
	entry            *gtk.Entry
}

const (
	descriptionFormat = "<span  style='italic' font_desc='9' foreground='#AAA555'>%s</span>"
	promptFormat = "<span  font_desc='10' foreground='#999999'>%s</span>"
)

func New(app *gtk.Application) *DialogWithOneField {
	if dialog, err := gtk.DialogNew(); tr.IsOK(err) {
		dialog.SetTransientFor(app.GetActiveWindow())
		dialog.SetTitle(shared.AppName)
		if acceptButton, err := dialog.AddButton("Accept", gtk.RESPONSE_APPLY); tr.IsOK(err) {
			if cancelButton, err := dialog.AddButton("Cancel", gtk.RESPONSE_CANCEL); tr.IsOK(err) {
				acceptButton.SetRelief(gtk.RELIEF_HALF)
				cancelButton.SetRelief(gtk.RELIEF_HALF)
				instance := &DialogWithOneField{self: dialog}

				if box, err := dialog.GetContentArea(); tr.IsOK(err) {
					if instance.setContentIn(box) {
						dialog.SetPosition(gtk.WIN_POS_CENTER)
						return instance
					}
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

func (dialog *DialogWithOneField) setContentIn(box *gtk.Box) bool {
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

						box.Add(grid)
						return true
					}
				}
			}
		}
	}
	return false
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
		return text
	}
	return ""
}
