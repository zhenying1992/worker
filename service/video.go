package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"worker/common"
	"worker/model/task"
	"worker/model/video"
)

type VideoParams struct {
	Title     string `json:"title"`
	WebUrl   string `json:"web_link"`
	M3u8Url  string `json:"m3u8_link"`
	PicUrl   string `json:"pic_link"`
	PrefixUrl    string `json:"prefix_link"`
	KeyUrl string `json:"key_link"`
}

type VideoService struct {
	task        *task.TaskModel
	VideoParams // 检查参数后放入到此
}

func (T *VideoService) Check() error {
	err := json.Unmarshal([]byte(T.task.Params), &T.VideoParams)
	if err != nil {
		return err
	}

	if T.Title == "" {
		return errors.New("缺少视频名称: Title")
	}

	if T.M3u8Url == "" {
		return errors.New("缺少M3U8链接")
	}

	if T.PicUrl == "" {
		return errors.New("缺少图片链接")
	}
	return nil
}

func (T *VideoService) Run() {
	/*
		5% 图片下载完成
		10% m3u8文件下载完成
		15% key文件下载完成
		...
		90% 所有ts文件下载完成
		95% 文件格式化完成
		100% 入库
	 */
	defer common.UnLock()
	defer func() {
		err := recover()
		if err != nil {
			x := fmt.Sprintf("%s", err)
			logrus.Error(x)
			if innerError := T.task.SetStatus(common.TASK_FAILED); innerError != nil {
				logrus.Error(innerError)
			}
			if innerError := T.task.AppendLog(x); innerError != nil {
				logrus.Error(innerError)
			}
		} else {
			if innerError := T.task.SetStatus(common.TASK_SUCCESS); innerError != nil {
				logrus.Error(innerError)
			}
		}
	}()

	logrus.Info("开始运行任务 【" + strconv.Itoa(T.task.Id) + ":" + T.task.Name + "】")

	if err := T.task.SetStatus(common.TASK_RUNNING); err != nil {
		panic(err)
	}

	if err := T.Check(); err != nil {
		panic(err)
	}
	logrus.Debug(T.VideoParams)

	// 1. 下载图片 进度5% TODO:后缀暂时写死
	if err := T.task.AppendLog("开始下载图片"); err != nil {
		panic(err)
	}

	savePath := strings.TrimRight(common.VideoDir, "/") + "/" + T.Title + "/" + "index.jpg"
	if err := DownLoadToFile(T.PicUrl, savePath, nil, nil, 0); err != nil {
		panic(err)
	}

	if err := T.task.AppendLog("图片下载完成"); err != nil {
		panic(err)
	}
	if err := T.task.UpdateProgress(5); err != nil {
		panic(err)
	}
	logrus.Info("1. 图片下载完成")

	// 2. 下载视频 进度95%
	if err := T.task.AppendLog("开始下载视频"); err != nil {
		panic(err)
	}

	if err := DownloadVideo(T.VideoParams, common.VideoDir, T.task); err != nil {
		panic(err)
	}

	if err := T.task.AppendLog("视频下载完成"); err != nil {
		panic(err)
	}
	if err := T.task.UpdateProgress(95); err != nil {
		panic(err)
	}
	logrus.Info("2. 视频下载完成")

	// 3. 入库 进度100%
	if _, err := video.Create(T.Title); err != nil {
		panic(err)
	}

	if err := T.task.AppendLog("任务完成"); err != nil {
		panic(err)
	}
	if err := T.task.UpdateProgress(100); err != nil {
		panic(err)
	}
	logrus.Info("3. 任务执行完成")
}
