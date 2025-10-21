package handler

import (
	"pulse/common/models"
	"pulse/common/pkg/httpclient"
	"strings"
	"time"
)

// HTTPHandler 处理基于 HTTP 的任务的执行
type HTTPHandler struct{}

// HttpExecTimeout 定义 HTTP 请求的最大允许超时时间
const HttpExecTimeout = 300

// Run 执行一个 HTTP 任务
func (h *HTTPHandler) Run(job *Job) (result string, err error) {
	proc := &JobProc{
		JobProc: &models.JobProc{
			ID:       0, // HTTP 任务没有操作系统进程 ID
			JobID:    job.ID,
			NodeUUID: job.RunOn,
		},
	}
	proc.JobProcVal.Time = time.Now()

	err = proc.Start()
	if err != nil {
		return
	}
	defer proc.Stop()

	if job.Timeout <= 0 || job.Timeout > HttpExecTimeout {
		job.Timeout = HttpExecTimeout
	}

	if job.HttpMethod == models.HttpMethodGet {
		result, err = httpclient.Get(job.Command, job.Timeout)
	} else {
		// 对于 POST 请求，命令按“?”分割成 URL 和 body
		urlFields := strings.Split(job.Command, "?")
		url := urlFields[0]
		var body string
		if len(urlFields) >= 2 {
			body = urlFields[1]
		}
		result, err = httpclient.PostJson(url, body, job.Timeout)
	}

	return
}
