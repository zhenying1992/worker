package model

import (
	"github.com/beego/beego/v2/client/orm"
)

func init() {
	orm.RegisterModel(new(Task))
}