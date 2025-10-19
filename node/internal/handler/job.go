package handler

import (
	"encoding/json"
	"fmt"
	"pulse/common/models"
	"pulse/common/pkg/config"
	"pulse/common/pkg/etcdclient"
	"pulse/common/pkg/logger"
	"pulse/common/pkg/notify"
	"pulse/common/pkg/utils"
	"pulse/common/pkg/utils/errors"
	"runtime"
	"strconv"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"github.com/jakecoffman/cron"
)

// Job wraps models.Job for node-side operations.
type Job struct {
	*models.Job
}

// Jobs is a map of jobs, keyed by job ID.
type Jobs map[int]*Job

// JobKey generates a unique etcd key for a job on a specific node.
func JobKey(nodeUUID string, jobId int) string {
	return fmt.Sprintf(etcdclient.KeyEtcdJob, nodeUUID, jobId)
}

// GetJobAndRev retrieves a job and its revision from etcd.
func GetJobAndRev(nodeUUID string, jobId int) (job *Job, rev int64, err error) {
	resp, err := etcdclient.Get(JobKey(nodeUUID, jobId))
	if err != nil {
		return
	}

	if resp.Count == 0 {
		err = errors.ErrNotFound
		return
	}

	rev = resp.Kvs[0].ModRevision
	if err = json.Unmarshal(resp.Kvs[0].Value, &job); err != nil {
		return
	}

	job.SplitCmd()
	return
}

// GetJobs retrieves all jobs for a specific node from etcd.
func GetJobs(nodeUUID string) (jobs Jobs, err error) {
	resp, err := etcdclient.Get(fmt.Sprintf(etcdclient.KeyEtcdJobProfile, nodeUUID), clientv3.WithPrefix())
	if err != nil {
		return
	}

	count := len(resp.Kvs)
	jobs = make(Jobs, count)
	if count == 0 {
		return
	}

	for _, j := range resp.Kvs {
		job := new(Job)
		if e := json.Unmarshal(j.Value, job); e != nil {
			logger.GetLogger().Warn(fmt.Sprintf("job[%s] unmarshal err: %s", string(j.Key), e.Error()))
			continue
		}
		if err := job.Check(); err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("job[%s] is invalid: %s", string(j.Key), err.Error()))
			continue
		}
		jobs[job.ID] = job
	}
	return
}

// RunWithRecovery executes the job with a panic recovery mechanism.
func (j *Job) RunWithRecovery() {
	defer func() {
		if r := recover(); r != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			logger.GetLogger().Warn(fmt.Sprintf("panic running job: %v\n%s", r, buf))
		}
	}()

	t := time.Now()
	jobLogId, err := j.CreateJobLog()
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("Failed to write to job log with jobID:%d nodeUUID: %s error: %s", j.ID, j.RunOn, err.Error()))
	}

	h := CreateHandler(j)
	if h == nil {
		return
	}

	result, runErr := h.Run(j)
	if runErr != nil {
		err = j.Fail(jobLogId, t, runErr.Error(), 0)
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("Failed to write to job log with jobID:%d nodeUUID: %s error: %s", j.ID, j.RunOn, err.Error()))
		}

		node := &models.Node{UUID: j.RunOn}
		node.FindByUUID()
		var to []string
		for _, userId := range j.NotifyToArray {
			userModel := &models.User{ID: userId}
			err = userModel.FindById()
			if err != nil {
				continue
			}
			if j.NotifyType == notify.NotifyTypeMail {
				to = append(to, userModel.Email)
			} else if j.NotifyType == notify.NotifyTypeWebHook && config.GetConfigModels().WebHook.Kind == "feishu" {
				to = append(to, userModel.UserName)
			}
		}

		msg := &notify.Message{
			Type:      j.NotifyType,
			IP:        fmt.Sprintf("%s:%s", node.IP, node.PID),
			Subject:   fmt.Sprintf("任务[%s]立即执行失败", j.Name),
			Body:      fmt.Sprintf("job[%d] run on node[%s] oince execute failed, output: %s, eror:%s", j.ID, j.RunOn, result, runErr.Error()),
			To:        to,
			OccurTime: time.Now().Format(utils.TimeFormatSecond),
		}
		go notify.Send(msg)
	} else {
		err = j.Success(jobLogId, t, result, 0)
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("Failed to write to job log with jobID: %d nodeUUID: %s error:%s", j.ID, j.RunOn, err.Error()))
		}
	}
}

// CreateJob wraps a Job into a cron.FuncJob.
func CreateJob(j *Job) cron.FuncJob {
	h := CreateHandler(j)
	if h == nil {
		return nil
	}

	jobFunc := func() {
		logger.GetLogger().Info(fmt.Sprintf("start the job#%s#command-%s", j.Name, j.Command))
		var execTimes int = 1
		if j.RetryTimes > 0 {
			execTimes += j.RetryTimes
		}

		var i = 0
		var output string
		var runErr error
		var err error
		var jobLogId int
		t := time.Now()

		jobLogId, err = j.CreateJobLog()
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("Failed to write to job with jobId: %d nodeUUID: %s error:%s", j.ID, j.RunOn, err.Error()))
		}

		for i < execTimes {
			output, runErr = h.Run(j)
			if runErr == nil {
				err = j.Success(jobLogId, t, output, i)
				if err != nil {
					logger.GetLogger().Warn(fmt.Sprintf("Failed to write to job log with jobID: %d nodeUUID: %s error: %s", j.ID, j.RunOn, err.Error()))
				}
			}
			i++
			if i < execTimes {
				logger.GetLogger().Warn(fmt.Sprintf("job execution failure#jobId-%d#retry %d times #output-%s#error-%v", j.ID, i, output, runErr))
				if j.RetryInterval > 0 {
					time.Sleep(time.Duration(j.RetryInterval) * time.Second)
				} else {
					// Default retry interval increases with each attempt.
					time.Sleep(time.Duration(i) * time.Minute)
				}
			}
		}

		err = j.Fail(jobLogId, t, runErr.Error(), execTimes-1)
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("Failed to write to job with jobID:%d nodeID: %s error: %s", j.ID, j.RunOn, err.Error()))
		}

		node := &models.Node{UUID: j.RunOn}
		err = node.FindByUUID()
		if err != nil {
			logger.GetLogger().Warn(fmt.Sprintf("Failed to find node with jobID: %d nodeUUID: %s error:%s", j.ID, j.RunOn, err.Error()))
		}
		var to []string
		for _, userId := range j.NotifyToArray {
			userModel := &models.User{ID: userId}
			err = userModel.FindById()
			if err != nil {
				continue
			}
			if j.NotifyType == notify.NotifyTypeMail {
				to = append(to, userModel.Email)
			} else if j.NotifyType == notify.NotifyTypeWebHook && config.GetConfigModels().WebHook.Kind == "feishu" {
				to = append(to, userModel.UserName)
			}
		}
		msg := &notify.Message{
			Type:      j.NotifyType,
			IP:        fmt.Sprintf("%s:%s", node.IP, node.PID),
			Subject:   fmt.Sprintf("任务[%s]执行失败", j.Name),
			Body:      fmt.Sprintf("job[%d] run on node[%s] execute failed ,retry %d times ,output :%s, error:%v", j.ID, j.RunOn, j.RetryTimes, output, runErr),
			To:        to,
			OccurTime: time.Now().Format(utils.TimeFormatSecond),
		}
		go notify.Send(msg)
	}
	return jobFunc
}

// WatchJobs creates a watch channel for job changes on a specific node.
func WatchJobs(nodeUUID string) clientv3.WatchChan {
	return etcdclient.Watch(fmt.Sprintf(etcdclient.KeyEtcdJobProfile, nodeUUID), clientv3.WithPrefix())
}

// GetJobIDFromKey extracts the job ID from an etcd key.
func GetJobIDFromKey(key string) int {
	index := strings.LastIndex(key, "/")
	if index < 0 {
		return 0
	}
	jobId, err := strconv.Atoi(key[index+1:])
	if err != nil {
		return 0
	}
	return jobId
}

// CreateJobLog creates a log entry for a job execution.
func (j *Job) CreateJobLog() (int, error) {
	start := time.Now()
	jobLog := &models.JobLog{
		Name:      j.Name,
		JobId:     j.ID,
		Command:   j.Command,
		IP:        j.Ip,
		Hostname:  j.Hostname,
		NodeUUID:  j.RunOn,
		Spec:      j.Spec,
		StartTime: start.Unix(),
	}
	return jobLog.Insert()
}

// UpdateJobLog updates a job log entry.
func UpdateJobLog(jobLogId int, start time.Time, output string, retry int, success bool) error {
	end := time.Now()
	jobLog := &models.JobLog{
		ID:         jobLogId,
		StartTime:  start.Unix(),
		RetryTimes: retry,
		Success:    success,
		Output:     output,
		EndTime:    end.Unix(),
	}
	return jobLog.Update()
}

// Success marks the job log as successful.
func (j *Job) Success(jobLogId int, start time.Time, output string, retry int) error {
	return UpdateJobLog(jobLogId, start, output, retry, true)
}

// Fail marks the job log as failed.
func (j *Job) Fail(jobLogId int, start time.Time, errMsg string, retry int) error {
	return UpdateJobLog(jobLogId, start, errMsg, retry, false)
}
