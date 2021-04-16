package main

import (
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	_ "github.com/go-sql-driver/mysql"
	"github.com/orandin/lumberjackrus"
	"github.com/sirupsen/logrus"
	"strconv"
	"worker/model"
)


func init_log() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
	})

	general_log, _ := web.AppConfig.String("general_log")
	info_log, _ := web.AppConfig.String("info_log")
	error_log, _ := web.AppConfig.String("error_log")

	hook, err := lumberjackrus.NewHook(
		&lumberjackrus.LogFile{
			Filename:   general_log,
			MaxSize:    100,
			MaxAge:     1,
			MaxBackups: 1,
			LocalTime:  false,
			Compress:   false,
		},
		logrus.InfoLevel,
		&logrus.TextFormatter{
				DisableColors: true,
		},
		&lumberjackrus.LogFileOpts{
			logrus.InfoLevel: &lumberjackrus.LogFile{
				Filename:   info_log,
				MaxSize:    100,
				MaxAge:     1,
				MaxBackups: 1,
				LocalTime:  false,
				Compress:   false,
			},
			logrus.ErrorLevel: &lumberjackrus.LogFile{
				Filename:   error_log,
				MaxSize:    100,
				MaxAge:     1,
				MaxBackups: 1,
				LocalTime:  false,
				Compress:   false,
			},
		},
	)

	if err != nil {
		panic(err)
	}
	logrus.AddHook(hook)

}

func init_db() {
	db_host, _ := web.AppConfig.String("database_host")
	db_port, _ := web.AppConfig.Int("database_port")
	db_port_string := strconv.Itoa(db_port)
	db_name, _ := web.AppConfig.String("database_name")
	db_user, _ := web.AppConfig.String("database_user")
	db_passwd, _ := web.AppConfig.String("database_passwd")

	data_source := db_user + ":" + db_passwd + "@tcp(" + db_host + ":" + db_port_string + ")/" + db_name + "?charset=utf8&timeout=3s"

	err := orm.RegisterDataBase("default", "mysql", data_source)
	if err != nil {
		logrus.Panic("数据库连接失败")
	}

	logrus.Info("数据库连接成功")

}

func init() {
	init_log()
	init_db()
	logrus.Debug("hahah")
	logrus.Warn("warn")
	logrus.Error("error")
}

func main() {
	fmt.Print(model.GetTaskByName("歌之撸图片"))
	//web.Run()
}
