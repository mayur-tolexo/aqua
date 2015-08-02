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
	// TODO: check if the same key alread exists
	// TODO: AddModule must be called before AddService
	me.mods[name] = f
}

func (me *RestServer) AddCache(name string, c cache.Cacher) {
	// TODO: check if the same key already exists
	// TODO: AddCache must be called before AddService
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
		fmt.Println("Loading...")
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
			// Skip non http method fields
			continue
		}

		v := ""
		if fix.Version != "" {
			v = fix.Version
		}
		serviceId := cleanUrl(method, fix.Prefix, v, fix.Root, fix.Url)

		if _, found := me.apis[serviceId]; found {
			panik.Do("Cannot load same service again %s in %s", serviceId, svcType.String())
		}

		exec := NewMethodInvoker(svc, upFirstChar(field.Name))
		if exec.exists || fix.Stub != "" {
			ep := NewEndPoint(exec, fix, method, me.mods, me.stores)
			ep.setupMuxHandlers(me.mux)
			me.apis[serviceId] = ep
			fmt.Printf("%s\n", serviceId)
			//fmt.Println(fix)
		}
	}
}

func (me *RestServer) validateService(svc interface{}) {
	svcType := reflect.TypeOf(svc)
	code := getSymbolFromType(svcType)

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
