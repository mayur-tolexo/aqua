package aqua

import (
	"net/http"
	"runtime"
	"time"

	"github.com/pivotal-golang/bytefmt"
)

type CoreService struct {
	RestService `root:"/aqua/"`
	ping        GET `url:"/ping"`
	status      GET `url:"/status" pretty:"true"`
	date        GET `url:"/time"`
}

func (me *CoreService) Ping() string {
	return "pong"
}

func (me *CoreService) Status() map[string]interface{} {

	out := make(map[string]interface{})

	m := runtime.MemStats{}
	runtime.ReadMemStats(&m)

	mem := make(map[string]interface{})

	mem_gen := make(map[string]interface{})
	mem_gen["alloc"] = bytefmt.ByteSize(m.Alloc)
	mem_gen["total_alloc"] = bytefmt.ByteSize(m.TotalAlloc)
	mem["general"] = mem_gen

	mem_hp := make(map[string]interface{})
	mem_hp["alloc"] = bytefmt.ByteSize(m.HeapAlloc)
	mem_hp["sys"] = bytefmt.ByteSize(m.HeapAlloc)
	mem_hp["idle"] = bytefmt.ByteSize(m.HeapIdle)
	mem_hp["inuse"] = bytefmt.ByteSize(m.HeapInuse)
	mem_hp["released"] = bytefmt.ByteSize(m.HeapReleased)
	mem_hp["objects"] = bytefmt.ByteSize(m.HeapObjects)
	mem["heap"] = mem_hp

	out["mem"] = mem
	out["server-time"] = time.Now().Format("2006-01-02 15:04:05 MST")
	out["go-version"] = runtime.Version()[2:]
	out["aqua-version"] = release

	return out
}

func (me *CoreService) Date(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(time.Now().Format("2006-01-02 15:04:05 MST")))
}
