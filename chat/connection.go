package chat

import (
	"Carmel/connector/message"
	"Carmel/connector/session"
	"Carmel/shared"
	"Carmel/shared/vtc"
	"fmt"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

func finalInit(app *gtk.Application, role vtc.RoleType, buddyName string, ssn *session.Session) bool {
	// Wszystko do tej pory poszło dobrze, ale może się okazać że
	// nie mamy publicznego klucza RSA dla wskazanej osoby.
	// Jeśli tak by było to dupa.
	if ssn.In.Enigma.SetBuddyRSAPublicKey(buddyName) {
		// Możemy kontynuuować komunikację, ale czy na pewno chcemy?
		if dialogCanConnectWith(app, buddyName) {
			switch role {
			case vtc.Server:
				if !sendAcceptance(ssn) {
					return false
				}
				if !ssn.SendKeys() {
					return false
				}
				if !ssn.ExchangeBlockIdentifiersAsServer() {
					return false
				}
			case vtc.Client:
				if !ssn.ReadKeys() {
					return false
				}
				if !ssn.ExchangeBlockIdentifiersAsClient() {
					return false
				}
			}
			return true
		}
	}
	return false
}

func sendAcceptance(ssn *session.Session) bool {
	msg := message.NewWithType(vtc.Answer)
	msg.Id = vtc.Login
	msg.Status = vtc.Accepted
	msg.Tstamp = shared.Now()
	if data := msg.ToJsonSnapped(); data != nil {
		if cipher := ssn.In.Enigma.EncryptRSA(data); cipher != nil {
			return ssn.In.Requester.SendRawMessage(cipher)
		}
	}
	return false
}

func dialogCanConnectWith(app *gtk.Application, buddyName string) bool {
	if dialog := gtk.MessageDialogNew(app.GetActiveWindow(), gtk.DIALOG_MODAL, gtk.MESSAGE_QUESTION, gtk.BUTTONS_YES_NO, ""); dialog != nil {
		defer dialog.Destroy()
		dialog.FormatSecondaryText(fmt.Sprintf(canConnectFormat, buddyName))
		if dialog.Run() == gtk.RESPONSE_YES {
			return true
		}
	}
	return false
}

// Wysyła informację o zakończeniu sesji.
func (w *Window) sendDisconnectRequest() bool {
	if request := w.ssn.Out.Requester.Send(vtc.Logout, nil, nil); request != nil {
		if answer := w.ssn.Out.Responder.Read(request); answer != nil {
			if answer.Status == vtc.Ok {
				return true
			}
		}
	}
	return false
}

func (w *Window) dialogConnectionClosed() {
	glib.IdleAdd(func() {
		if dialog := gtk.MessageDialogNew(w.win, gtk.DIALOG_MODAL, gtk.MESSAGE_INFO, gtk.BUTTONS_CLOSE, ""); dialog != nil {
			defer dialog.Destroy()
			dialog.FormatSecondaryText(fmt.Sprintf(connectionClosed, w.buddyName))
			dialog.Run()
		}
	})
}
