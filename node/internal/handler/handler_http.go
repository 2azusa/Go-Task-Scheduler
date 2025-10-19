package handler

import (
	"pulse/common/models"
	"pulse/common/pkg/httpclient"
	"strings"
	"time"
)

// HTTPHandler handles the execution of HTTP-based jobs.
type HTTPHandler struct{}

// HttpExecTimeout defines the maximum allowed timeout for an HTTP request.
const HttpExecTimeout = 300

// Run executes an HTTP job.
func (h *HTTPHandler) Run(job *Job) (result string, err error) {
	proc := &JobProc{
		JobProc: &models.JobProc{
			ID:       0, // HTTP tasks have no OS process ID.
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
		// For POST requests, the command is split by '?' into URL and body.
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
