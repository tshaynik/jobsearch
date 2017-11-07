package jobsearch

import mgo "gopkg.in/mgo.v2"

const app = "jobsearch"

// DB is a wrapper for a MongoDB session, that abstracts database operations.
type DB struct {
	session *mgo.Session
}

// CreateJob commits a job to the database
func (db DB) CreateJob(j *Job) error {
	s := db.session.Copy()
	defer s.Close()
	return s.DB(app).C("jobs").Insert(j)
}

// DeleteJob deletes a job with a given ID from the database
func (db DB) DeleteJob(id string) error {
	s := db.session.Copy()
	defer s.Close()
	return s.DB(app).C("jobs").RemoveId(id)
}

// UpdateJob replaces an existing job in the database with a new job.
func (db DB) UpdateJob(j *Job) error {
	s := db.session.Copy()
	defer s.Close()
	return s.DB(app).C("jobs").Update(j.ID, j)
}

// FindAllJobs returns all jobs in the database.
func (db DB) FindAllJobs() ([]*Job, error) {
	s := db.session.Copy()
	defer s.Close()
	var jobs []*Job
	if err := s.DB(app).C("jobs").Find(nil).All(&jobs); err != nil {
		return nil, err
	}
	return jobs, nil
}
