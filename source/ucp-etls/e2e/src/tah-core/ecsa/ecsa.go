package ecsa

import (
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsecs "github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ssm"
)

type Config struct {
	Client     *awsecs.ECS //sdk client to make call to the AWS API
	ClusterArn string
	SsmClient  *ssm.SSM
	EcsaRole   string
}

type Instance struct {
	Arn                    string
	AgentConnected         bool
	AgentUpdateStatus      string
	OsType                 string
	Status                 string
	HealthStatus           string
	HeathStatusLastUpdated time.Time
	RunningTask            int64
}

type Task struct {
	// The attributes of the task
	Attributes []TaskAttribute `json:"Attributes"`
	// The ARN of the cluster that hosts the task.
	ClusterArn string `json:"clusterArn"`
	// The connectivity status of a task.
	Connectivity string `json:"connectivity"`
	// The Unix timestamp for the time when the task last went into CONNECTED status.
	ConnectivityAt time.Time `json:"connectivityAt"`
	// The ARN of the container instances that host the task.
	ContainerInstanceArn string `json:"containerInstanceArn"`
	Cpu                  string `json:"cpu"`
	// The Unix timestamp for the time when the task was created. More specifically,
	// it's for the time when the task entered the PENDING state.
	CreatedAt time.Time `json:"createdAt"`
	// The desired status of the task. For more information, see Task Lifecycle
	// (https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-lifecycle.html).
	DesiredStatus string `json:"desiredStatus"`
	// The Unix timestamp for the time when the task execution stopped.
	ExecutionStoppedAt time.Time `json:"executionStoppedAt"`
	HealthStatus       string    `json:"healthStatus"`
	// The last known status for the task. For more information, see Task Lifecycle
	// (https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-lifecycle.html).
	LastStatus string `json:"lastStatus"`
	// The infrastructure where your task runs on. For more information, see Amazon
	// ECS launch types (https://docs.aws.amazon.com/AmazonECS/latest/developerguide/launch_types.html)
	// in the Amazon Elastic Container Service Developer Guide.
	LaunchType string `json:"launchType"`
	// The amount of memory (in MiB) that the task uses as expressed in a task definition.
	Memory string `json:"memory"`
	// The Unix timestamp for the time when the container image pull began.
	PullStartedAt time.Time `json:"pullStartedAt"`
	// The Unix timestamp for the time when the container image pull completed.
	PullStoppedAt time.Time `json:"pullStoppedAt"`
	// The Unix timestamp for the time when the task started. More specifically,
	// it's for the time when the task transitioned from the PENDING state to the
	// RUNNING state.
	StartedAt time.Time `json:"startedAt"`
	// The tag specified when a task is started. If an Amazon ECS service started
	// the task, the startedBy parameter contains the deployment ID of that service.
	StartedBy string `json:"startedBy"`
	// The stop code indicating why a task was stopped. The stoppedReason might
	// contain additional details.
	StopCode string `json:"stopCode"`
	// The Unix timestamp for the time when the task was stopped. More specifically,
	// it's for the time when the task transitioned from the RUNNING state to the
	// STOPPED state.
	StoppedAt time.Time `json:"stoppedAt"`
	// The reason that the task was stopped.
	StoppedReason string `json:"stoppedReason"`
	// The Unix timestamp for the time when the task stops. More specifically, it's
	// for the time when the task transitions from the RUNNING state to STOPPED.
	StoppingAt time.Time `json:"stoppingAt"`
	Tags       []string  `json:"tags"`
	// The Amazon Resource Name (ARN) of the task.
	TaskArn string `json:"taskArn"`
	// The ARN of the task definition that creates the task.
	TaskDefinitionArn string `json:"taskDefinitionArn"`
	// The version counter for the task. Every time a task experiences a change
	// that starts a CloudWatch event, the version counter is incremented. If you
	// replicate your Amazon ECS task state with CloudWatch Events, you can compare
	// the version of a task reported by the Amazon ECS API actions with the version
	// reported in CloudWatch Events for the task (inside the detail object) to
	// verify that the version in your event stream is current.
	Version int64 `json:"version"`
}

type TaskAttribute struct {
	Name  string
	Value string
}

func Init(clusterArn string, ecsaRole string) Config {
	// Initialize a session that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials
	// and region from the shared configuration file ~/.aws/config.
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	// Create client
	return Config{
		Client:     awsecs.New(sess),
		ClusterArn: clusterArn,
		SsmClient:  ssm.New(sess),
		EcsaRole:   ecsaRole,
	}
}

func (cfg Config) ListContainerInstances() ([]Instance, error) {
	in := &awsecs.ListContainerInstancesInput{
		Cluster: aws.String(cfg.ClusterArn),
	}
	out, err := cfg.Client.ListContainerInstances(in)
	if err != nil {
		log.Printf("[core][listContainerInstances] error returned listing instances: %+v", err)
		return []Instance{}, err
	}
	log.Printf("[core][listContainerInstances] arn returned: %+v", out)
	list := []Instance{}
	//TODO: do this in parallele an bulk
	out2, err2 := cfg.Client.DescribeContainerInstances(&awsecs.DescribeContainerInstancesInput{
		Cluster:            aws.String(cfg.ClusterArn),
		ContainerInstances: out.ContainerInstanceArns,
	})
	if err2 != nil {
		return []Instance{}, err2
	}
	for _, instance := range out2.ContainerInstances {
		log.Printf("[core][listContainerInstances] container instance returned: %+v", instance)
		in := Instance{
			Arn:    *instance.ContainerInstanceArn,
			OsType: getOsType(instance.Attributes),
		}
		if instance.AgentConnected != nil {
			in.AgentConnected = *instance.AgentConnected
		}
		if instance.AgentUpdateStatus != nil {
			in.AgentUpdateStatus = *instance.AgentUpdateStatus
		}
		if instance.Status != nil {
			in.Status = *instance.Status
		}
		if instance.RunningTasksCount != nil {
			in.RunningTask = *instance.RunningTasksCount
		}
		list = append(list, in)
	}
	return list, nil
}

func getOsType(attrs []*awsecs.Attribute) string {
	for _, attr := range attrs {
		if *attr.Name == "ecs.os-type" && attr.Value != nil {
			return *attr.Value
		}
	}
	return ""
}

func (cfg Config) InstanceExists(arn string) bool {
	//to be implemented
	return true
}

func (cfg Config) Run(instanceArn string) (Task, error) {
	in := &awsecs.StartTaskInput{
		Cluster:            aws.String(cfg.ClusterArn),
		ContainerInstances: []*string{aws.String(instanceArn)},
	}
	out, err := cfg.Client.StartTask(in)
	if err != nil {
		return Task{}, err
	}
	if len(out.Tasks) == 0 {
		return Task{}, errors.New("No task started")
	}
	task := out.Tasks[0]
	return Task{
		Attributes:           []TaskAttribute{},
		ClusterArn:           *task.ClusterArn,
		Connectivity:         *task.Connectivity,
		ConnectivityAt:       *task.ConnectivityAt,
		ContainerInstanceArn: *task.ContainerInstanceArn,
		Cpu:                  *task.Cpu,
		CreatedAt:            *task.CreatedAt,
		DesiredStatus:        *task.DesiredStatus,
		ExecutionStoppedAt:   *task.ExecutionStoppedAt,
		HealthStatus:         *task.HealthStatus,
		LastStatus:           *task.LastStatus,
		LaunchType:           *task.LaunchType,
		Memory:               *task.Memory,
		PullStartedAt:        *task.PullStartedAt,
		PullStoppedAt:        *task.PullStoppedAt,
		StartedAt:            *task.StartedAt,
		StartedBy:            *task.StartedBy,
		StopCode:             *task.StopCode,
		StoppedAt:            *task.StoppedAt,
		StoppedReason:        *task.StoppedReason,
		StoppingAt:           *task.StoppingAt,
		Tags:                 []string{},
		TaskArn:              *task.TaskArn,
		TaskDefinitionArn:    *task.TaskArn,
		Version:              *task.Version,
	}, nil
}

func (cfg Config) UpdateDesiredTaskCount(service string, desired int64) error {
	in := &awsecs.UpdateServiceInput{
		Cluster:      aws.String(cfg.ClusterArn),
		DesiredCount: aws.Int64(desired),
		Service:      aws.String(service),
	}
	_, err := cfg.Client.UpdateService(in)
	if err != nil {
		return err
	}
	return nil
}

func (cfg Config) GenerateActivationCode() (string, string, error) {
	log.Printf("[core][listContainerInstances] Generating activation codes for role: %+v", cfg.EcsaRole)
	in := &ssm.CreateActivationInput{
		IamRole: aws.String(cfg.EcsaRole),
	}
	out, err := cfg.SsmClient.CreateActivation(in)
	if err != nil {
		return "", "", err
	}
	return *out.ActivationCode, *out.ActivationId, nil
}
