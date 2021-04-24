package task

import (
	"github.com/beego/beego/v2/client/orm"
	"time"
	"worker/common"
)

type TaskModel struct {
	Id int
	Name string
	StartTime time.Time `orm:"auto_now_add;type(datetime)"`
	UpdateTime time.Time `orm:"auto_now;type(datetime)"`
	Status string
	Log string `orm:"size(10000)"`
	Params string `orm:"size(2048)"`
	Progress int
}

func (task *TaskModel) TableName() string {
	return "task"
}

func (task *TaskModel) SetStatus (status string) error {
	task.Status = status

	o := orm.NewOrm()
	if _, err := o.Update(task, "Status"); err != nil {
		return err
	}
	return nil
}

func (task *TaskModel) AppendLog (content string) error {
	content = time.Now().Format("2006-01-02 15:04:00") + " " + content
	if task.Log != "" {
		content = "\n" + content
	}

	task.Log += content

	o := orm.NewOrm()
	if _, err := o.Update(task, "Log"); err != nil {
		return err
	}
	return nil
}

func (task *TaskModel) UpdateProgress(progress int) error {
	o := orm.NewOrm()
	task.Progress = progress
	if _, err := o.Update(task, "Progress"); err != nil {
		return err
	}
	return nil
}

func (task *TaskModel) AddProgress(progress int) error {
	o := orm.NewOrm()
	task.Progress += progress

	if _, err := o.Update(task, "Progress"); err != nil {
		return err
	}
	return nil
}

func ListWaitTask() ([]*TaskModel, error) {
	var taskList []*TaskModel
	o := orm.NewOrm()

	qs := o.QueryTable(new(TaskModel))
	_, err := qs.Filter("status", common.TASK_WAITING).All(&taskList)
	return taskList, err
}
