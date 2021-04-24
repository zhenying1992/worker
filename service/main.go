package service

import (
	"errors"
	. "worker/common"
	"worker/model/task"
)


type TaskServicer interface {
	Check() error
	Run()
}

func GetService(task *task.TaskModel) (TaskServicer, error) {
	name := task.Name

	switch name {
	case VIDEO_DOWNLOAD:
		return &VideoService{task: task}, nil
	default:
		return nil, errors.New("未找到任务【" + name + "】")
	}
}
