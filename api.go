package aqua

import (
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
	Table string
	Model func() interface{}
}

func (c CrudApi) validate() {
	panik.If(c.Storage.Engine == "", "Crud storage engine not specified")
	panik.If(c.Storage.Conn == "", "Crud storage conn not spefieid")
	panik.If(c.Model == nil, "Model not specified")
	panik.If(c.Model() == nil, "Model method returns nil")
	panik.If(!strings.HasPrefix(getSignOfObject(c.Model()), "*st:"), "Model() method must return address of a gorm struct")
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

	return map[string]interface{}{"rows_affected": stmt.RowsAffected}
}

//TODO: update and delete methods
