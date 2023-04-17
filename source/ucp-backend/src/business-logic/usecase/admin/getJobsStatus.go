package admin

import (
	"sync"
	"tah/core/core"
	model "tah/ucp/src/business-logic/model/common"
	"tah/ucp/src/business-logic/usecase/registry"

	"github.com/aws/aws-lambda-go/events"
)

type GetJobsStatus struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewGetJobsStatus() *GetJobsStatus {
	return &GetJobsStatus{name: "GetJobsStatus"}
}

func (u *GetJobsStatus) Name() string {
	return u.name
}
func (u *GetJobsStatus) Tx() core.Transaction {
	return *u.tx
}
func (u *GetJobsStatus) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *GetJobsStatus) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *GetJobsStatus) Registry() *registry.Registry {
	return u.reg
}

func (u *GetJobsStatus) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return registry.CreateRequest(u, req)
}

func (u *GetJobsStatus) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Log("Validating request")
	return nil
}

func (u *GetJobsStatus) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	res := model.ResponseWrapper{}
	jobNames := []string{
		u.reg.Env["HOTEL_BOOKING_JOB_NAME"],
		u.reg.Env["AIR_BOOKING_JOB_NAME"],
		u.reg.Env["GUEST_PROFILE_JOB_NAME"],
		u.reg.Env["PAX_PROFILE_JOB_NAME"],
		u.reg.Env["CLICKSTREAM_JOB_NAME"],
		u.reg.Env["HOTEL_STAY_JOB_NAME"],
		u.reg.Env["HOTEL_BOOKING_JOB_NAME_CUSTOMER"],
		u.reg.Env["AIR_BOOKING_JOB_NAME_CUSTOMER"],
		u.reg.Env["GUEST_PROFILE_JOB_NAME_CUSTOMER"],
		u.reg.Env["PAX_PROFILE_JOB_NAME_CUSTOMER"],
		u.reg.Env["CLICKSTREAM_JOB_NAME_CUSTOMER"],
		u.reg.Env["HOTEL_STAY_JOB_NAME_CUSTOMER"],
	}
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	jobs := []model.JobSummary{}
	wg.Add(len(jobNames))
	var jobStatusErr error
	for _, jobName := range jobNames {
		go func(name string) {
			lastRun, status, err := u.reg.Glue.GetJobRunStatus(name)
			if err != nil {
				u.tx.Log("Error getting job run status %v => %v", name, err)
				jobStatusErr = err
			}

			mu.Lock()
			jobs = append(jobs, model.JobSummary{
				JobName:     name,
				LastRunTime: lastRun,
				Status:      status,
			})
			mu.Unlock()
			wg.Done()
		}(jobName)
	}
	wg.Wait()

	res.AwsResources = model.AwsResources{
		Jobs: jobs,
	}

	return res, jobStatusErr
}

func (u *GetJobsStatus) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
