package common

import (
	"errors"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path"
	"sync"
)

var (
	amount = 0
	mutex  sync.Mutex
)

func GetLock() bool {
	mutex.Lock()
	if amount == MAX_SCHEDULE_TASK {
		mutex.Unlock()
		return false
	}
	amount += 1
	mutex.Unlock()
	return true
}

func UnLock() {
	mutex.Lock()
	if amount == 0 {
		return
	}
	amount -= 1
	mutex.Unlock()
}

func ProtectRun(entry func()) {
	defer func() {
		err := recover()
		if err != nil {
			logrus.Error(err)
		}
	}()

	entry()
}

func DownLoadToFile(
	url string,
	file string,
) error {
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
	if resp.StatusCode != 200 {
		return errors.New("http返回值错误")
	}

	if _, err = io.Copy(f, resp.Body); err != nil {
		return err
	}

	return nil
}
