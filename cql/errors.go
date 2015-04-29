package cql

type Errors []error

func (e Errors) Error() string {
	msg := ""

	if len(e) == 1 {
		msg = e[0].Error()
	}
	if len(e) > 1 {
		msg = "Multiple Errors:"
		for _, err := range e {
			msg += "\n" + "  " + err.Error()
		}
	}
	return msg
}
