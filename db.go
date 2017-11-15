package jobsearch

import (
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const app = "jobsearch"

// DB is a wrapper for a MongoDB session, that abstracts database operations.
type DB struct {
	session *mgo.Session
}

// NewDB returns a new instance of DB
func NewDB(session *mgo.Session) *DB {
	return &DB{session: session}
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
func (db DB) FindAllJobs(login string) ([]*Job, error) {
	s := db.session.Copy()
	defer s.Close()
	var jobs []*Job
	if err := s.DB(app).C("jobs").Find(bson.M{"user_login": login}).All(&jobs); err != nil {
		return nil, err
	}
	return jobs, nil
}

// SaveAuthState commits the state object used for validating the OAuth 2.0
// authentication into the database.
func (db DB) SaveAuthState(st *State) error {
	s := db.session.Copy()
	defer s.Close()
	return s.DB(app).C("auth_state").Insert(st)
}

// IsValidAuthState returns if a state object with a matching random_string
// has been registered. The state entry is removed from the database, as it is
// intended for one time use only.
func (db DB) IsValidAuthState(random string) (bool, error) {
	s := db.session.Copy()
	defer s.Close()
	err := s.DB(app).C("auth_state").Remove(bson.M{"random_string": random})
	if err == mgo.ErrNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// SaveAuthJWT stores an authorization token in the database.
func (db DB) SaveAuthJWT(at *AuthJWT) error {
	s := db.session.Copy()
	defer s.Close()
	return s.DB(app).C("auth_token").Insert(at)
}

// ValidateAuthJWT ensures that the correct token is stored in the database
func (db DB) ValidateAuthJWT(token string) bool {
	s := db.session.Copy()
	defer s.Close()
	count, err := s.DB(app).C("auth_token").Find(bson.M{"bearer_token": token}).Count()
	if count == 0 || err != nil {
		return false
	}
	return true
}

// RemoveAuthJWT removes an AuthJWT from the database to log out.
func (db DB) RemoveAuthJWT(token string) error {
	s := db.session.Copy()
	defer s.Close()
	return s.DB(app).C("auth_token").Remove(bson.M{"bearer_token": token})
}

// UpsertUser matches an existing user by username (Github login) in the database, or
// inserts a new user if none exists already.
func (db DB) UpsertUser(u *User) error {
	s := db.session.Copy()
	defer s.Close()
	_, err := s.DB(app).C("users").Upsert(bson.M{"login": u.Login}, u)
	return err
}

// GetUser retrieves a user from the database with a given github login.
func (db DB) GetUser(login string) (*User, error) {
	s := db.session.Copy()
	defer s.Close()
	var u User
	err := s.DB(app).C("users").Find(bson.M{"login": login}).One(&u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
