package sdk

import (
	"bufio"
	"github.com/sirupsen/logrus"
	"os"
	"regexp"
	"strings"
	"sync"
	"worker/common"
)

type M3u8DownloadTool struct {
	name     string        // 视频名
	saveDir  string        // 保存目录
	link     string        // m3u8链接
	keyLink  string        // key链接
	prefix   string        // 视频前缀
	progress int           // 进度
	status   common.STATUS // 状态
	log      string        // 日志
}

func (tool *M3u8DownloadTool) GetStatus() common.STATUS {
	return tool.status
}

func (tool *M3u8DownloadTool) GetProgress() int {
	return tool.progress
}

func (tool *M3u8DownloadTool) GetLog() string {
	return tool.log
}

func (tool *M3u8DownloadTool) Run() {
	logrus.Debug("开始下载视频")

	videoDir := strings.TrimRight(tool.saveDir, "/") + "/" + tool.name
	m3u8File := videoDir + "/index.m3u8"

	// 1. 下载m3u8文件
	if err := common.DownLoadToFile(tool.link, m3u8File); err != nil {
		tool.log += err.Error()
		tool.status = common.TASK_FAILED
		return
	}
	logrus.Debug("下载m3u8文件完成")

	// 2. 下载key文件
	if tool.keyLink != "" {
		if err := common.DownLoadToFile(tool.keyLink, videoDir+"/"+"key.key"); err != nil {
			tool.log += err.Error()
			tool.status = common.TASK_FAILED
			return
		}
	}
	logrus.Debug("下载key文件完成")

	// 3. 下载ts文件
	rowList, err := getRowList(m3u8File)
	if err != nil {
		tool.log += err.Error()
		tool.status = common.TASK_FAILED
		return
	}

	tsList := getTsList(rowList)
	logrus.Debug("ts文件", tsList)

	prefixUrl := strings.TrimRight(tool.prefix, "/")
	var wg sync.WaitGroup
	var lock sync.Mutex

	for _, ts := range tsList {
		wg.Add(1)
		go DownLoadToFileParallel(prefixUrl+"/"+ts, videoDir+"/"+ts, &wg, &lock, tool, 100/len(rowList))
	}

	wg.Wait()
	tool.progress = 100
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
		tool.status = common.TASK_FAILED
		tool.log += err.Error()
		return
	}
	newContent := strings.Join(rowList, "\n")

	if _, err := file.WriteString(newContent); err != nil {
		tool.status = common.TASK_FAILED
		tool.log += err.Error()
		return
	}
	if err := file.Close(); err != nil {
		tool.status = common.TASK_FAILED
		tool.log += err.Error()
		return
	}
	tool.status = common.TASK_SUCCESS
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
		return nil, err
	}
	return rowList, nil
}

func DownLoadToFileParallel(
	url string,
	file string,
	waitGroup *sync.WaitGroup,
	lock *sync.Mutex,
	tool *M3u8DownloadTool,
	step int,
) {
	defer func() {
		waitGroup.Done()
		lock.Unlock()
	}()

	if err := common.DownLoadToFile(url, file); err != nil {
		lock.Lock()
		tool.status = common.TASK_FAILED
		tool.log += err.Error()
	} else {
		lock.Lock()
		tool.progress += step
	}
}

func NewM3u8DownloadTool(
	name string,
	saveDir string,
	link string,
	keyLink string,
	prefix string,
) *M3u8DownloadTool {
	return &M3u8DownloadTool{
		name,
		saveDir,
		link,
		keyLink,
		prefix,
		0,
		common.TASK_WAITING,
		"",
	}
}
