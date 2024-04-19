package task

import (
	"github.com/sirupsen/logrus"
	"gochat/config"
	"runtime"
)

type Task struct {
}

func New() *Task {
	return new(Task)
}

func (task *Task) Run() {
	taskConfig := config.Conf.Task
	runtime.GOMAXPROCS(taskConfig.TaskBase.CpuNum)
	//if err := task.InitQueueRedisClient(); err != nil {
	//	logrus.Panicf("task init publishRedisClient fail, err: %s", err.Error())
	//}
	consumer := Consumer{Task: task}
	if err := task.InitConsumerGroup(consumer); err != nil {
		logrus.Panicf("task init consumer err: %s", err)
	}
	if err := task.InitConnectRpcClient(); err != nil {
		logrus.Panicf("task init InitConnectRpcClient fail, err: %s", err.Error())
	}
	task.GoPush()
}
