package aqua

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func cleanUrl(pieces ...string) string {

	var buffer bytes.Buffer

	// init the buffer to be a relative url
	buffer.WriteString("/")

	for _, p := range pieces {
		if p != "" && p != "-" {
			buffer.WriteString("/")
			buffer.WriteString(p)
		}
	}

	url := removeMultSlashes(buffer.String())
	//url = dropPrefix(url, "/")

	return url
}

func dropPrefix(s string, prefix string) string {
	if strings.HasPrefix(s, prefix) {
		return s[len(prefix):]
	}
	return s
}

func getServiceId(method string, prefix string, version string, url string) string {
	if version != defaults.Version {
		version = "v" + version
	}
	return removeMultSlashes(fmt.Sprintf("%s/%s/%s%s", method, prefix, version, url))
}

var find *regexp.Regexp

func removeMultSlashes(inp string) string {
	if find == nil {
		find, _ = regexp.Compile("[\\/]+")
	}

	return find.ReplaceAllString(inp, "/")
}

func getUrl(url string, headers map[string]string) (httpCode int, contentType string, content string) {
	req, _ := http.NewRequest("GET", url, nil)
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return resp.StatusCode, resp.Header.Get("Content-Type"), string(data)
}

func postUrl(uri string, post map[string]string, headers map[string]string) (httpCode int, contentType string, content string) {
	form := url.Values{}
	for key, val := range post {
		form.Set(key, val)
	}
	req, err := http.NewRequest("POST", uri, strings.NewReader(form.Encode()))
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return resp.StatusCode, resp.Header.Get("Content-Type"), string(data)
}

var portForTesting int = 8095

func getUniquePortForTestCase() int {
	portForTesting++
	return portForTesting
}

func getHttpMethod(field reflect.StructField) string {
	var out string = ""
	switch field.Type.String() {
	case "aqua.GetApi", "aqua.PostApi", "aqua.PutApi", "aqua.PatchApi", "aqua.DeleteApi", "aqua.CrudApi":
		out = field.Type.String()
		out = out[5 : len(out)-3]
		out = strings.ToUpper(out)
	case "aqua.GET", "aqua.POST", "aqua.PUT", "aqua.PATCH", "aqua.DELETE", "aqua.CRUD":
		out = field.Type.String()
		out = out[5:]
		out = strings.ToUpper(out)
	}

	return out
}

var muxStyle *regexp.Regexp

func extractRouteVars(url string) []string {

	if muxStyle == nil {
		muxStyle, _ = regexp.Compile(`{[^/]+}`)
	}

	matches := muxStyle.FindAllString(url, -1)
	var colonPos int
	for i, m := range matches {
		m = m[1 : len(m)-1] // drop { and }
		colonPos = strings.Index(m, ":")
		if colonPos > 0 {
			m = m[0:colonPos]
		}
		matches[i] = m
	}

	return matches
}

func convertToType(vars []string, typ []string) []reflect.Value {
	vals := make([]reflect.Value, len(vars))
	for i, v := range vars {
		t := typ[i]
		switch t {
		case "string":
			vals[i] = reflect.ValueOf(v)
		case "int":
			j, err := strconv.Atoi(v)
			if err != nil {
				panic(fmt.Sprintf("Cannot convert [%s] to 'int'", v))
			}
			vals[i] = reflect.ValueOf(j)
		case "uint":
			j, err := strconv.ParseUint(v, 10, 32)
			if err != nil {
				panic(fmt.Sprintf("Cannot convert [%s] to 'uint'", v))
			}
			vals[i] = reflect.ValueOf(uint(j))
		default:
			panic(fmt.Sprintf("Type [%s] is not supported", t))
		}
	}
	return vals
}

func isError(e interface{}) bool {
	_, ok := e.(error)
	return ok
}
