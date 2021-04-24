package common

import (
	"github.com/sirupsen/logrus"
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