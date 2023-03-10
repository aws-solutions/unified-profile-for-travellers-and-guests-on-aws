package traveller

import (
	"tah/core/core"
	"tah/core/customerprofiles"
	model "tah/ucp/src/business-logic/model/common"
	travModel "tah/ucp/src/business-logic/model/traveller"
	"tah/ucp/src/business-logic/usecase/registry"

	"github.com/aws/aws-lambda-go/events"
)

type RetreiveProfile struct {
	name string
	tx   *core.Transaction
	reg  *registry.Registry
}

func NewRetreiveProfile() *RetreiveProfile {
	return &RetreiveProfile{name: "RetreiveProfile"}
}

func (u *RetreiveProfile) Name() string {
	return u.name
}
func (u *RetreiveProfile) Tx() core.Transaction {
	return *u.tx
}
func (u *RetreiveProfile) SetTx(tx *core.Transaction) {
	u.tx = tx
}
func (u *RetreiveProfile) SetRegistry(reg *registry.Registry) {
	u.reg = reg
}
func (u *RetreiveProfile) Registry() *registry.Registry {
	return u.reg
}

func (u *RetreiveProfile) CreateRequest(req events.APIGatewayProxyRequest) (model.RequestWrapper, error) {
	rw := model.RequestWrapper{
		ID: req.PathParameters["id"],
	}
	return rw, nil
}

func (u *RetreiveProfile) ValidateRequest(rq model.RequestWrapper) error {
	u.tx.Log("Validating request")
	return nil
}

func (u *RetreiveProfile) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	profile, err := u.reg.Accp.GetProfile(req.ID)
	return model.ResponseWrapper{Profiles: []travModel.Traveller{profileToTraveller(profile)}, Matches: profileToMatches(profile)}, err
}

func (u *RetreiveProfile) CreateResponse(res model.ResponseWrapper) (events.APIGatewayProxyResponse, error) {
	return registry.CreateResponse(u, res)
}

func profileToMatches(profile customerprofiles.Profile) []model.Match {
	matches := []model.Match{}
	for _, m := range profile.Matches {
		matches = append(matches, model.Match{
			ConfidenceScore: m.ConfidenceScore,
			ID:              m.ProfileID,
			FirstName:       m.FirstName,
			LastName:        m.LastName,
			BirthDate:       m.BirthDate,
			PhoneNumber:     m.PhoneNumber,
			EmailAddress:    m.EmailAddress,
		})
	}
	//Demo stub. to remove
	matches = append(matches, model.Match{
		ConfidenceScore: 0.99,
		ID:              "4caf3d602b84460db31650159dcff896",
		FirstName:       "Geoffroy",
		LastName:        "Rollat",
	})

	return matches
}
