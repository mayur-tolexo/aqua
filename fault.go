package aqua

import (
	"fmt"
	"strconv"
)

type Fault struct {
	HttpCode int    `json:"-"`
	Message  string `json:"message"`
	Issue    error  `json:"issue"`
}

func (f Fault) MarshalJSON() ([]byte, error) {

	b := fmt.Sprintf(`{"message":%s`, strconv.Quote(f.Message))
	if f.Issue != nil {
		b += fmt.Sprintf(`, "issue": %s`, strconv.Quote(f.Issue.Error()))
	}
	b += "}"

	return []byte(b), nil
}

// Fault implements error interface
func (f Fault) Error() string {
	return f.Issue.Error()
}
