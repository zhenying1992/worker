package model

import (
	"github.com/beego/beego/v2/client/orm"
	"github.com/sirupsen/logrus"
	"worker/model/task"
	"worker/model/video"
)

func init() {
	dbPrefix := "my_"
	orm.RegisterModelWithPrefix(dbPrefix, new(task.TaskModel))
	orm.RegisterModelWithPrefix(dbPrefix, new(video.Model), new(video.CategoryModel), new(video.TagModel))
	logrus.Info("初始化模型完成")
}
