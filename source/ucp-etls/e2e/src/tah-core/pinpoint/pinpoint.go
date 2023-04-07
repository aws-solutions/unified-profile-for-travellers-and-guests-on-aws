package pinpoint

import (
	"errors"
	"regexp"
	core "tah/core/core"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	pinpointSdk "github.com/aws/aws-sdk-go/service/pinpoint"
)

type PinpointConfig struct {
	Service       *pinpointSdk.Pinpoint
	ApplicationId string
	Tx            core.Transaction
	EmailFrom     string
}

type PinpointLocation struct {
	City       string
	Country    string
	Latitude   float64
	Longitude  float64
	PostalCode string
	Region     string
}

type PinpointEndpoint struct {
	Cfg            PinpointConfig
	Email          string
	Attributes     map[string][]string
	Metrics        map[string]float64
	Location       PinpointLocation
	UserId         string
	UserAttributes map[string][]string
}
type PinpointSegment struct {
	Cfg        PinpointConfig
	Name       string
	Id         string
	Attributes map[string][]string
}

type PinpointJourney struct {
	Id         string
	Cfg        PinpointConfig
	Name       string
	SegmentId  string
	Activities []PinpointJourneyActivity
}

type PinpointJourneyActivity struct {
	Cfg          PinpointConfig
	Id           string
	Description  string
	TemplateName string
	WaitUntil    *time.Time
}

type PinpointTemplate struct {
	Name        string
	Subject     string
	HtmlContent string
	TxtContent  string
}

func InitWithRegion(applicationId string, region string) PinpointConfig {
	mySession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region)},
	))
	// Create a Pinpoint client from just a session.
	svc := pinpointSdk.New(mySession)
	// Create DynamoDB client
	return PinpointConfig{
		Service:       svc,
		ApplicationId: applicationId,
		Tx:            core.NewTransaction("PINPOINT", ""),
	}
}

func Init(applicationId string) PinpointConfig {
	mySession := session.Must(session.NewSession())
	// Create a Pinpoint client from just a session.
	svc := pinpointSdk.New(mySession)
	// Create DynamoDB client
	return PinpointConfig{
		Service:       svc,
		ApplicationId: applicationId,
		Tx:            core.NewTransaction("PINPOINT", ""),
	}
}

func InitWithEmailFrom(applicationId string, emailFrom string) PinpointConfig {
	mySession := session.Must(session.NewSession())
	// Create a Pinpoint client from just a session.
	svc := pinpointSdk.New(mySession)
	// Create DynamoDB client
	return PinpointConfig{
		Service:       svc,
		ApplicationId: applicationId,
		EmailFrom:     emailFrom,
		Tx:            core.NewTransaction("PINPOINT", ""),
	}
}

func (pc PinpointConfig) SetTx(tx core.Transaction) {
	tx.LogPrefix = "PINPOINT"
	pc.Tx = tx
}

func (pc PinpointConfig) AddAttribute(email string, attributeKey string, attibuteValue string) error {
	ept, err := pc.GetEndpoint(email)
	if err != nil {
		return err
	}
	ept.Attributes[attributeKey] = append(ept.Attributes[attributeKey], attibuteValue)
	_, err = pc.CreateUpdateEndpoint(email, ept.UserId, ept.Location, ept.Attributes, ept.Metrics, ept.UserAttributes)
	return err
}

func (pc PinpointConfig) GetEndpoint(email string) (PinpointEndpoint, error) {
	input := pinpointSdk.GetEndpointInput{
		ApplicationId: aws.String(pc.ApplicationId),
		EndpointId:    aws.String(email),
	}
	res, err := pc.Service.GetEndpoint(&input)
	//pc.Tx.Log("GetEndpoint response: %+v", res)
	if err != nil {
		pc.Tx.Log("GetEndpoint ERROR: %+v", err)
		return PinpointEndpoint{}, err
	}
	endpoint := PinpointEndpoint{
		Cfg:   pc,
		Email: email,
	}
	if res.EndpointResponse.Location != nil {
		endpoint.Location = PinpointLocation{
			City:       *res.EndpointResponse.Location.City,
			Country:    *res.EndpointResponse.Location.Country,
			Latitude:   *res.EndpointResponse.Location.Latitude,
			Longitude:  *res.EndpointResponse.Location.Longitude,
			PostalCode: *res.EndpointResponse.Location.PostalCode,
			Region:     *res.EndpointResponse.Location.Region,
		}
	}
	if res.EndpointResponse.Attributes != nil {
		endpoint.Attributes = unformatAttributes(res.EndpointResponse.Attributes)
	}
	if res.EndpointResponse.User != nil {
		if res.EndpointResponse.User.UserId != nil {
			endpoint.UserId = *res.EndpointResponse.User.UserId
		}
		if res.EndpointResponse.User.UserAttributes != nil {
			endpoint.UserAttributes = unformatAttributes(res.EndpointResponse.User.UserAttributes)
		}
	}
	if res.EndpointResponse.Metrics != nil {
		endpoint.Metrics = unformatMetrics(res.EndpointResponse.Metrics)
	}

	return endpoint, nil
}

func (pc PinpointConfig) GetSegment(id string) (PinpointSegment, error) {
	input := pinpointSdk.GetSegmentInput{
		ApplicationId: aws.String(pc.ApplicationId),
		SegmentId:     aws.String(id),
	}
	res, err := pc.Service.GetSegment(&input)
	//pc.Tx.Log("GetSegment response: %+v", res)
	if err != nil {
		pc.Tx.Log("GetSegment ERROR: %+v", err)
		return PinpointSegment{}, err
	}
	segment := PinpointSegment{
		Cfg:  pc,
		Name: *res.SegmentResponse.Name,
		Id:   *res.SegmentResponse.Id,
	}
	return segment, nil
}

func (pc PinpointConfig) CreateUpdateEndpoint(email string, userId string, location PinpointLocation, attributes map[string][]string, metrics map[string]float64, userAttributes map[string][]string) (PinpointEndpoint, error) {
	input := pinpointSdk.UpdateEndpointInput{

		// ApplicationId is a required field
		ApplicationId: aws.String(pc.ApplicationId),

		// EndpointId is a required field
		EndpointId: aws.String(email),

		// Specifies the channel type and other settings for an endpoint.
		// EndpointRequest is a required field
		EndpointRequest: &pinpointSdk.EndpointRequest{
			Address:     aws.String(email),
			Attributes:  formatAttributes(attributes),
			ChannelType: aws.String("EMAIL"),
			// The demographic information for the endpoint, such as the time zone and platform.
			Demographic: &pinpointSdk.EndpointDemographic{},
			//Mon Jan 2 15:04:05 MST 2006
			EffectiveDate:  aws.String(time.Now().Format("2006-01-02T15:04:05Z")),
			EndpointStatus: aws.String("ACTIVE"),

			// The geographic information for the endpoint.
			Location: &pinpointSdk.EndpointLocation{
				City:       aws.String(location.City),
				Country:    aws.String(location.Country),
				Latitude:   aws.Float64(location.Latitude),
				Longitude:  aws.Float64(location.Longitude),
				PostalCode: aws.String(location.PostalCode),
				Region:     aws.String(location.Region),
			},
			// One or more custom metrics that your app reports to Amazon Pinpoint for the
			// endpoint.
			Metrics: formatMetrics(metrics),

			// One or more custom attributes that describe the user who's associated with
			// the endpoint.
			User: &pinpointSdk.EndpointUser{
				UserId:         aws.String(userId),
				UserAttributes: formatAttributes(userAttributes),
			},
			// contains filtered or unexported fields
		},
	}
	//pc.Tx.Log("UpdateEndpoint request: %+v", input)
	_, err := pc.Service.UpdateEndpoint(&input)
	//pc.Tx.Log("UpdateEndpoint response: %+v", res)
	if err != nil {
		pc.Tx.Log("UpdateEndpoint ERROR: %+v", err)
	}
	endpoint := PinpointEndpoint{
		Cfg:   pc,
		Email: email,
	}
	return endpoint, err
}

func formatAttributes(attrs map[string][]string) map[string][]*string {
	formated := map[string][]*string{}
	for key, vals := range attrs {
		formated[key] = []*string{}
		for _, val := range vals {
			formated[key] = append(formated[key], aws.String(val))
		}
	}
	return formated
}

func unformatAttributes(attrs map[string][]*string) map[string][]string {
	unformated := map[string][]string{}
	for key, vals := range attrs {
		unformated[key] = []string{}
		for _, val := range vals {
			unformated[key] = append(unformated[key], *val)
		}
	}
	return unformated
}

func formatMetrics(metrics map[string]float64) map[string]*float64 {
	formated := map[string]*float64{}
	for key, val := range metrics {
		formated[key] = aws.Float64(val)
	}
	return formated
}

func unformatMetrics(metrics map[string]*float64) map[string]float64 {
	unformated := map[string]float64{}
	for key, val := range metrics {
		unformated[key] = *val
	}
	return unformated
}

func (pc PinpointConfig) SendEmail(to string, subject string, htmlMsg string, txtMsg string) error {
	emailFrom := "adm.cloudrack@gmail.com"
	if pc.EmailFrom == "" {
		emailFrom = pc.EmailFrom
	}
	input := pinpointSdk.SendMessagesInput{
		ApplicationId: aws.String(pc.ApplicationId),
		MessageRequest: &pinpointSdk.MessageRequest{
			Addresses: map[string]*pinpointSdk.AddressConfiguration{
				to: &pinpointSdk.AddressConfiguration{
					//"GCM"|"APNS"|"APNS_SANDBOX"|"APNS_VOIP"|"APNS_VOIP_SANDBOX"|"ADM"|"SMS"|"VOICE"|"EMAIL"|"BAIDU"|"CUSTOM"
					ChannelType: aws.String("EMAIL"),
				},
			},
			MessageConfiguration: &pinpointSdk.DirectMessageConfiguration{
				EmailMessage: &pinpointSdk.EmailMessage{
					FromAddress: aws.String(emailFrom),
					SimpleEmail: &pinpointSdk.SimpleEmail{
						Subject: &pinpointSdk.SimpleEmailPart{
							Charset: aws.String("utf-8"),
							Data:    aws.String(subject),
						},
						HtmlPart: &pinpointSdk.SimpleEmailPart{
							Charset: aws.String("utf-8"),
							Data:    aws.String(htmlMsg),
						},
						TextPart: &pinpointSdk.SimpleEmailPart{
							Charset: aws.String("utf-8"),
							Data:    aws.String(txtMsg),
						},
					},
					Substitutions: map[string][]*string{},
				},
			},
			TraceId: aws.String("cloudrack-email-test"),
		},
	}
	//pc.Tx.Log("SendMessages request: %+v", input)
	res, err := pc.Service.SendMessages(&input)
	//pc.Tx.Log("SendMessages response: %+v", res)
	if err != nil {
		pc.Tx.Log("SendMessages ERROR: %+v", err)
	}
	if res.MessageResponse.Result[to] != nil && (*res.MessageResponse.Result[to].DeliveryStatus != "SUCCESSFUL") {
		return errors.New("Error while sending email. Delivery Status: " + *res.MessageResponse.Result[to].DeliveryStatus)
	}
	return err
}

func (pc PinpointConfig) SendEmailFromTemplate(to string, templateName string, templateData map[string]string) error {
	emailFrom := "adm.cloudrack@gmail.com"
	if pc.EmailFrom != "" {
		emailFrom = pc.EmailFrom
	}
	attributes := map[string][]*string{}
	for key, val := range templateData {
		attributes[key] = []*string{aws.String(val)}
	}
	input := pinpointSdk.SendMessagesInput{
		ApplicationId: aws.String(pc.ApplicationId),
		MessageRequest: &pinpointSdk.MessageRequest{
			Addresses: map[string]*pinpointSdk.AddressConfiguration{
				to: &pinpointSdk.AddressConfiguration{
					//"GCM"|"APNS"|"APNS_SANDBOX"|"APNS_VOIP"|"APNS_VOIP_SANDBOX"|"ADM"|"SMS"|"VOICE"|"EMAIL"|"BAIDU"|"CUSTOM"
					ChannelType: aws.String("EMAIL"),
				},
			},
			MessageConfiguration: &pinpointSdk.DirectMessageConfiguration{
				EmailMessage: &pinpointSdk.EmailMessage{
					FromAddress:   aws.String(emailFrom),
					Substitutions: attributes,
				},
			},
			TemplateConfiguration: &pinpointSdk.TemplateConfiguration{
				EmailTemplate: &pinpointSdk.Template{
					Name:    aws.String(templateName),
					Version: aws.String("1"),
				},
			},
			TraceId: aws.String("cloudrack-email-test"),
		},
	}
	pc.Tx.Log("SendMessages request: %+v", input)
	res, err := pc.Service.SendMessages(&input)
	//pc.Tx.Log("SendMessages response: %+v", res)
	if err != nil {
		pc.Tx.Log("SendMessages ERROR: %+v", err)
		return err
	}
	if res.MessageResponse.Result[to] != nil && (*res.MessageResponse.Result[to].DeliveryStatus != "SUCCESSFUL") {
		return errors.New("Error while sendinng email. Delivery Status: " + *res.MessageResponse.Result[to].DeliveryStatus)
	}
	return err
}

func (pc PinpointConfig) CreateSegment(name string, tags map[string]string, sourceSegmentId string, attributeValues map[string]string) (PinpointSegment, error) {

	attributes := map[string]*pinpointSdk.AttributeDimension{}
	for key, val := range attributeValues {
		attributes[key] = &pinpointSdk.AttributeDimension{
			AttributeType: aws.String("INCLUSIVE"),
			Values:        []*string{aws.String(val)},
		}
	}
	sourceSegments := []*pinpointSdk.SegmentReference{}
	if sourceSegmentId != "" {
		sourceSegments = append(sourceSegments, &pinpointSdk.SegmentReference{
			Id: aws.String(sourceSegmentId),
		})
	}
	input := &pinpointSdk.CreateSegmentInput{
		ApplicationId: aws.String(pc.ApplicationId),
		WriteSegmentRequest: &pinpointSdk.WriteSegmentRequest{
			Name: aws.String(name),
			SegmentGroups: &pinpointSdk.SegmentGroupList{
				Groups: []*pinpointSdk.SegmentGroup{
					&pinpointSdk.SegmentGroup{
						SourceSegments: sourceSegments,
						Dimensions: []*pinpointSdk.SegmentDimensions{&pinpointSdk.SegmentDimensions{
							Attributes: attributes,
						},
						},
					},
				},
			},
			Tags: formatTags(tags),
		},
	}
	//pc.Tx.Log("CreateSegment request: %+v", input)
	res, err := pc.Service.CreateSegment(input)
	//pc.Tx.Log("CreateSegment response: %+v", res)
	if err != nil {
		pc.Tx.Log("CreateSegment ERROR: %+v", err)
		return PinpointSegment{}, err
	}
	segment := PinpointSegment{
		Name: *res.SegmentResponse.Name,
		Id:   *res.SegmentResponse.Id,
	}
	return segment, nil
}

func (pc PinpointConfig) CreateJourney(startDate time.Time, endDate time.Time, journeyName string, segmentId string, activities []PinpointJourneyActivity) (PinpointJourney, error) {
	valErr := pc.validateActivities(activities)
	if valErr != nil {
		return PinpointJourney{}, valErr
	}
	input := &pinpointSdk.CreateJourneyInput{
		ApplicationId: aws.String(pc.ApplicationId),
		WriteJourneyRequest: &pinpointSdk.WriteJourneyRequest{
			Activities:       pc.buildActivityMap(activities),
			CreationDate:     aws.String(time.Now().Format("2006-01-02T15:04:05Z")),
			LastModifiedDate: aws.String(time.Now().Format("2006-01-02T15:04:05Z")),
			Name:             aws.String(journeyName),
			RefreshFrequency: aws.String("PT1H"),
			StartActivity:    aws.String(activities[0].Id),
			StartCondition: &pinpointSdk.StartCondition{
				SegmentStartCondition: &pinpointSdk.SegmentCondition{
					SegmentId: aws.String(segmentId),
				},
			},
			State: aws.String("ACTIVE"),
			Schedule: &pinpointSdk.JourneySchedule{
				EndTime:   &endDate,
				StartTime: &startDate,
				//fix that
				Timezone: aws.String("UTC"),
			},
		},
	}
	//pc.Tx.Log("CreateJourney request: %+v", input)
	res, err := pc.Service.CreateJourney(input)
	//pc.Tx.Log("CreateJourney response: %+v", res)
	if err != nil {
		pc.Tx.Log("CreateJourney ERROR: %+v", err)
		return PinpointJourney{}, err
	}

	journey := PinpointJourney{
		Id:         *res.JourneyResponse.Id,
		Name:       journeyName,
		SegmentId:  segmentId,
		Activities: activities,
	}
	return journey, nil
}

func (pc PinpointConfig) validateActivities(activities []PinpointJourneyActivity) error {
	var isAlphaNumberic = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString
	for _, activity := range activities {
		if !isAlphaNumberic(activity.Id) {
			//This is a pinpoint limitation (2020-05-26)
			return errors.New("journey activities must have alphanumberic ID")
		}
	}
	return nil
}

//For thsi function to set the right time for activity, WaitTime is expected to be in hotel timezone
func (pc PinpointConfig) buildActivityMap(activities []PinpointJourneyActivity) map[string]*pinpointSdk.Activity {
	activityMap := map[string]*pinpointSdk.Activity{}
	nextActivityId := ""
	utc, _ := time.LoadLocation("UTC")
	for i, activity := range activities {
		if i < len(activities)-1 {
			nextActivityId = activities[i+1].Id
		} else {
			nextActivityId = ""
		}
		if activity.WaitUntil != nil {
			pc.Tx.Log("[CORE][PINPOINT][JOURNEY] Creating wait activity with WaitTime: %+v", activity.WaitUntil)
			activityMap[activity.Id] = &pinpointSdk.Activity{
				Description: aws.String(activity.Description),
				Wait: &pinpointSdk.WaitActivity{
					WaitTime: &pinpointSdk.WaitTime{
						WaitUntil: aws.String(activity.WaitUntil.In(utc).Format("2006-01-02T15:04:05Z")),
					},
				},
			}
			if nextActivityId != "" {
				activityMap[activity.Id].Wait.NextActivity = aws.String(nextActivityId)
			}
		} else {
			activityMap[activity.Id] = &pinpointSdk.Activity{
				Description: aws.String(activity.Description),
				EMAIL: &pinpointSdk.EmailMessageActivity{
					TemplateName: aws.String(activity.TemplateName),
				},
			}
			if nextActivityId != "" {
				activityMap[activity.Id].EMAIL.NextActivity = aws.String(nextActivityId)
			}
		}

	}
	return activityMap
}

func (pc PinpointConfig) DeleteSegment(id string) error {
	input := &pinpointSdk.DeleteSegmentInput{
		ApplicationId: aws.String(pc.ApplicationId),
		SegmentId:     aws.String(id),
	}
	//pc.Tx.Log("DeleteSegment request: %+v", input)
	_, err := pc.Service.DeleteSegment(input)
	//pc.Tx.Log("DeleteSegment response: %+v", res)
	if err != nil {
		pc.Tx.Log("DeleteSegment ERROR: %+v", err)
		return err
	}
	return nil
}

func (pc PinpointConfig) DeleteEndpoint(endpointId string) error {
	input := &pinpointSdk.DeleteEndpointInput{
		ApplicationId: aws.String(pc.ApplicationId),
		EndpointId:    aws.String(endpointId),
	}
	//pc.Tx.Log("DeleteEndpoint request: %+v", input)
	_, err := pc.Service.DeleteEndpoint(input)
	//pc.Tx.Log("DeleteEndpoint response: %+v", res)
	if err != nil {
		pc.Tx.Log("DeleteEndpoint ERROR: %+v", err)
		return err
	}
	return nil
}

func (pc PinpointConfig) DeleteJourney(journeyId string) error {
	input := &pinpointSdk.DeleteJourneyInput{
		ApplicationId: aws.String(pc.ApplicationId),
		JourneyId:     aws.String(journeyId),
	}
	//pc.Tx.Log("DeleteJourney request: %+v", input)
	_, err := pc.Service.DeleteJourney(input)
	//pc.Tx.Log("DeleteJourney response: %+v", res)
	if err != nil {
		pc.Tx.Log("DeleteJourney ERROR: %+v", err)
		return err
	}
	return nil
}

func (pc PinpointConfig) CreateTemplate(name string, subject string, html string, txt string) error {
	input := &pinpointSdk.CreateEmailTemplateInput{
		EmailTemplateRequest: &pinpointSdk.EmailTemplateRequest{
			Subject:  aws.String(subject),
			HtmlPart: aws.String(html),
			TextPart: aws.String(txt),
		},
		TemplateName: aws.String(name),
	}
	//pc.Tx.Log("CreateTemplate request: %+v", input)
	_, err := pc.Service.CreateEmailTemplate(input)
	//pc.Tx.Log("CreateTemplate response: %+v", res)
	if err != nil {
		pc.Tx.Log("CreateTemplate ERROR: %+v", err)
		return err
	}
	return nil
}

func (pc PinpointConfig) UpdateTemplate(name string, subject string, html string, txt string) error {
	input := &pinpointSdk.UpdateEmailTemplateInput{
		EmailTemplateRequest: &pinpointSdk.EmailTemplateRequest{
			Subject:  aws.String(subject),
			HtmlPart: aws.String(html),
			TextPart: aws.String(txt),
		},
		TemplateName: aws.String(name),
	}
	//pc.Tx.Log("CreateTemplate request: %+v", input)
	_, err := pc.Service.UpdateEmailTemplate(input)
	//pc.Tx.Log("CreateTemplate response: %+v", res)
	if err != nil {
		pc.Tx.Log("UpdateTemplate ERROR: %+v", err)
		return err
	}
	return nil
}

func (pc PinpointConfig) GetTemplate(name string) (PinpointTemplate, error) {
	input := &pinpointSdk.GetEmailTemplateInput{
		TemplateName: aws.String(name),
	}
	//pc.Tx.Log("GetTemplate request: %+v", input)
	res, err := pc.Service.GetEmailTemplate(input)
	//pc.Tx.Log("GetTemplate response: %+v", res)
	if err != nil {
		pc.Tx.Log("GetTemplate ERROR: %+v", err)
		return PinpointTemplate{}, err
	}
	pt := PinpointTemplate{
		Name: *res.EmailTemplateResponse.TemplateName,
	}
	if res.EmailTemplateResponse.Subject != nil {
		pt.Subject = *res.EmailTemplateResponse.Subject
	}
	if res.EmailTemplateResponse.HtmlPart != nil {
		pt.HtmlContent = *res.EmailTemplateResponse.HtmlPart
	}
	if res.EmailTemplateResponse.TextPart != nil {
		pt.TxtContent = *res.EmailTemplateResponse.TextPart
	}
	return pt, nil
}

func (pc PinpointConfig) DeleteTemplate(name string) error {
	input := &pinpointSdk.DeleteEmailTemplateInput{
		TemplateName: aws.String(name),
	}
	//pc.Tx.Log("DeleteTemplate request: %+v", input)
	_, err := pc.Service.DeleteEmailTemplate(input)
	//pc.Tx.Log("DeleteTemplate response: %+v", res)
	if err != nil {
		pc.Tx.Log("DeleteTemplate ERROR: %+v", err)
		return err
	}
	return nil
}

func formatTags(attrs map[string]string) map[string]*string {
	formated := map[string]*string{}
	for key, tag := range attrs {
		formated[key] = aws.String(tag)
	}
	return formated
}
