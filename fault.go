package aqua

import (
	"fmt"
	"strconv"
)

type Fault struct {
	HTTPCode int    `json:"-"`
	Message  string `json:"message"`
	Desc     string `json:"desc"`
	Issue    error  `json:"issue"`
}

func (f Fault) MarshalJSON() ([]byte, error) {

	// TODO: use buffer, and not immutable strings

	b := "{"

	b += fmt.Sprintf(`"message":%s`, strconv.Quote(f.Message))
	b += fmt.Sprintf(`,"desc": %s`, strconv.Quote(f.Desc))
	if f.Issue != nil {
		b += fmt.Sprintf(`,"issue": %s`, strconv.Quote(f.Issue.Error()))
	}
	b += "}"

	return []byte(b), nil
}

func (f *Fault) Set(err error, desc ...string) *Fault {
	if err == nil {
		return nil
	}

	f.Issue = err
	if len(desc) > 0 {
		f.Desc = desc[0]
	}
	if len(desc) > 1 {
		f.Message = desc[1]
	}
	return f
}

// Fault implements error interface
func (f Fault) Error() string {
	return f.Issue.Error()
}
