package aqua

import (
	"encoding/json"
	"errors"
	"github.com/thejackrabbit/aero/db/cstr"
	"github.com/thejackrabbit/aero/db/orm"
	"github.com/thejackrabbit/aero/panik"
	"github.com/thejackrabbit/aero/strukt"
	"strconv"
	"strings"
)

type Api struct{ Fixture }

type GetApi struct{ Api }
type PostApi struct{ Api }
type PutApi struct{ Api }
type PatchApi struct{ Api }
type DeleteApi struct{ Api }

type CrudApi struct {
	Api
	cstr.Storage
	Model  func() interface{}
	Models func() interface{}
}

// If DB infomraiton was not set by user, then try to use the master
func (c *CrudApi) useMasterIfMissing() {
	if c.Storage.Engine == "" && c.Storage.Conn == "" {
		c.Storage = cstr.Get(true)
	}
}

func (c *CrudApi) validate() {
	panik.If(c.Storage.Engine == "", "Crud storage engine not specified")
	panik.If(c.Storage.Conn == "", "Crud storage conn not spefieid")
	panik.If(c.Model == nil, "Model not specified")
	panik.If(c.Model() == nil, "Model method returns nil")
	panik.If(!strings.HasPrefix(getSignOfObject(c.Model()), "*st:"), "Model() method must return address of a gorm struct")

	if c.Models != nil {
		panik.If(!strings.HasPrefix(getSignOfObject(c.Models()), "*sl:"), "Models() method must return address of a gorm struct")
	}
}

func (c *CrudApi) Crud_Read(primKey string) interface{} {
	m := c.Model()

	dbo := orm.From(c.Storage.Engine, c.Storage.Conn)

	if err := dbo.Debug().First(m, primKey).Error; err != nil {
		return err
	}
	return m
}

func (c *CrudApi) Crud_Create(j Jar) interface{} {
	j.LoadVars()

	m := c.Model()
	err := strukt.FromJson(m, j.Body)
	if err != nil {
		return err
	}

	dbo := orm.From(c.Engine, c.Conn)

	stmt := dbo.Debug().Create(m)

	if stmt.Error != nil {
		return stmt.Error
	}

	return map[string]interface{}{"rows_affected": stmt.RowsAffected, "success": 1}
}

func (c *CrudApi) Crud_Delete(primKey string) interface{} {
	m := c.Model()

	dbo := orm.From(c.Storage.Engine, c.Storage.Conn)

	if err := dbo.Debug().Where(primKey).Delete(m).Error; err != nil {
		return err
	}

	return map[string]interface{}{"success": 1}
}

func (c *CrudApi) Crud_Update(primKey string, j Jar) interface{} {
	j.LoadVars()

	var data map[string]interface{}
	err := json.Unmarshal([]byte(j.Body), &data)
	if err != nil {
		return err
	}

	dbo := orm.From(c.Storage.Engine, c.Storage.Conn)

	if err := dbo.Debug().Model(c.Model()).Where(primKey).UpdateColumns(data).Error; err != nil {
		return err
	}

	return map[string]interface{}{"success": 1}
}

func (c *CrudApi) Crud_FetchSql(j Jar) interface{} {
	j.LoadVars()
	a := c.Models()

	dbo := orm.From(c.Storage.Engine, c.Storage.Conn)

	if err := dbo.Debug().Model(c.Model()).Where(j.Body).Find(a).Error; err != nil {
		return err
	}
	return a
}

func (c *CrudApi) Crud_FetchSqlJson(j Jar) interface{} {
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
	lim := ""
	limit, ok := data["limit"]
	if ok {
		f, ok := limit.(float64)
		if !ok {
			return errors.New("limit must be integer")
		} else {
			lim = strconv.Itoa(int(f))
		}
	}

	// if offset is not specified then set it to ""
	off := ""
	offset, ok := data["offset"]
	if ok {
		f, ok := offset.(float64)
		if !ok {
			return errors.New("offset must be integer")
		} else {
			off = strconv.Itoa(int(f))
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
				if ok {
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

	a := c.Models()
	dbo := orm.From(c.Storage.Engine, c.Storage.Conn)

	if err := dbo.Debug().Model(c.Model()).
		Where(whr, p...).
		Order(ord).
		Limit(lim).
		Offset(off).
		Find(a).Error; err != nil {
		return err
	}

	return a
}

// TODO: write test cases for CRUD and fetch methods
