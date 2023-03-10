package traveller

import (
	"tah/core/core"
	"tah/core/customerprofiles"
	model "tah/ucp/src/business-logic/model/common"
	"tah/ucp/src/business-logic/usecase/registry"

	"github.com/aws/aws-lambda-go/events"
)

var SEARCH_KEY_LAST_NAME = "LastName"
var SEARCH_KEY_FIRST_NAME = "FirstName"
var SEARCH_KEY_EMAIL = "PersonalEmailAddress"
var SEARCH_KEY_PHONE = "PhoneNumber"
var SEARCH_KEY_ACCOUNT_NUMBER = "AccountNumber"
var SEARCH_KEY_CONF_NUMBER = "confirmationNumber"

type SearchProfile struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewSearchProfile() *SearchProfile {
	return &SearchProfile{name: "SearchProfile"}
}

func (u *SearchProfile) Name() string {
	return u.name
}
func (u *SearchProfile) Tx() core.Transaction {
	return *u.tx
}
func (u *SearchProfile) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *SearchProfile) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *SearchProfile) Registry() *registry.Registry {
	return u.reg
}

func (u *SearchProfile) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	rw := model.RequestWrapper{}
	rw.SearchRq = model.SearchRq{}
	rw.SearchRq.LastName = req.QueryStringParameters["lastName"]
	rw.SearchRq.Phone = req.QueryStringParameters["phone"]
	rw.SearchRq.Email = req.QueryStringParameters["email"]
	rw.SearchRq.LoyaltyID = req.QueryStringParameters["loyaltyId"]
	return model.RequestWrapper{}, nil
}

func (u *SearchProfile) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Log("Validating request")
	return nil
}

func (u *SearchProfile) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	profiles := []customerprofiles.Profile{}
	var err error
	if req.SearchRq.LastName != "" {
		profiles, err = u.reg.Accp.SearchProfiles(SEARCH_KEY_LAST_NAME, []string{req.SearchRq.LastName})
	}
	if req.SearchRq.LoyaltyID != "" {
		profiles, err = u.reg.Accp.SearchProfiles(SEARCH_KEY_ACCOUNT_NUMBER, []string{req.SearchRq.LoyaltyID})
	}
	if req.SearchRq.Phone != "" {
		profiles, err = u.reg.Accp.SearchProfiles(SEARCH_KEY_PHONE, []string{req.SearchRq.Phone})
	}
	if req.SearchRq.Email != "" {
		profiles, err = u.reg.Accp.SearchProfiles(SEARCH_KEY_EMAIL, []string{req.SearchRq.Email})
	}
	if err != nil {
		return model.ResponseWrapper{}, err
	}
	return model.ResponseWrapper{Profiles: profilesToTravellers(profiles)}, nil
}

func (u *SearchProfile) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}
