package jobsearch

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

// Job represents a job opportunity, and keeps track of the process of applying
// the given job.
type Job struct {
	ID              bson.ObjectId
	UserLogin       string    `json:"user_login" bson:"user_login"`
	Title           string    `json:"title" bson:"title"`
	Employer        string    `json:"employer" bson:"employer"`
	CalloutURL      string    `json:"callout_url" bson:"callout_url"`
	ApplicationTime time.Time `json:"application_time,omitempty" bson:"application_time,omitempty"`
}

// NewJob returns a new job.
func NewJob(login, title, employer, url string) *Job {
	j := &Job{
		ID:         bson.NewObjectId(),
		UserLogin:  login,
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
