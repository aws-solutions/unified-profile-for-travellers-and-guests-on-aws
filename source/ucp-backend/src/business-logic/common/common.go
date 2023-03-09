package common

import (
	"log"
	"strconv"
	"tah/core/core"
)

type Context struct {
	Tx     *core.Transaction
	Region string
}

func Init(tx *core.Transaction, region string) *Context {
	cx := Context{}
	cx.Tx = tx
	cx.Region = region
	return &cx
}

func (c Context) Log(format string, v ...interface{}) {
	if c.Tx != nil {
		c.Tx.Log(format, v)
	}
	log.Printf("[no_tx] "+format, v)
}

func (c Context) ParseQueryParamInt(param string) int {
	i, err := strconv.Atoi(param)
	if err != nil {
		c.Log("WARNING invalid query param of type int")
		return 0
	}
	return i
}
