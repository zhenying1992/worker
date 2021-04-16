package model

import (
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/sirupsen/logrus"
)

type Task struct {
	Id int
	Name string
}

func GetTaskByName(name string) Task {
	logrus.Info("i m here")
	var task Task
	o := orm.NewOrm()
	qs := o.QueryTable(new(Task))
	err := qs.Filter("name", name).One(&task)
	if err == orm.ErrMultiRows {
		fmt.Printf("Return Multi Rows Error")
	}

	if err == orm.ErrNoRows {
		fmt.Println("Not row found")
	}
	return task
}

