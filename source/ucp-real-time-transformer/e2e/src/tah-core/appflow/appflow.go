package appflow

import (
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	appflowSdk "github.com/aws/aws-sdk-go/service/appflow"
)

type Config struct {
	Client *appflowSdk.Appflow
}

type Flow struct {
	Name           string
	Description    string
	Status         string
	StatusMessage  string
	LastRun        time.Time
	LastRunStatus  string
	LastRunMessage string
	SourceType     string
	SouceDetails   string
	TargetType     string
	TargetDetails  string
	Trigger        string
}

func Init() Config {
	session := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	return Config{
		Client: appflowSdk.New(session),
	}

}

func (c Config) GetFlow(flowName string) (Flow, error) {
	input := &appflowSdk.DescribeFlowInput{
		FlowName: aws.String(flowName),
	}
	out, err := c.Client.DescribeFlow(input)
	if err != nil {
		return Flow{}, err
	}
	flow := Flow{
		Name:          *out.FlowName,
		Description:   *out.Description,
		Status:        *out.FlowStatus,
		SourceType:    *out.SourceFlowConfig.ConnectorType,
		SouceDetails:  buildSourceDetails(*out.SourceFlowConfig),
		TargetType:    *out.DestinationFlowConfigList[0].ConnectorType,
		TargetDetails: buildTargetDetails(*out.DestinationFlowConfigList[0]),
		Trigger:       buildTriggerDescription(*out.TriggerConfig),
	}
	if out.FlowStatusMessage != nil {
		flow.StatusMessage = *out.FlowStatusMessage
	}
	if out.LastRunExecutionDetails != nil {
		flow.LastRun = *out.LastRunExecutionDetails.MostRecentExecutionTime
		flow.LastRunStatus = *out.LastRunExecutionDetails.MostRecentExecutionStatus
		if out.LastRunExecutionDetails.MostRecentExecutionMessage != nil {
			flow.LastRunMessage = *out.LastRunExecutionDetails.MostRecentExecutionMessage
		}
	}
	return flow, nil
}

func buildTriggerDescription(cfg appflowSdk.TriggerConfig) string {
	if *cfg.TriggerType == "Scheduled" {
		return "Scheduled on " + (*(cfg.TriggerProperties.Scheduled.ScheduleExpression)) + " starting " + (*(cfg.TriggerProperties.Scheduled.ScheduleStartTime)).Format("Mon Jan 2 15:04:05")
	}
	return *cfg.TriggerType
}

func buildSourceDetails(sfc appflowSdk.SourceFlowConfig) string {
	if *sfc.ConnectorType == "S3" {
		props := sfc.SourceConnectorProperties.S3
		if props != nil {
			return *props.BucketName + "/" + *props.BucketPrefix
		}
	}
	return ""
}

func buildTargetDetails(dfc appflowSdk.DestinationFlowConfig) string {
	if *dfc.ConnectorType == "CustomerProfiles" {
		props := (*dfc.DestinationConnectorProperties).CustomerProfiles
		if props != nil {
			return *props.DomainName + "/" + *props.ObjectTypeName
		}
	}
	return ""
}

func (c Config) GetFlows(names []string) ([]Flow, error) {
	log.Printf("[core][appflow] Getting all Flows")
	var err error
	flows := []Flow{}
	var wg sync.WaitGroup
	wg.Add(len(names))
	for _, name := range names {
		go func(flowName string) {
			flow, err2 := c.GetFlow(flowName)
			if err2 != nil {
				err = err2
			}
			//use thread safe datastructure
			flows = append(flows, flow)
			wg.Done()
		}(name)
	}
	wg.Wait()
	log.Printf("[core][appflow] Flows: %+v", flows)
	return flows, err
}
