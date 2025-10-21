package handler

import "pulse/common/models"

// Handler is an interface that defines the methods that all concrete task handlers must implement.
type Handler interface {
	// Run 执行一个任务
	Run(job *Job) (string, error)
}

// CreateHandler is a factory function that creates a Handler instance based on the job type.
func CreateHandler(j *Job) Handler {
	var handler Handler
	switch j.Type {
	case models.JobTypeCmd:
		handler = new(CMDHandler)
	case models.JobTypeHttp:
		handler = new(HTTPHandler)
	}
	return handler
}
