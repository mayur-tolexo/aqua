package aqua

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/thejackrabbit/aero/db/cstr"
	"github.com/thejackrabbit/aero/db/orm"
	"github.com/thejackrabbit/aero/ds"
	"github.com/thejackrabbit/aero/engine"
	"github.com/thejackrabbit/aero/panik"
	"github.com/thejackrabbit/aero/refl"
)

type Api struct{ Fixture }

type GET struct{ Api }
type POST struct{ Api }
type PUT struct{ Api }
type PATCH struct{ Api }
type DELETE struct{ Api }

type CRUD struct {
	Api
	cstr.Storage
	Model func() (interface{}, interface{})
}

// If DB infomraiton was not set by user, then try to use the master
func (c *CRUD) useMasterIfMissing() {
	if c.Engine == "" && c.Conn == "" {
		c.Storage = cstr.Get(true)
	}
}

func (c *CRUD) validate() {
	panik.If(c.Engine == "", "Crud storage engine not specified")
	panik.If(c.Conn == "", "Crud storage conn not spefieid")

	if c.getMethod("create") == "Rdbms_Create" { // Model is a must
		panik.If(c.Model == nil, "Model not specified")

		m, arr := c.Model()
		panik.If(m == nil, "Model method returns nil")
		panik.If(!strings.HasPrefix(refl.ObjSignature(m), "*st:"), "Model() method param 1 must be address of a gorm struct")
		if arr != nil {
			panik.If(!strings.HasPrefix(refl.ObjSignature(arr), "*sl:"), "Model() method param 2 must be address of a slice of gorm struct")
		}
	}
}

func (c *CRUD) getMethod(action string) string {

	switch c.Engine {
	case "mysql", "maria", "mariadb", "postgres", "sqlite3":
		switch action {
		case "create":
			return "Rdbms_Create"
		case "read":
			return "Rdbms_Read"
		case "update":
			return "Rdbms_Update"
		case "delete":
			return "Rdbms_Delete"
		case "sql":
			return "Rdbms_FetchSql"
		case "sqlJson":
			return "Rdbms_FetchSqlJson"
		}

	case "memcache":
		switch action {
		case "read":
			return "Memcache_Read"
		case "update":
			return "Memcache_Update"
		case "delete":
			return "Memcache_Delete"
		}
	}

	return ""
}

func (c *CRUD) Rdbms_Read(primKey string) interface{} {
	m, _ := c.Model()

	dbo := orm.GetConn(c.Engine, c.Conn)

	if err := dbo.Debug().First(m, primKey).Error; err != nil {
		return err
	}
	return m
}

func (c *CRUD) Rdbms_Create(j Aide) interface{} {
	j.LoadVars()

	m, _ := c.Model()
	//err := ds.LoadStruct(m, []byte(j.Body))
	err := ds.Load(m, []byte(j.Body))
	if err != nil {
		return err
	}

	dbo := orm.GetConn(c.Engine, c.Conn)

	stmt := dbo.Debug().Create(m)

	if stmt.Error != nil {
		return stmt.Error
	}

	return map[string]interface{}{"rows_affected": stmt.RowsAffected, "success": 1}
}

func (c *CRUD) Rdbms_Delete(primKey string) interface{} {
	m, _ := c.Model()

	dbo := orm.GetConn(c.Engine, c.Conn)

	if err := dbo.Debug().Where(primKey).Delete(m).Error; err != nil {
		return err
	}

	return map[string]interface{}{"success": 1}
}

func (c *CRUD) Rdbms_Update(primKey string, j Aide) interface{} {
	j.LoadVars()

	var data map[string]interface{}
	err := json.Unmarshal([]byte(j.Body), &data)
	if err != nil {
		return err
	}

	dbo := orm.GetConn(c.Engine, c.Conn)

	m, _ := c.Model()

	if err := dbo.Debug().Model(m).Where(primKey).UpdateColumns(data).Error; err != nil {
		return err
	}

	return map[string]interface{}{"success": 1}
}

func (c *CRUD) Rdbms_FetchSql(j Aide) interface{} {
	j.LoadVars()
	m, col := c.Model()

	dbo := orm.GetConn(c.Engine, c.Conn)

	if err := dbo.Debug().Model(m).Where(j.Body).Find(col).Error; err != nil {
		return err
	}
	return col
}

func (c *CRUD) Rdbms_FetchSqlJson(j Aide) interface{} {
	j.LoadVars()

	var data map[string]interface{}
	err := json.Unmarshal([]byte(j.Body), &data)
	if err != nil {
		return err
	}

	whr := ""
	where, ok := data["where"]
	if ok {
		w, ok := where.(string)
		if ok {
			whr = w
		}
	}

	p := make([]interface{}, 0)
	params, ok := data["params"]
	if ok {
		p, ok = params.([]interface{})
		if !ok {
			return errors.New("params must be an array")
		}
	}

	// if limit is not specified then set it to ""
	lim := 100
	limit, ok := data["limit"]
	if ok {
		f, ok := limit.(int)
		if !ok {
			return errors.New("limit must be integer")
		} else {
			lim = f
		}
	}

	// if offset is not specified then set it to ""
	off := 1
	offset, ok := data["offset"]
	if ok {
		f, ok := offset.(int)
		if !ok {
			return errors.New("offset must be integer")
		} else {
			off = f
		}
	}

	// if order by is string or array (of string), then use it
	ord := ""
	order, ok := data["order"]
	if ok {
		s, ok := order.(string)
		if ok {
			ord = s
		} else if sl, ok := order.([]interface{}); ok {
			for _, v := range sl {
				t, ok := v.(string)
				if !ok {
					return errors.New("order must be string or array of string")
				}
				if ord == "" {
					ord = t
				} else {
					ord += "," + t
				}
			}
		} else {
			return errors.New("order must be string or array of string")
		}
	}

	m, col := c.Model()
	dbo := orm.GetConn(c.Engine, c.Conn)

	if err := dbo.Debug().Model(m).
		Where(whr, p...).
		Order(ord).
		Limit(lim).
		Offset(off).
		Find(col).Error; err != nil {
		return err
	}

	return col
}

func (c *CRUD) Memcache_Read(primKey string) interface{} {

	// Memcache object
	spl := strings.Split(c.Conn, ":")
	host := spl[0]
	port, err := strconv.Atoi(spl[1])
	panik.On(err)
	memc := engine.NewMemcache(host, port)
	defer memc.Close()

	data, err := memc.Get(primKey)

	if err == nil {
		return string(data)
	} else {
		return err
	}
}

func (c *CRUD) Memcache_Update(primKey string, j Aide) interface{} {

	// Memcache object
	spl := strings.Split(c.Conn, ":")
	host := spl[0]
	port, err := strconv.Atoi(spl[1])
	panik.On(err)
	memc := engine.NewMemcache(host, port)
	defer memc.Close()

	ttl, err := time.ParseDuration(c.Ttl)
	panik.On(err)
	panik.If(ttl == 0, "ttl cache duration should not be 0")

	j.LoadVars()
	memc.Set(primKey, []byte(j.Body), ttl)

	return ""
}

func (c *CRUD) Memcache_Delete(primKey string, j Aide) interface{} {

	// Memcache object
	spl := strings.Split(c.Conn, ":")
	host := spl[0]
	port, err := strconv.Atoi(spl[1])
	panik.On(err)
	memc := engine.NewMemcache(host, port)
	defer memc.Close()

	err = memc.Delete(primKey)

	if err == nil {
		return ""
	} else {
		return err
	}
}

// TODO: write test cases for CRUD and fetch methods
