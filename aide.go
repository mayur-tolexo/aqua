package aqua

import (
	"io/ioutil"
	"net/http"
	"strings"
)

var separator string = ","

// Aide allows access to Request, Response and other vars (post, get, body)
type Aide struct {
	// request handle
	Request  *http.Request
	Response http.ResponseWriter

	// variables
	PostVar  map[string]string
	QueryVar map[string]string
	Body     string
}

// NewAide creates a new Aide object
func NewAide(w http.ResponseWriter, r *http.Request) Aide {
	return Aide{
		Request:  r,
		Response: w,
	}
}

// LoadVars parses an intializes PostVars, GetVars and Body variables
func (j *Aide) LoadVars() {

	if j.PostVar != nil {
		panic("Aide.LoadVars can be called only once per request")
	} else {
		j.PostVar = make(map[string]string)
		j.QueryVar = make(map[string]string)
	}

	if j.Request.Method == "POST" || j.Request.Method == "PUT" {
		ctype := j.Request.Header.Get("Content-Type")
		switch {
		case ctype == "application/x-www-form-urlencoded":
			j.Request.ParseForm()
			j.loadPostVar(j.Request)
			j.loadQueryVar(j.Request, true)
		case strings.HasPrefix(ctype, "multipart/form-data;"):
			// ParseMultiPart form should ideally populate
			// r.PostForm, but instead it fills r.Form
			// https://github.com/golang/go/issues/9305
			j.Request.ParseMultipartForm(1024 * 1024)
			j.loadPostVar(j.Request)
			j.loadQueryVar(j.Request, true)
		default:
			j.Body = getBody(j.Request)
		}
	} else if j.Request.Method == "GET" {
		j.Request.ParseForm()
		j.loadQueryVar(j.Request, false)
	}
}

func getBody(r *http.Request) string {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()
	return string(b)
}

func (j *Aide) loadPostVar(r *http.Request) {
	for k := range r.PostForm {
		j.PostVar[k] = strings.Join(r.PostForm[k], separator)
	}
}

func (j *Aide) loadQueryVar(r *http.Request, skipPostVars bool) {
	for k := range r.Form {
		if skipPostVars {
			// only add to query-vars if it is NOT a post var
			if _, found := j.PostVar[k]; !found {
				j.QueryVar[k] = strings.Join(r.Form[k], separator)
			}
		} else {
			j.QueryVar[k] = strings.Join(r.Form[k], separator)
		}
	}

}
