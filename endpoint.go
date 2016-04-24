package aqua

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/carbocation/interpose"
	"github.com/gorilla/mux"
	"github.com/rightjoin/aero/cache"
	"github.com/rightjoin/aero/panik"
)

type endPoint struct {
	exec       Invoker
	config     Fixture
	httpMethod string

	stdHandler bool
	needsAide  bool

	urlWithVersion string
	urlWoVersion   string
	muxVars        []string
	modules        []func(http.Handler) http.Handler
	stash          cache.Cacher
	auth           Authorizer

	svcUrl string
	svcId  string
}

func NewEndPoint(inv Invoker, f Fixture, httpMethod string, mods map[string]func(http.Handler) http.Handler,
	caches map[string]cache.Cacher, a Authorizer) endPoint {

	out := endPoint{
		exec:           inv,
		config:         f,
		stdHandler:     false,
		needsAide:      false,
		urlWithVersion: cleanUrl(f.Prefix, "v"+f.Version, f.Root, f.Url),
		urlWoVersion:   cleanUrl(f.Prefix, f.Root, f.Url),
		muxVars:        extractRouteVars(f.Url),
		httpMethod:     httpMethod,
		modules:        make([]func(http.Handler) http.Handler, 0),
		stash:          nil,
		auth:           a,
	}

	// Perform all validations, unless it is a mock stub
	if f.Stub == "" {
		out.stdHandler = out.signatureMatchesDefaultHttpHandler()
		out.needsAide = out.needsAideInput()

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

func (me *endPoint) needsAideInput() bool {
	// aide the last parameter?
	for i := 0; i < len(me.exec.inpParams)-1; i++ {
		if me.exec.inpParams[i] == "st:github.com/rightjoin/aqua.Aide" {
			panic("Aide parameter should be the last one: " + me.exec.name)
		}
	}
	return me.exec.inpCount > 0 && me.exec.inpParams[me.exec.inpCount-1] == "st:github.com/rightjoin/aqua.Aide"
}

func (me *endPoint) validateMuxVarsMatchFuncInputs() {
	// for non-standard http handlers, the mux vars count should match
	// the count of inputs to the user's method
	if !me.stdHandler {
		inputs := me.exec.inpCount
		if me.needsAide {
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
			case "st:github.com/rightjoin/aqua.Aide":
			case "int":
			case "uint":
			case "string":
			default:
				panic("Func input params should be 'int' or 'string'. Observed: " + s + " in " + me.exec.name)
			}
		}
	}
}

func (me *endPoint) validateFuncOutputsAreCorrect() {

	if me.httpMethod == "CRUD" {
		panik.If(me.exec.outCount != 1, "CrudApi must return 1 param only")
		panik.If(me.exec.outParams[0] != "st:github.com/rightjoin/aqua.CRUD", "CRUD return must be of type CRUD")
	} else if !me.stdHandler {
		switch me.exec.outCount {
		case 1:
			if !me.isAcceptableType(me.exec.outParams[0]) {
				panic("Incorrect return type found in: " + me.exec.name + " - " + me.exec.outParams[0])
			}
		case 2:
			if me.exec.outParams[0] == "int" {
				if !me.isAcceptableType(me.exec.outParams[1]) {
					panic("Two param func must have type <int> followed by an acceptable type. Found: " + me.exec.name)
				}
			} else if me.exec.outParams[1] == "i:.error" {
				if !me.isAcceptableType(me.exec.outParams[0]) {
					panic("Two param func must have type acceptable type followed by an error." + me.exec.name + "Found:" + me.exec.outParams[0])
				}
			} else {
				panic("Two param func must have type int,<something> or <something>,error." + me.exec.name + "Found:" + me.exec.outParams[0] + "," + me.exec.outParams[1])
			}
		default:
			panik.Do("Incorrect number of returns for Func: %s", me.exec.name)
		}
	}
}

func (me *endPoint) isAcceptableType(dataType string) bool {
	var accepts = make(map[string]bool)
	accepts["string"] = true
	accepts["map"] = true
	accepts["st:github.com/rightjoin/aqua.Sac"] = true
	accepts["*st:github.com/rightjoin/aqua.Sac"] = true
	accepts["i:."] = true

	_, found := accepts[dataType]
	if found {
		return true
	}

	if strings.HasPrefix(dataType, "st:") || strings.HasPrefix(dataType, "sl:") || strings.HasPrefix(dataType, "*st:") {
		return true
	}

	// no acceptable data type match
	return false
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

		// Authorization
		if e.auth != nil {
			if !e.auth.Authorize(r, e.config.Allow, e.config.Deny) {
				w.WriteHeader(401)
				w.Write([]byte(`{"message":"Unauthorized Access"}`))
				return
			}
		}

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
			if e.needsAide {
				ref = append(ref, reflect.ValueOf(NewAide(w, r)))
			}

			if useCache {
				val, err = e.stash.Get(r.RequestURI)
				if err == nil {
					out = decode(val, e.exec.outParams)
				} else {
					out = e.exec.Do(ref)
					if len(out) == 2 && e.exec.outParams[0] == "int" {
						code := out[0].Int()
						if code < 200 || code > 299 {
							useCache = false
						}
					}
					if useCache {
						bytes := encode(out, e.exec.outParams)
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
