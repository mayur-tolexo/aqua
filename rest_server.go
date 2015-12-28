package aqua

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/thejackrabbit/aero/cache"
	"github.com/thejackrabbit/aero/panik"
	"net/http"
	"reflect"
	"strings"
	"time"
)

var defaults Fixture = Fixture{
	Root:    "",
	Url:     "",
	Version: "",
	Pretty:  "false",
	Vendor:  "vnd.api",
	Modules: "",
	Cache:   "",
	Ttl:     "",
}

var release string = "0.0.1"
var defaultPort int = 8090

type RestServer struct {
	Fixture
	http.Server
	Port   int
	mux    *mux.Router
	svcs   []interface{}
	apis   map[string]endPoint
	mods   map[string]func(http.Handler) http.Handler
	stores map[string]cache.Cacher
}

func NewRestServer() RestServer {
	r := RestServer{
		Fixture: defaults,
		Server:  http.Server{},
		Port:    defaultPort,
		svcs:    make([]interface{}, 0),
		mux:     mux.NewRouter(),
		apis:    make(map[string]endPoint),
		mods:    make(map[string]func(http.Handler) http.Handler),
		stores:  make(map[string]cache.Cacher),
	}
	r.AddService(&CoreService{})
	return r
}

var printed bool = false

func (me *RestServer) AddModule(name string, f func(http.Handler) http.Handler) {
	me.mods[name] = f
}

func (me *RestServer) AddCache(name string, c cache.Cacher) {
	me.stores[name] = c
}

func (me *RestServer) AddService(svc interface{}) {
	me.svcs = append(me.svcs, svc)
}

func (me *RestServer) loadAllEndpoints() {
	for _, i := range me.svcs {
		me.loadServiceEndpoints(i)
	}
}

func (me *RestServer) loadServiceEndpoints(svc interface{}) {

	// Validations
	me.validateService(svc)

	if !printed {
		fmt.Println("Loading endpoints...")
		printed = true
	}

	fixSvcTag := NewFixtureFromTag(svc, "RestService")

	obj := reflect.ValueOf(svc).Elem()
	fixSvcObj := obj.FieldByName("RestService").FieldByName("Fixture").Interface().(Fixture)

	var fix Fixture
	var method string

	svcType := reflect.TypeOf(svc)
	objType := svcType.Elem()

	for i := 0; i < objType.NumField(); i++ {
		field := objType.FieldByIndex([]int{i})
		fixField := NewFixtureFromTag(svc, field.Name)
		fix = resolveInOrder(fixField, fixSvcObj, fixSvcTag, me.Fixture)

		// If root or url are missing then set it
		if fix.Root == "" {
			tmp := objType.Name()
			if strings.HasSuffix(tmp, "Service") {
				tmp = tmp[0 : len(tmp)-len("Service")]
			}
			fix.Root = toUrlCase(tmp)
		}
		if fix.Url == "" {
			fix.Url = toUrlCase(field.Name)
		}

		method = getHttpMethod(field)
		if method == "" {
			// Skip fields with unsupported types
			continue
		}

		if method == "CRUD" {

			// Validate crud method parameters (inputs and outputs)
			NewEndPoint(NewMethodInvoker(svc, upFirstChar(field.Name)), fix, "CRUD", me.mods, me.stores)

			// Validate crud struct fields
			vals := reflect.ValueOf(svc).MethodByName(upFirstChar(field.Name)).Call([]reflect.Value{})
			crud := vals[0].Interface().(CrudApi)

			crud.useMasterIfMissing()
			crud.validate()
			crud.Fixture = fix

			var exec Invoker
			var f Fixture

			// Setup GET endpoint and handler (for Reads)
			{
				f = fix
				f.Url += "/{pkey}"
				meth := crud.getMethod("read")
				if meth != "" {
					exec = NewMethodInvoker(&crud, meth)
					ep := NewEndPoint(exec, f, "GET", me.mods, me.stores)
					svcUrl := ep.setupMuxHandlers(me.mux)
					svcId := fmt.Sprintf("%s:%s", "GET", svcUrl)
					if _, found := me.apis[svcId]; found {
						panik.Do("Multiple services found in: %s on same URL %s", svcType, svcId)
					} else {
						me.apis[svcId] = ep
					}

				}
			}

			// Setup Create (post) endpoint
			{
				f = fix
				meth := crud.getMethod("create")
				if meth != "" {
					exec = NewMethodInvoker(&crud, meth)
					ep := NewEndPoint(exec, f, "POST", me.mods, me.stores)
					svcUrl := ep.setupMuxHandlers(me.mux)
					svcId := fmt.Sprintf("%s:%s", "POST", svcUrl)
					if _, found := me.apis[svcId]; found {
						panik.Do("Multiple services found in: %s on same URL %s", svcType, svcId)
					} else {
						me.apis[svcId] = ep
					}
				}
			}

			// Setup DELETE (delete) endpoint
			{
				f = fix
				f.Url += "/{pkey}"
				meth := crud.getMethod("delete")
				if meth != "" {
					exec = NewMethodInvoker(&crud, meth)
					ep := NewEndPoint(exec, f, "DELETE", me.mods, me.stores)
					svcUrl := ep.setupMuxHandlers(me.mux)
					svcId := fmt.Sprintf("%s:%s", "DELETE", svcUrl)
					if _, found := me.apis[svcId]; found {
						panik.Do("Multiple services found in: %s on same URL %s", svcType, svcId)
					} else {
						me.apis[svcId] = ep
					}
				}
			}

			//Setup Update (put) endpoint
			{
				f = fix
				f.Url += "/{pkey}"
				meth := crud.getMethod("update")
				if meth != "" {
					exec = NewMethodInvoker(&crud, meth)
					ep := NewEndPoint(exec, f, "PUT", me.mods, me.stores)
					svcUrl := ep.setupMuxHandlers(me.mux)
					svcId := fmt.Sprintf("%s:%s", "PUT", svcUrl)
					if _, found := me.apis[svcId]; found {
						panik.Do("Multiple services found in: %s on same URL %s", svcType, svcId)
					} else {
						me.apis[svcId] = ep
					}
				}
			}

			// Setup additional POST handler for ad-hoc queries
			fn := crud.Model
			var col interface{}
			if fn != nil {
				_, col = crud.Model()
			}

			if col != nil {

				// POST endpoint /[]
				// SQL is found in Post body
				{
					f = fix
					f.Url += "/!"
					meth := crud.getMethod("sql")
					if meth != "" {
						exec = NewMethodInvoker(&crud, meth)
						ep := NewEndPoint(exec, f, "POST", me.mods, me.stores)
						svcUrl := ep.setupMuxHandlers(me.mux)
						svcId := fmt.Sprintf("%s:%s", "POST", svcUrl)
						if _, found := me.apis[svcId]; found {
							panik.Do("Multiple services found in: %s on same URL %s", svcType, svcId)
						} else {
							me.apis[svcId] = ep
						}
					}
				}

				// POST endpoint /$
				// SQL and params are found in Post body in json form
				{
					f = fix
					f.Url += "/$"
					meth := crud.getMethod("sqlJson")
					if meth != "" {
						exec = NewMethodInvoker(&crud, meth)
						ep := NewEndPoint(exec, f, "POST", me.mods, me.stores)
						svcUrl := ep.setupMuxHandlers(me.mux)
						svcId := fmt.Sprintf("%s:%s", "POST", svcUrl)
						if _, found := me.apis[svcId]; found {
							panik.Do("Multiple services found in: %s on same URL %s", svcType, svcId)
						} else {
							me.apis[svcId] = ep
						}
					}
				}
			}

		} else {

			exec := NewMethodInvoker(svc, upFirstChar(field.Name))
			if exec.exists || fix.Stub != "" {
				ep := NewEndPoint(exec, fix, method, me.mods, me.stores)
				svcUrl := ep.setupMuxHandlers(me.mux)
				svcId := fmt.Sprintf("%s:%s", method, svcUrl)
				if _, found := me.apis[svcId]; found {
					panik.Do("Multiple services found in: %s on same URL %s", svcType, svcId)
				} else {
					me.apis[svcId] = ep
				}
			}
		}
	}
}

func (me *RestServer) validateService(svc interface{}) {
	svcType := reflect.TypeOf(svc)
	code := getSignOfType(svcType)

	// validation: must be pointer
	if !strings.HasPrefix(code, "*st:") {
		panic("RestServer.AddService() expects address of your Service object")
	}

	// validation: RestService field must be present and be anonymous
	rs, ok := svcType.Elem().FieldByName("RestService")
	if !ok || !rs.Anonymous || !rs.Type.ConvertibleTo(reflect.TypeOf(RestService{})) {
		panic("RestServer.AddService() expects object that contains anonymous RestService field")
	}
}

func (me *RestServer) Run() {
	me.loadAllEndpoints()
	startup(me, me.Port)
}

func (me *RestServer) RunAsync() {
	me.loadAllEndpoints()
	go startup(me, me.Port)

	// TODO: don't sleep, check for the server to come up, and panic if
	// it doesn't even after 5 sec
	time.Sleep(time.Millisecond * 50)
}

// For backward compatibility
func (me *RestServer) RunWith(port int, sync bool) {
	me.Port = port
	if sync {
		me.Run()
	} else {
		me.RunAsync()
	}
}

func startup(r *RestServer, port int) {
	if port > 0 {
		r.Addr = fmt.Sprintf(":%d", port)
	} else if r.Server.Addr == "" {
		r.Addr = fmt.Sprintf(":%d", port)
	}
	r.Server.Handler = r.mux
	fmt.Println(r.ListenAndServe())
}
