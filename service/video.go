package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
	. "worker/common"
	"worker/model/task"
	"worker/model/video"
	"worker/sdk"
)

type VideoParams struct {
	Title     string `json:"title"`
	WebUrl    string `json:"web_link"`
	M3u8Url   string `json:"m3u8_link"`
	PicUrl    string `json:"pic_link"`
	PrefixUrl string `json:"prefix_link"`
	KeyUrl    string `json:"key_link"`
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
	defer UnLock()
	defer func() {
		err := recover()
		if err != nil {
			task.SetStatus(T.task, TASK_FAILED)
			task.AppendLog(T.task, fmt.Sprintf("%s", err))
		} else {
			task.AppendLog(T.task, "任务执行完成")
			task.SetStatus(T.task, TASK_SUCCESS)
		}
	}()

	logrus.Info("开始运行任务 【" + strconv.Itoa(T.task.Id) + ":" + T.task.Name + "】")
	task.SetStatus(T.task, TASK_RUNNING)

	if err := T.Check(); err != nil {
		panic(err)
	}
	logrus.Debug(T.VideoParams)

	// 1. 下载图片 TODO:后缀暂时写死
	task.AppendLog(T.task, "开始下载图片")

	savePath := strings.TrimRight(VideoDir, "/") + "/" + T.Title + "/" + "index.jpg"
	if err := DownLoadToFile(T.PicUrl, savePath); err != nil {
		panic(err)
	}

	task.AppendLog(T.task, "图片下载完成")

	// 2. 下载视频
	task.AppendLog(T.task, "开始下载视频")

	tool := sdk.NewM3u8DownloadTool(
		T.VideoParams.Title,
		VideoDir,
		T.VideoParams.M3u8Url,
		T.VideoParams.KeyUrl,
		T.VideoParams.PrefixUrl,
	)

	go tool.Run()
	for {
		status := tool.GetStatus()
		if status == TASK_SUCCESS {
			task.AppendLog(T.task, "视频下载完成")
			break
		}

		if status == TASK_FAILED {
			panic(tool.GetLog())
		}

		task.SetProgress(T.task, tool.GetProgress())
		time.Sleep(1 * time.Second)
	}
	task.SetProgress(T.task, tool.GetProgress())

	// 3. 入库
	if _, err := video.Create(T.Title); err != nil {
		panic(err)
	}
}
