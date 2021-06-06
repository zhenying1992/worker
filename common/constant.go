package common

const ENV_PREFIX = "MY_"
const MAX_SCHEDULE_TASK = 2

const (
	VIDEO_DOWNLOAD = "视频下载"
)

type STATUS string

const (
	TASK_WAITING STATUS = "waiting"
	TASK_RUNNING STATUS = "running"
	TASK_FAILED  STATUS = "failed"
	TASK_SUCCESS STATUS = "success"
)
