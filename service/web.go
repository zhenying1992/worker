package service

import (
	"bufio"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"worker/model/task"
)

func DownLoadToFile(
	url string,
	file string,
	waitGroup *sync.WaitGroup,
	task *task.TaskModel,
	step int,
) error {
	if waitGroup != nil {
		defer waitGroup.Done()
	}

	if err := os.MkdirAll(path.Dir(file), 0777); err != nil {
		return err
	}

	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	if _, err = io.Copy(f, resp.Body); err != nil {
		return err
	}

	if task != nil {
		_ = task.AddProgress(step)
	}
	return nil
}

func getTsList(rowList []string) []string {
	/*
			存在/20191221/Y0ViQgW4/800kb/hls/VJkC1KcU.ts情况，保存时会多级目录,需要改为一级目录，同时m3u8文件也修改
		 	本方法只返回VJkC1KcU.ts
	*/
	var tsList []string

	for _, row := range rowList {
		if strings.HasSuffix(row, "ts") {
			rowList := strings.Split(row, "/")
			tsList = append(tsList, rowList[len(rowList)-1])
		}
	}
	return tsList
}

func getRowList(file string) ([]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	s := bufio.NewScanner(f)
	var rowList []string

	for s.Scan() {
		rowList = append(rowList, s.Text())
	}
	if err = f.Close(); err != nil {
		return nil,err
	}
	return rowList, nil
}

func DownloadVideo(
	videoParams VideoParams,
	saveDir string,
	task *task.TaskModel,
) error {
	logrus.Debug("开始下载视频")

	videoDir := strings.TrimRight(saveDir, "/") + "/" + videoParams.Title
	m3u8File := videoDir + "/index.m3u8"

	// 1. 下载m3u8文件
	if err := DownLoadToFile(videoParams.M3u8Url, m3u8File, nil, nil, 0); err != nil {
		return err
	}
	if task != nil {
		_ = task.UpdateProgress(10)
	}
	logrus.Debug("下载m3u8文件完成")

	// 2. 下载key文件
	if err := DownLoadToFile(videoParams.KeyUrl, videoDir+"/"+"key.key", nil, nil, 0); err != nil {
		return err
	}
	if task != nil {
		_ = task.UpdateProgress(15)
	}
	logrus.Debug("下载key文件完成")

	// 3. 下载ts文件
	rowList, err := getRowList(m3u8File)
	if err != nil {
		return nil
	}

	tsList := getTsList(rowList)
	logrus.Debug("ts文件", tsList)

	prefixUrl := strings.TrimRight(videoParams.PrefixUrl, "/")
	var wg sync.WaitGroup
	step := 75 / len(tsList)
	for _, ts := range tsList {
		wg.Add(1)
		go DownLoadToFile(prefixUrl+"/"+ts, videoDir+"/"+ts, &wg, task, step)
	}

	wg.Wait()
	if task != nil {
		_ = task.UpdateProgress(90)
	}
	logrus.Debug("ts文件下载完成")

	// 4. 文件中视频路径处理, 把key和ts中的路径全部替换掉
	for idx, row := range rowList {
		if strings.HasPrefix(row, "#EXT-X-KEY") {
			re, _ := regexp.Compile("URI=\".*\"")
			row = re.ReplaceAllString(row, "URI=\"key.key\"")
		}
		if strings.HasSuffix(row, ".ts") {
			colList := strings.Split(row, "/")
			row = colList[len(colList)-1]
		}
		rowList[idx] = row
	}

	// 替换后的文件覆盖到原文件
	file, err := os.OpenFile(m3u8File, os.O_WRONLY, 0777)
	if err != nil {
		return err
	}
	newContent := strings.Join(rowList, "\n")

	if _, err := file.WriteString(newContent); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if task != nil {
		_ = task.UpdateProgress(95)
	}
	return nil
}
