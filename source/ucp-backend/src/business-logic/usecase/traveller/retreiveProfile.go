package traveller

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"tah/core/core"
	"tah/core/customerprofiles"
	model "tah/ucp/src/business-logic/model/common"
	travModel "tah/ucp/src/business-logic/model/traveller"
	"tah/ucp/src/business-logic/usecase/registry"
	"tah/ucp/src/business-logic/utils"

	"github.com/aws/aws-lambda-go/events"
)

const (
	QUERY_PARAM_PAGES      string = "pages"
	QUERY_PARAM_PAGE_SIZES string = "pageSizes"
	QUERY_PARAM_OBJECTS    string = "objects"
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
	pages := utils.ParseIntArray(strings.Split(req.QueryStringParameters[QUERY_PARAM_PAGES], ","))
	pageSizes := utils.ParseIntArray(strings.Split(req.QueryStringParameters[QUERY_PARAM_PAGE_SIZES], ","))
	objects := strings.Split(req.QueryStringParameters[QUERY_PARAM_OBJECTS], ",")

	paginationOptions, err := buildPaginationOptions(pages, pageSizes, objects)
	if err != nil {
		return model.RequestWrapper{}, err
	}
	rw := model.RequestWrapper{
		ID:         req.PathParameters["id"],
		Pagination: paginationOptions,
	}
	return rw, nil
}

func buildPaginationOptions(pages, pageSizes []int, objects []string) ([]model.PaginationOptions, error) {
	if len(pages) != len(pageSizes) || len(pages) != len(objects) || len(pageSizes) != len(objects) {
		return []model.PaginationOptions{}, errors.New("Invalid multi-object pagination options. number of query param mismatch betwen objects, pages and pageSizes")
	}
	pos := []model.PaginationOptions{}
	for i, obj := range objects {
		if obj != "" {
			pos = append(pos, model.PaginationOptions{
				Page:       pages[i],
				PageSize:   pageSizes[i],
				ObjectType: obj,
			})
		}
	}
	return pos, nil
}

func (u *RetreiveProfile) ValidateRequest(rq model.RequestWrapper) error {
	isValid := map[string]bool{}
	for _, obj := range COMBINED_PROFILE_OBJECT_TYPES {
		isValid[obj] = true
	}
	u.tx.Log("Validating request")
	for _, po := range rq.Pagination {
		if !isValid[po.ObjectType] {
			return errors.New(fmt.Sprintf("Invalid object type %v in pagination. Supported values are %v", po.ObjectType, COMBINED_PROFILE_OBJECT_TYPES))
		}
	}
	return nil
}

func (u *RetreiveProfile) Run(req model.RequestWrapper) (model.ResponseWrapper, error) {
	//we don't paginate at the ACCp level because we want to control the order sort items by timestamp
	profile, err := u.reg.Accp.GetProfile(req.ID, COMBINED_PROFILE_OBJECT_TYPES, []customerprofiles.PaginationOptions{})
	traveller := profileToTraveller(profile)
	traveller = paginate(traveller, req.Pagination)
	return model.ResponseWrapper{Profiles: []travModel.Traveller{traveller}, Matches: profileToMatches(profile)}, err
}

func paginate(traveller travModel.Traveller, pos []model.PaginationOptions) travModel.Traveller {
	//putting pagination options in a map to allow for qiuck access
	poMap := map[string]model.PaginationOptions{}
	for _, po := range pos {
		poMap[po.ObjectType] = po
	}
	traveller.AirBookingRecords = getPage(traveller.AirBookingRecords, poMap[OBJECT_TYPE_AIR_BOOKING]).([]travModel.AirBooking)
	traveller.AirLoyaltyRecords = getPage(traveller.AirLoyaltyRecords, poMap[OBJECT_TYPE_AIR_LOYALTY]).([]travModel.AirLoyalty)
	traveller.ClickstreamRecords = getPage(traveller.ClickstreamRecords, poMap[OBJECT_TYPE_CLICKSTREAM]).([]travModel.Clickstream)
	traveller.EmailHistoryRecords = getPage(traveller.EmailHistoryRecords, poMap[OBJECT_TYPE_EMAIL_HISTORY]).([]travModel.EmailHistory)
	traveller.HotelBookingRecords = getPage(traveller.HotelBookingRecords, poMap[OBJECT_TYPE_HOTEL_BOOKING]).([]travModel.HotelBooking)
	traveller.HotelLoyaltyRecords = getPage(traveller.HotelLoyaltyRecords, poMap[OBJECT_TYPE_HOTEL_LOYALTY]).([]travModel.HotelLoyalty)
	traveller.HotelStayRecords = getPage(traveller.HotelStayRecords, poMap[OBJECT_TYPE_HOTEL_STAY]).([]travModel.HotelStay)
	traveller.PhoneHistoryRecords = getPage(traveller.PhoneHistoryRecords, poMap[OBJECT_TYPE_PHONE_HISTORY]).([]travModel.PhoneHistory)
	return traveller
}

//TODO: move thos to core. also, this should be benchmarked for performance and fully unit tested
func getPage(arr interface{}, pagination model.PaginationOptions) interface{} {
	if pagination.PageSize == 0 {
		return arr
	}
	// Convert input array to a slice of reflect.Values
	arrValue := reflect.ValueOf(arr)
	slice := make([]reflect.Value, arrValue.Len())
	for i := 0; i < arrValue.Len(); i++ {
		slice[i] = arrValue.Index(i)
	}
	// Filter the slice based on the condition
	filtered := make([]reflect.Value, 0)
	from := pagination.Page * pagination.PageSize
	to := (pagination.Page + 1) * pagination.PageSize
	if to > len(slice) {
		to = len(slice)
	}
	filtered = slice[from:to]
	// Convert filtered slice to an array of the same type as the input array
	arrType := arrValue.Type().Elem()
	result := reflect.New(reflect.SliceOf(arrType)).Elem()
	for _, v := range filtered {
		result = reflect.Append(result, v)
	}
	return result.Interface()
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
	// TODO: Demo stub. to remove
	matches = append(matches, model.Match{
		ConfidenceScore: 0.99,
		ID:              "4caf3d602b84460db31650159dcff896",
		FirstName:       "Geoffroy",
		LastName:        "Rollat",
	})

	return matches
}
