package shared

import "github.com/dracory/hb"

// ErrorAlert returns an inline HTML error alert for the given message.
func ErrorAlert(message string) string {
	return hb.Div().
		Class("alert alert-danger").
		Style("margin: 20px;").
		HTML(message).
		ToHTML()
}
