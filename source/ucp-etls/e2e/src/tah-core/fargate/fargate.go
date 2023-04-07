package fargate

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	ec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	awsecs "github.com/aws/aws-sdk-go/service/ecs"
	iam "github.com/aws/aws-sdk-go/service/iam"
	sts "github.com/aws/aws-sdk-go/service/sts"
)

var EXECUSION_ROLE_POLICY_PREFIX = "AmazonECSTaskExecutionRolePolicy"
var ECS_EXECUTION_ROLE_TRUST_POLICY = `{
	"Version": "2012-10-17",
	"Statement": [
	  {
		"Sid": "",
		"Effect": "Allow",
		"Principal": {
		  "Service": "ecs-tasks.amazonaws.com"
		},
		"Action": "sts:AssumeRole"
	  }
	]
  }`
var ECS_EXECUTION_ROLE_POLICY = `{
	"Version": "2012-10-17",
	"Statement": [
	  {
		"Effect": "Allow",
		"Action": [
		  "ecr:GetAuthorizationToken",
		  "ecr:BatchCheckLayerAvailability",
		  "ecr:GetDownloadUrlForLayer",
		  "ecr:BatchGetImage",
		  "logs:CreateLogStream",
		  "logs:PutLogEvents"
		],
		"Resource": "*"
	  }
	]
  }`

type Config struct {
	Client    *awsecs.ECS //sdk client to make call to the AWS API
	IamClient *iam.IAM    //sdk client to make call to the AWS API
	STSClient *sts.STS    //sdk client to make call to the AWS API
	Ec2Client *ec2.EC2    //sdk client to make call to the AWS API
}

type Cluster struct {
	Name             string
	Arn              string
	ServiceName      string
	Size             int64
	Cpu              string
	Mem              string
	Spot             bool
	RoleArn          string //task role Arn
	ExecusionRoleArn string //execusion role arne
	VpcID            string
	Image            string
	Env              map[string]string
}

func Init() Config {
	// Initialize a session that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials
	// and region from the shared configuration file ~/.aws/config.
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	// Create client
	return Config{
		Client:    awsecs.New(sess),
		IamClient: iam.New(sess),
		STSClient: sts.New(sess),
		Ec2Client: ec2.New(sess),
	}
}

func (cfg Config) GetSubnets(vpcID string) ([]*string, error) {
	res := []*string{}
	input := &ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("vpc-id"),
				Values: []*string{
					aws.String(vpcID),
				},
			},
		},
	}
	result, err := cfg.Ec2Client.DescribeSubnets(input)
	if err != nil {
		return []*string{}, err
	}
	for _, subnet := range result.Subnets {
		res = append(res, subnet.SubnetId)
	}
	return res, nil
}

func (cfg Config) GetRegionAccount() (string, string, error) {
	identity, err := cfg.STSClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	region := cfg.Client.Config.Region
	return *region, *identity.Account, err
}

//Create a fargate cluster, execution role, task definitionand service
func (c Config) CreateCluster(cluster Cluster) (Cluster, error) {
	//IMPORTANT: we prefer passing the excusion role rather than creating it in ordre to avoid
	//having to give IAM creat policy and role rights tp the caller of this function
	///////////////////////////
	// Creating IAM resources for the cluster (execution role, policy...)
	//////////////////////////
	/*policyInput := &iam.CreatePolicyInput{
		Description:    aws.String("Execution role  policyfor fargate cluster allowed to managed cloudwatch log resources and pull container images from ECR"),
		PolicyDocument: aws.String(ECS_EXECUTION_ROLE_POLICY),
		PolicyName:     aws.String(EXECUSION_ROLE_POLICY_PREFIX + cluster.Name),
	}
	policyOut, err0 := c.IamClient.CreatePolicy(policyInput)
	policyArn := ""
	if err0 != nil {
		if aerr, ok := err0.(awserr.Error); ok {
			switch aerr.Code() {
			//If the policy already exists, we fetch the account to generate the Arn of the policy
			//if fetching the account fails, we return an error
			case iam.ErrCodeEntityAlreadyExistsException:
				log.Printf("[FARGATE] Execution role policy %s already exists. Not a blocking error", policyInput.PolicyName)
				_, account, err01 := c.GetRegionAccount()
				//if fetching the account fails, we return an error
				if err01 != nil {
					return Cluster{}, err01
				}
				policyArn = "arn:aws:iam::" + account + ":policy/" + *policyInput.PolicyName
			default:
				return Cluster{}, err0
			}
		}
	}
	policyArn = *policyOut.Policy.Arn

	roleInput := &iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(ECS_EXECUTION_ROLE_TRUST_POLICY),
		RoleName:                 aws.String("AmazonECSTaskExecutionRole" + cluster.Name),
		Description:              aws.String("Execution role for fargate cluster allowed to managed cloudwatch log resources and pull container images from ECR"),
	}
	roleOut, err1 := c.IamClient.CreateRole(roleInput)
	if err1 != nil {
		if aerr, ok := err1.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeEntityAlreadyExistsException:
				log.Printf("[FARGATE] Execution role  %s already exists. Not a blocking error", roleInput.RoleName)
			default:
				return Cluster{}, err1
			}
		}
	}
	attachInputs := &iam.AttachRolePolicyInput{
		PolicyArn: aws.String(policyArn),
		RoleName:  roleInput.RoleName,
	}
	_, err2 := c.IamClient.AttachRolePolicy(attachInputs)
	if err2 != nil {
		return Cluster{}, err2
	}*/

	executionRoleArn := cluster.ExecusionRoleArn

	/////////////////////////////
	//Creating Cluster
	////////////////////////////
	capProv := "FARGATE"
	if cluster.Spot {
		capProv = "FARGATE_SPOT"
	}
	input := &awsecs.CreateClusterInput{
		CapacityProviders: []*string{aws.String(capProv)},
		ClusterName:       aws.String(cluster.Name),
	}
	clusterOutput, err3 := c.Client.CreateCluster(input)
	if err3 != nil {
		return Cluster{}, err3
	}
	clusterArn := *clusterOutput.Cluster.ClusterArn
	/////////////////////////////
	//Creating Task Definition
	////////////////////////////
	//    * 512 (0.5 GB), 1024 (1 GB), 2048 (2 GB) - Available cpu values: 256 (.25
	//    vCPU)
	//    * 1024 (1 GB), 2048 (2 GB), 3072 (3 GB), 4096 (4 GB) - Available cpu values:
	//    512 (.5 vCPU)
	//    * 2048 (2 GB), 3072 (3 GB), 4096 (4 GB), 5120 (5 GB), 6144 (6 GB), 7168
	//    (7 GB), 8192 (8 GB) - Available cpu values: 1024 (1 vCPU)
	//    * Between 4096 (4 GB) and 16384 (16 GB) in increments of 1024 (1 GB) -
	//    Available cpu values: 2048 (2 vCPU)
	//    * Between 8192 (8 GB) and 30720 (30 GB) in increments of 1024 (1 GB) -
	//    Available cpu values: 4096 (4 vCPU)

	//Creating environment variable array form map
	var env []*ecs.KeyValuePair
	for key, val := range cluster.Env {
		env = append(env, &ecs.KeyValuePair{
			Name:  aws.String(key),
			Value: aws.String(val),
		})
	}
	//Create task definitions
	taskDefInput := &awsecs.RegisterTaskDefinitionInput{
		ContainerDefinitions: []*ecs.ContainerDefinition{
			&ecs.ContainerDefinition{
				Image:       aws.String(cluster.Image),
				Environment: env,
			},
		},
		Cpu:         aws.String(cluster.Cpu),
		Family:      aws.String(cluster.Name),
		Memory:      aws.String(cluster.Mem),
		NetworkMode: aws.String("awsvpc"),
		//The execusion role is created at cluster creation
		ExecutionRoleArn: aws.String(executionRoleArn),
		//The task role is passed to the init function has each cluster needs different task role
		TaskRoleArn: aws.String(cluster.RoleArn),
	}
	taskDefOutput, err4 := c.Client.RegisterTaskDefinition(taskDefInput)
	if err4 != nil {
		return Cluster{}, err4
	}
	taskDefinitionArn := *taskDefOutput.TaskDefinition.TaskDefinitionArn

	/////////////////////////////
	//Creating Service
	////////////////////////////
	subnets, err5 := c.GetSubnets(cluster.VpcID)
	if err5 != nil {
		return Cluster{}, err5
	}
	serviceInput := &awsecs.CreateServiceInput{
		ServiceName:  aws.String(cluster.Name),
		Cluster:      aws.String(clusterArn),
		DesiredCount: aws.Int64(cluster.Size),
		NetworkConfiguration: &awsecs.NetworkConfiguration{
			AwsvpcConfiguration: &awsecs.AwsVpcConfiguration{
				Subnets: subnets,
			},
		},
		SchedulingStrategy: aws.String("REPLICA"),
		TaskDefinition:     aws.String(taskDefinitionArn),
	}
	serviceOutput, err6 := c.Client.CreateService(serviceInput)
	if err6 != nil {
		return Cluster{}, err6
	}
	return Cluster{
		Arn:         *serviceOutput.Service.ClusterArn,
		Size:        *serviceOutput.Service.DesiredCount,
		ServiceName: *serviceOutput.Service.ServiceName,
		Name:        cluster.Name,
		Cpu:         cluster.Cpu,
		Mem:         cluster.Mem,
		Spot:        cluster.Spot,
		RoleArn:     cluster.RoleArn,
		VpcID:       cluster.VpcID,
		Image:       cluster.Image,
		Env:         cluster.Env,
	}, nil
}

func (c Config) DeleteCluster(clusterName string, clusterArn string) error {
	serviceInput := &awsecs.DeleteServiceInput{
		Service: aws.String(clusterName),
		Cluster: aws.String(clusterArn),
		Force:   aws.Bool(true),
	}
	_, err := c.Client.DeleteService(serviceInput)
	if err != nil {
		return err
	}
	//Check if necessary
	//taskDefInput := &awsecs.DeregisterTaskDefinitionInput{
	//	TaskDefinition: aws.String(cluster.RoleArn),
	//}
	//_, err = c.Client.DeregisterTaskDefinition(taskDefInput)
	//if err != nil {
	//	return err
	//}
	clusterInput := &awsecs.DeleteClusterInput{
		Cluster: aws.String(clusterArn),
	}
	_, err = c.Client.DeleteCluster(clusterInput)
	if err != nil {
		return err
	}
	return nil
}
