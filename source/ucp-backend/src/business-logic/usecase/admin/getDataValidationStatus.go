package admin

import (
	"errors"
	"sync"
	"tah/core/core"
	"tah/core/customerprofiles"
	accpmappings "tah/ucp/src/business-logic/model/accp-mappings"
	model "tah/ucp/src/business-logic/model/common"
	"tah/ucp/src/business-logic/usecase/registry"
	"tah/ucp/src/business-logic/validator"

	"github.com/aws/aws-lambda-go/events"
)

type GetDataValidationStatus struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewGetDataValidationStatus() *GetDataValidationStatus {
	return &GetDataValidationStatus{name: "GetDataValidationStatus"}
}

func (u *GetDataValidationStatus) Name() string {
	return u.name
}
func (u *GetDataValidationStatus) Tx() core.Transaction {
	return *u.tx
}
func (u *GetDataValidationStatus) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *GetDataValidationStatus) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *GetDataValidationStatus) Registry() *registry.Registry {
	return u.reg
}

func (u *GetDataValidationStatus) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	return registry.CreateRequest(u, req)
}

func (u *GetDataValidationStatus) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Log("Validating request")
	if rq.Pagination.PageSize == 0 {
		return errors.New("Pagse size must be greater than 0")
	}
	return nil
}

func (u *GetDataValidationStatus) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	accpSourceBucket := u.reg.Env["CONNECT_PROFILE_SOURCE_BUCKET"]
	uc := validator.Usecase{Uc: u}
	res := model.ResponseWrapper{
		DataValidation: []model.ValidationError{},
	}
	folders := []string{
		ACCP_SUB_FOLDER_AIR_BOOKING,
		ACCP_SUB_FOLDER_EMAIL_HISTORY,
		ACCP_SUB_FOLDER_PHONE_HISTORY,
		ACCP_SUB_FOLDER_AIR_LOYALTY,
		ACCP_SUB_FOLDER_CLICKSTREAM,
		ACCP_SUB_FOLDER_GUEST_PROFILE,
		ACCP_SUB_FOLDER_HOTEL_LOYALTY,
		ACCP_SUB_FOLDER_HOTEL_BOOKING,
		ACCP_SUB_FOLDER_PAX_PROFILE,
	}
	businessMap := map[string]func() []customerprofiles.FieldMapping{
		ACCP_SUB_FOLDER_AIR_BOOKING:        accpmappings.BuildAirBookingMapping,
		ACCP_SUB_FOLDER_EMAIL_HISTORY:      accpmappings.BuildEmailHistoryMapping,
		ACCP_SUB_FOLDER_PHONE_HISTORY:      accpmappings.BuildPhoneHistoryMapping,
		ACCP_SUB_FOLDER_AIR_LOYALTY:        accpmappings.BuildAirLoyaltyMapping,
		ACCP_SUB_FOLDER_CLICKSTREAM:        accpmappings.BuildClickstreamMapping,
		ACCP_SUB_FOLDER_GUEST_PROFILE:      accpmappings.BuildGuestProfileMapping,
		ACCP_SUB_FOLDER_HOTEL_LOYALTY:      accpmappings.BuildHotelLoyaltyMapping,
		ACCP_SUB_FOLDER_HOTEL_BOOKING:      accpmappings.BuildHotelBookingMapping,
		ACCP_SUB_FOLDER_PAX_PROFILE:        accpmappings.BuildPassengerProfileMapping,
		ACCP_SUB_FOLDER_HOTEL_STAY_MAPPING: accpmappings.BuildHotelStayMapping,
	}
	var lastErr error
	wg := sync.WaitGroup{}
	wg.Add(len(folders))
	mu := sync.Mutex{}
	for _, path := range folders {
		go func(filePath string) {
			valErrs, err := uc.ValidateAccpRecords(req.Pagination, accpSourceBucket, filePath, businessMap[filePath]())
			for _, valErr := range valErrs {
				mu.Lock()
				res.DataValidation = append(res.DataValidation, valErr)
				mu.Unlock()
			}
			if err != nil {
				lastErr = err
			}
			wg.Done()
		}(path)
	}
	wg.Wait()
	if lastErr != nil {
		return res, lastErr
	}
	//TODO: to centralize this
	bizObjectBuckets := map[string]string{
		"S3_HOTEL_BOOKING": u.reg.Env["S3_HOTEL_BOOKING"],
		"S3_AIR_BOOKING":   u.reg.Env["S3_AIR_BOOKING"],
		"S3_GUEST_PROFILE": u.reg.Env["S3_GUEST_PROFILE"],
		"S3_PAX_PROFILE":   u.reg.Env["S3_PAX_PROFILE"],
		"S3_STAY_REVENUE":  u.reg.Env["S3_STAY_REVENUE"],
		"S3_CLICKSTREAM":   u.reg.Env["S3_CLICKSTREAM"],
	}
	for _, bucketName := range bizObjectBuckets {
		valErrs, err := uc.ValidateBizObjects(bucketName, "")
		if err != nil {
			return res, err
		}
		for _, valErr := range valErrs {
			res.DataValidation = append(res.DataValidation, valErr)
		}
	}
	return res, nil
}

func (u *GetDataValidationStatus) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
