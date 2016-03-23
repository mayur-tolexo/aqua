package aqua

import "fmt"

type Fault struct {
	HttpCode int    `json:"-"`
	Message  string `json:"message"`
	Issue    error  `json:"issue"`
}

func (f Fault) MarshalJSON() ([]byte, error) {

	b := "{"

	b += fmt.Sprintf(`"message":"%s"`, f.Message)
	if f.Issue != nil {
		b += fmt.Sprintf(`, "issue": "%s"`, f.Issue.Error())
	}

	b += "}"

	return []byte(b), nil
}

// Fault implements error interface
func (f Fault) Error() string {
	return f.Issue.Error()
}
