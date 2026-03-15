package domain

type ProcessingStatus string

const (
	JobPending    ProcessingStatus = "pending"
	JobProcessing ProcessingStatus = "processing"
	JobDone       ProcessingStatus = "done"
	JobFailed     ProcessingStatus = "failed"
)
