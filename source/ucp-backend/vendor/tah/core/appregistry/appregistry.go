package appregistry

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/appregistry"
)

type Config struct {
	Client *appregistry.AppRegistry
	Region string
}

type ApplicationSummary struct {
	Arn            string
	CreationTime   time.Time
	Description    string
	Id             string
	LastUpdateTime time.Time
	Name           string
}

type Application struct {
	Arn            string
	CreationTime   time.Time
	Description    string
	Id             string
	LastUpdateTime time.Time
	Name           string
	Tags           map[string]string
}

func Init(region string) Config {
	mySession := session.Must(session.NewSession())
	cfg := aws.NewConfig().WithRegion(region)
	svc := appregistry.New(mySession, cfg)
	return Config{
		Client: svc,
		Region: region,
	}
}

func (c Config) ListApplications() ([]ApplicationSummary, error) {
	input := &appregistry.ListApplicationsInput{}
	output, err := c.Client.ListApplications(input)
	applicationSummaryList := []ApplicationSummary{}
	for _, application := range output.Applications {
		app := ApplicationSummary{
			Arn:            *application.Arn,
			CreationTime:   *application.CreationTime,
			Description:    *application.Description,
			Id:             *application.Id,
			LastUpdateTime: *application.LastUpdateTime,
			Name:           *application.Name,
		}
		applicationSummaryList = append(applicationSummaryList, app)
	}
	return applicationSummaryList, err
}

func (c Config) CreateApplication(name, description string, tags map[string]string) (Application, error) {
	tagsWithPointer := make(map[string]*string)
	for key, value := range tags {
		tagsWithPointer[key] = &value
	}
	input := &appregistry.CreateApplicationInput{
		Name:        &name,
		Description: &description,
		Tags:        tagsWithPointer,
	}
	output, err := c.Client.CreateApplication(input)
	if err != nil {
		return Application{}, err
	}
	tagsWithoutPointer := make(map[string]string)
	for key, value := range output.Application.Tags {
		tagsWithoutPointer[key] = *value
	}
	app := Application{
		Arn:            *output.Application.Arn,
		CreationTime:   *output.Application.CreationTime,
		Description:    *output.Application.Description,
		Id:             *output.Application.Id,
		LastUpdateTime: *output.Application.LastUpdateTime,
		Name:           *output.Application.Name,
		Tags:           tagsWithoutPointer,
	}
	return app, nil
}

// DeleteApplication can use the application's Name or ID as the identifier
func (c Config) DeleteApplication(identifier string) error {
	input := &appregistry.DeleteApplicationInput{
		Application: &identifier,
	}
	_, err := c.Client.DeleteApplication(input)
	return err
}
