package main

import (
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	_ "github.com/go-sql-driver/mysql"
	"github.com/orandin/lumberjackrus"
	"github.com/sirupsen/logrus"
	"time"
	. "worker/common"
	_ "worker/model"
	"worker/model/task"
	"worker/service"
)

func initLog() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
	})

	hook, err := lumberjackrus.NewHook(
		&lumberjackrus.LogFile{
			Filename:   GeneralLog,
			MaxSize:    100,
			MaxAge:     1,
			MaxBackups: 3,
			LocalTime:  false,
			Compress:   false,
		},
		logrus.DebugLevel,
		&logrus.TextFormatter{
			DisableColors: true,
		},
		&lumberjackrus.LogFileOpts{
			logrus.InfoLevel: &lumberjackrus.LogFile{
				Filename:   InfoLog,
				MaxSize:    1,
				MaxBackups: 3,
				LocalTime:  false,
				Compress:   false,
			},
			logrus.ErrorLevel: &lumberjackrus.LogFile{
				Filename:   ErrorLog,
				MaxSize:    100,
				MaxAge:     1,
				MaxBackups: 3,
				LocalTime:  false,
				Compress:   false,
			},
		},
	)

	if err != nil {
		panic(err)
	}

	logrus.AddHook(hook)
	logrus.Info("日志初始化完成")
}

func initDb() {
	dataSource := DbUser + ":" + DbPass + "@tcp(" + DbHost + ":" + DbPort + ")/" +
		DbName + "?charset=utf8&timeout=3s&loc=Asia%2FShanghai"
	logrus.Debug(dataSource)

	err := orm.RegisterDataBase("default", "mysql", dataSource)
	if err != nil {
		logrus.Panic("数据库连接失败")
	}

	logrus.Info("数据库连接成功")

	orm.RunCommand()
	err = orm.RunSyncdb("default", false, true)
	if err != nil {
		logrus.Panic("数据库同步失败")
	}
}

func init() {
	initLog()
	initDb()
}

func scheduleTask() {
	defer func(){
		err := recover()
		if err != nil {
			logrus.Error(err)
		}
	}()

	taskList, err := task.ListWaitTask()
	if err != nil {
		panic(err)
	}

	for _, waitingTask := range taskList {
		taskService, err := service.GetService(waitingTask)
		if err != nil {
			logrus.Error(err)
			continue
		}

		if GetLock() {
			go ProtectRun(taskService.Run)
		} else {
			logrus.Warn("调度配额已满")
		}
	}

}

func listen() {
	listenLog := fmt.Sprintf("监听间隔: %d", ListenInterval)
	logrus.Info(listenLog)

	for {
		scheduleTask()
		time.Sleep(time.Duration(ListenInterval) * time.Second)
		logrus.Info("等待下一次轮训")
	}
}

func main() {
	logrus.Info("开始监听任务...")
	listen()
}
