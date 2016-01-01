package aqua

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/carbocation/interpose"
	"github.com/gorilla/mux"
	"github.com/thejackrabbit/aero/cache"
	"github.com/thejackrabbit/aero/panik"
	"net/http"
	"reflect"
	"strings"
	"time"
)

type endPoint struct {
	exec       Invoker
	config     Fixture
	httpMethod string

	stdHandler bool
	jarInput   bool

	urlWithVersion string
	urlWoVersion   string
	muxVars        []string
	modules        []func(http.Handler) http.Handler
	stash          cache.Cacher

	svcUrl string
	svcId  string
}

func NewEndPoint(inv Invoker, f Fixture, httpMethod string, mods map[string]func(http.Handler) http.Handler,
	caches map[string]cache.Cacher) endPoint {

	out := endPoint{
		exec:           inv,
		config:         f,
		stdHandler:     false,
		jarInput:       false,
		urlWithVersion: cleanUrl(f.Prefix, "v"+f.Version, f.Root, f.Url),
		urlWoVersion:   cleanUrl(f.Prefix, f.Root, f.Url),
		muxVars:        extractRouteVars(f.Url),
		httpMethod:     httpMethod,
		modules:        make([]func(http.Handler) http.Handler, 0),
		stash:          nil,
	}

	// Perform all validations, unless it is a mock stub
	if f.Stub == "" {
		out.stdHandler = out.signatureMatchesDefaultHttpHandler()
		out.jarInput = out.needsJarInput()

		out.validateMuxVarsMatchFuncInputs()
		out.validateFuncInputsAreOfRightType()
		out.validateFuncOutputsAreCorrect()
	}

	// Filter out relevant modules for later use
	if mods != nil && f.Modules != "" {
		names := strings.Split(f.Modules, ",")
		for _, name := range names {
			name = strings.TrimSpace(name)
			fn, found := mods[name]
			if !found {
				panic(fmt.Sprintf("Module:%s not found", name))
			}
			out.modules = append(out.modules, fn)
		}
	}

	// Figure out which cache store to use, unless it is a mock stub
	if f.Stub == "" {
		if c, ok := caches[f.Cache]; ok {
			out.stash = c
		} else if f.Cache != "" {
			if out.config.Version == "" {
				panik.Do("Cache provider %s is missing for %s", f.Cache, out.urlWoVersion)
			} else {
				panik.Do("Cache provider %s is missing for %s", f.Cache, out.urlWithVersion)
			}

		}
	}

	return out
}

func (me *endPoint) signatureMatchesDefaultHttpHandler() bool {
	return me.exec.outCount == 0 &&
		me.exec.inpCount == 2 &&
		me.exec.inpParams[0] == "i:net/http.ResponseWriter" &&
		me.exec.inpParams[1] == "*st:net/http.Request"
}

func (me *endPoint) needsJarInput() bool {
	// needs jar input as the last parameter
	for i := 0; i < len(me.exec.inpParams)-1; i++ {
		if me.exec.inpParams[i] == "st:github.com/thejackrabbit/aqua.Jar" {
			panic("Jar parameter should be the last one: " + me.exec.name)
		}
	}
	return me.exec.inpCount > 0 && me.exec.inpParams[me.exec.inpCount-1] == "st:github.com/thejackrabbit/aqua.Jar"
}

func (me *endPoint) validateMuxVarsMatchFuncInputs() {
	// for non-standard http handlers, the mux vars count should match
	// the count of inputs to the user's method
	if !me.stdHandler {
		inputs := me.exec.inpCount
		if me.jarInput {
			inputs += -1
		}
		if me.httpMethod == "CRUD" {
			panik.If(inputs != 0, "Crud methods should not take any inputs %s", me.exec.name)
		} else if len(me.muxVars) != inputs {
			panic(fmt.Sprintf("%s has %d inputs, but the func (%s) has %d",
				me.urlWithVersion, len(me.muxVars), me.exec.name, inputs))
		}
	}
}

func (me *endPoint) validateFuncInputsAreOfRightType() {
	if !me.stdHandler {
		for _, s := range me.exec.inpParams {
			switch s {
			case "st:github.com/thejackrabbit/aqua.Jar":
			case "int":
			case "string":
			default:
				panic("Func input params should be 'int' or 'string'. Observed: " + s + " in " + me.exec.name)
			}
		}
	}
}

func (me *endPoint) validateFuncOutputsAreCorrect() {

	var accepts = make(map[string]bool)
	accepts["string"] = true
	accepts["map"] = true
	accepts["st:github.com/thejackrabbit/aqua.Sac"] = true
	accepts["*st:github.com/thejackrabbit/aqua.Sac"] = true
	accepts["i:."] = true

	if me.httpMethod == "CRUD" {
		panik.If(me.exec.outCount != 1, "CrudApi must return 1 param only")
		panik.If(me.exec.outParams[0] != "st:github.com/thejackrabbit/aqua.CrudApi", "CrudApi return must be of type CrudApi")
	} else if !me.stdHandler {
		switch me.exec.outCount {
		case 1:
			_, found := accepts[me.exec.outParams[0]]
			correctPrefix := strings.HasPrefix(me.exec.outParams[0], "st:") || strings.HasPrefix(me.exec.outParams[0], "sl:")
			if !found && !correctPrefix {
				panic("Incorrect return type found in: " + me.exec.name + " - " + me.exec.outParams[0])
			}
		case 2:
			if me.exec.outParams[0] != "int" {
				panic("When a func returns two params, the first must be an int (http status code) : " + me.exec.name)
			}
			_, found := accepts[me.exec.outParams[1]]
			correctPrefix := strings.HasPrefix(me.exec.outParams[1], "st:") || strings.HasPrefix(me.exec.outParams[1], "sl:")
			if !found && !correctPrefix {
				panic("Incorrect return type for second return param found in: " + me.exec.name + " - " + me.exec.outParams[1])
			}
		default:
			panik.Do("Incorrect number of returns for Func: %s", me.exec.name)
		}
	}
}

func (me *endPoint) setupMuxHandlers(mux *mux.Router) (svcUrl string) {

	m := interpose.New()
	for i, _ := range me.modules {
		m.Use(me.modules[i])
		//fmt.Println("using module:", me.modules[i], reflect.TypeOf(me.modules[i]))
	}
	fn := handleIncoming(me)
	m.UseHandler(http.HandlerFunc(fn))

	if me.config.Version == "" {
		// url without version
		svcUrl = me.urlWoVersion
		mux.Handle(me.urlWoVersion, m).Methods(me.httpMethod)
		fmt.Printf("%s:%s\r\n", me.httpMethod, me.urlWoVersion)

		// TODO: should we add content type application+json here?
	} else {
		// versioned url
		svcUrl = me.urlWithVersion
		mux.Handle(me.urlWithVersion, m).Methods(me.httpMethod)
		fmt.Printf("%s:%s\r\n", me.httpMethod, me.urlWithVersion)

		// content type (style1)
		header1 := fmt.Sprintf("application/%s-v%s+json", me.config.Vendor, me.config.Version)
		mux.Handle(me.urlWoVersion, m).Methods(me.httpMethod).Headers("Accept", header1)

		// content type (style2)
		header2 := fmt.Sprintf("application/%s+json;version=%s", me.config.Vendor, me.config.Version)
		mux.Handle(me.urlWoVersion, m).Methods(me.httpMethod).Headers("Accept", header2)
	}

	me.svcUrl = svcUrl
	me.svcId = fmt.Sprintf("%s:%s", me.httpMethod, svcUrl)

	return me.svcId
}

func handleIncoming(e *endPoint) func(http.ResponseWriter, *http.Request) {

	// return stub
	if e.config.Stub != "" {
		return func(w http.ResponseWriter, r *http.Request) {
			d, err := getContent(e.config.Stub)
			if err == nil {
				fmt.Fprintf(w, "%s", d)
			} else {
				w.WriteHeader(400)
				fmt.Fprintf(w, "{ message: \"%s\"}", "Error reading stub content "+e.config.Stub)
			}
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {

		// TODO: create less local variables
		// TODO: move vars to closure level

		var out []reflect.Value

		var useCache bool = false
		var ttl time.Duration = 0 * time.Second
		var val []byte
		var err error

		if e.config.Ttl != "" {
			ttl, err = time.ParseDuration(e.config.Ttl)
			panik.On(err)
		}
		useCache = r.Method == "GET" && ttl > 0 && e.stash != nil

		muxVals := mux.Vars(r)
		params := make([]string, len(e.muxVars))
		for i, k := range e.muxVars {
			params[i] = muxVals[k]
		}

		if e.stdHandler {
			//TODO: caching of standard handler
			e.exec.Do([]reflect.Value{reflect.ValueOf(w), reflect.ValueOf(r)})
		} else {
			ref := convertToType(params, e.exec.inpParams)
			if e.jarInput {
				ref = append(ref, reflect.ValueOf(NewJar(r)))
			}

			if useCache {
				val, err = e.stash.Get(r.RequestURI)
				if err == nil {
					out = decomposeCachedValues(val, e.exec.outParams)
				} else {
					out = e.exec.Do(ref)
					if len(out) == 2 && e.exec.outParams[0] == "int" {
						code := out[0].Int()
						if code < 200 || code > 299 {
							useCache = false
						}
					}
					if useCache {
						bytes := prepareForCaching(out, e.exec.outParams)
						e.stash.Set(r.RequestURI, bytes, ttl)
					}
				}
			} else {
				out = e.exec.Do(ref)
			}
			writeOutput(w, r, e.exec.outParams, out, e.config.Pretty)
		}
	}
}

func prepareForCaching(r []reflect.Value, outputParams []string) []byte {

	var err error
	buf := new(bytes.Buffer)
	encd := json.NewEncoder(buf)

	for i, _ := range r {
		switch outputParams[i] {
		case "int":
			err = encd.Encode(r[i].Int())
			panik.On(err)
		case "map":
			err = encd.Encode(r[i].Interface().(map[string]interface{}))
			panik.On(err)
		case "string":
			err = encd.Encode(r[i].String())
			panik.On(err)
		case "*st:github.com/thejackrabbit/aqua.Sac":
			err = encd.Encode(r[i].Elem().Interface().(Sac).Data)
		default:
			panic("Unknown type of output to be sent to endpoint cache: " + outputParams[i])
		}
	}

	return buf.Bytes()
}

func decomposeCachedValues(data []byte, outputParams []string) []reflect.Value {

	var err error
	buf := bytes.NewBuffer(data)
	decd := json.NewDecoder(buf)
	out := make([]reflect.Value, len(outputParams))

	for i, o := range outputParams {
		switch o {
		case "int":
			var j int
			err = decd.Decode(&j)
			panik.On(err)
			out[i] = reflect.ValueOf(j)
		case "map":
			var m map[string]interface{}
			err = decd.Decode(&m)
			panik.On(err)
			out[i] = reflect.ValueOf(m)
		case "string":
			var s string
			err = decd.Decode(&s)
			panik.On(err)
			out[i] = reflect.ValueOf(s)
		case "*st:github.com/thejackrabbit/aqua.Sac":
			var m map[string]interface{}
			err = decd.Decode(&m)
			panik.On(err)
			s := NewSac()
			s.Data = m
			out[i] = reflect.ValueOf(s)
		default:
			panic("Unknown type of output to be decoded from endpoint cache:" + o)
		}
	}

	return out

}
