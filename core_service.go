package aqua

import (
	"github.com/pivotal-golang/bytefmt"
	"net/http"
	"runtime"
	"time"
)

type CoreService struct {
	RestService `root:"/aqua/"`
	ping   GET `url:"/ping"`
	status GET `url:"/status" pretty:"true"`
	date   GET `url:"/time"`
}

func (me *CoreService) Ping() string {
	return "pong"
}

func (me *CoreService) Status() *Sac {

	out := NewSac()

	m := runtime.MemStats{}
	runtime.ReadMemStats(&m)

	mem := NewSac()

	mem_gen := NewSac().
		Set("alloc", bytefmt.ByteSize(m.Alloc)).
		Set("total_alloc", bytefmt.ByteSize(m.TotalAlloc))
	mem.Set("general", mem_gen)

	mem_hp := NewSac().
		Set("alloc", bytefmt.ByteSize(m.HeapAlloc)).
		Set("sys", bytefmt.ByteSize(m.HeapAlloc)).
		Set("idle", bytefmt.ByteSize(m.HeapIdle)).
		Set("inuse", bytefmt.ByteSize(m.HeapInuse)).
		Set("released", bytefmt.ByteSize(m.HeapReleased)).
		Set("objects", bytefmt.ByteSize(m.HeapObjects))
	mem.Set("heap", mem_hp)

	out.Set("mem", mem).
		Set("server-time", time.Now().Format("2006-01-02 15:04:05 MST")).
		Set("go-version", runtime.Version()[2:]).
		Set("aqua-version", release)

	return out
}

func (me *CoreService) Date(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(time.Now().Format("2006-01-02 15:04:05 MST")))
}
