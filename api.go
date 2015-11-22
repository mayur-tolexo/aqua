package aqua

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/thejackrabbit/aero/db/cstr"
	"github.com/thejackrabbit/aero/db/orm"
	"github.com/thejackrabbit/aero/panik"
	"github.com/thejackrabbit/aero/strukt"
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
		fmt.Println("2b")
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

	where, ok := data["where"]
	if !ok {
		return errors.New("where clause not specified")
	}

	p := make([]interface{}, 0)
	params, ok := data["params"]
	if ok {
		p, ok = params.([]interface{})
		if !ok {
			return errors.New("params must be an array")
		}
	}

	a := c.Models()
	dbo := orm.From(c.Storage.Engine, c.Storage.Conn)

	if err := dbo.Debug().Model(c.Model()).Where(where, p...).Find(a).Error; err != nil {
		return err
	}
	return a
}

// TODO: write test cases for CRUD and fetch methods
