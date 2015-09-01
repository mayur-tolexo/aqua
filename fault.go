package aqua

import (
	"encoding/json"
	"fmt"
)

type Fault struct {
	message string
	issue   error
}

func NewFault(e error, msg string) Fault {
	f := Fault{
		message: msg,
		issue:   e,
	}
	return f
}

func (f Fault) MarshalJSON() ([]byte, error) {

	b := "{"

	b += fmt.Sprintf(`"message":"%s"`, f.message)
	if f.issue != nil {
		j, err := json.Marshal(f.issue)
		if err != nil {
			return nil, err
		}
		b += fmt.Sprintf(`, "error": { "title":"%s", "info":%s }`, f.issue.Error(), string(j))
	}

	b += "}"

	return []byte(b), nil
}
