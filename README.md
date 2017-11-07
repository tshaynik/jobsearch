# Jobsearch

### Overview
An app to make the job application process easier.
The core of the app is a RESTful API written in Golang, using MongoDB.
Jobsearch allows you to:

- Keep track of job opportunities that you are interested in and that you have applied to.
- Organize important dates, such as upcoming interviews.
- Rank jobs to better be able to make decisions about which job to choose.

### API routes
- GET /       : API information
- GET /jobs   : Get information about all jobs that have been added
- POST /jobs  : Add a new job
- UPDATE /job/{id}  : Update a job with the specified ID.
- UPDATE /job/{id}/apply : Mark a job with the specified ID. as applied.
- DELETE /job/{id} : Delete the job with the specified ID.
