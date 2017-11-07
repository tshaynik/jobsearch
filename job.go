package jobsearch

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

// Job represents a job opportunity, and keeps track of the process of applying
// the given job.
type Job struct {
	ID              bson.ObjectId
	Title           string    `json:"title"`
	Employer        string    `json:"employer"`
	CalloutURL      string    `json:"callout_url"`
	ApplicationTime time.Time `json:"application_time"`
}

// NewJob returns a new job.
func NewJob(title, employer, url string) *Job {
	j := &Job{
		Title:      title,
		Employer:   employer,
		CalloutURL: url,
	}
	return j
}

// Apply marks a job as having been applied to at the current time.
func (j *Job) Apply() {
	j.ApplicationTime = time.Now()
}
