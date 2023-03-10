package iot

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iot"
	awsiot "github.com/aws/aws-sdk-go/service/iot"
	"github.com/aws/aws-sdk-go/service/sts"
)

var IOT_ENPOINT_TYPE_DATA_ATS = "iot:Data-ATS"

type Config struct {
	Client    *awsiot.IoT //sdk client to make call to the AWS API
	StsClient *sts.STS    //sdk client to make call to the AWS API
	Topic     string      //main topic the device will publish on. this will ensure the device policy has the apropriate permissions
}

type Device struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	CertificateArn string   `json:"certificateArn"`
	CertificateId  string   `json:"certificateId"`
	CertificatePem string   `json:"certificatePem"`
	PrivateKey     string   `json:"privateKey"`
	PublicKey      string   `json:"publicKey"`
	CaCerts        []string `json:"caCerts"` //list of CA Certificates fo
}

func Init(topic string) Config {
	// Initialize a session that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials
	// and region from the shared configuration file ~/.aws/config.
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	// Create client
	return Config{
		Client:    awsiot.New(sess),
		Topic:     topic,
		StsClient: sts.New(sess),
	}
}

func (cfg Config) GetRegionAccount() (string, string, error) {
	identity, err := cfg.StsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	region := cfg.Client.Config.Region
	return *region, *identity.Account, err
}

func (cfg Config) GetEndpoint() (string, error) {
	input := &awsiot.DescribeEndpointInput{
		EndpointType: aws.String(IOT_ENPOINT_TYPE_DATA_ATS),
	}
	out, err := cfg.Client.DescribeEndpoint(input)
	if err != nil {
		return "", err
	}
	return *out.EndpointAddress, nil
}

////////////////////////////////////////
// Delete a device
////////////////////////////////////
// Notes on error management: we want to be able to retry  this function in case one deletion fails
// this is why we do not return on the first error and continue the flow.
// As an effect from this logic, we return the last error from the IOT core SDK. this way te caller
// can retry and eventually delete all resources. this assumes each step depends on all previous
// steps which implies is the last SDK call succeed, the whole deletion has succeded
func (cfg Config) DeleteDevice(device Device) error {
	policyName := device.Name + "Policy"
	//1-Detach policy from certificate
	policyAttachementInput := awsiot.DetachPolicyInput{
		PolicyName: aws.String(policyName),
		Target:     aws.String(device.CertificateArn),
	}
	_, err1 := cfg.Client.DetachPolicy(&policyAttachementInput)
	if err1 != nil {
		//non blocking error since if it can't detach. delete policy will fail
		log.Printf("[IOT][DELETE] Error during DetachPolicy : %+v", err1)
	}

	//2-Detach thing from certificate
	attachmentInput := awsiot.DetachThingPrincipalInput{
		Principal: aws.String(device.CertificateArn),
		ThingName: aws.String(device.Name),
	}
	_, err2 := cfg.Client.DetachThingPrincipal(&attachmentInput)
	if err2 != nil {
		//non blocking error since if it can't detach. delete policy will fail
		log.Printf("[IOT][DELETE] Error during DetachThingPrincipal: %+v", err2)
	}
	//3-Delete Policy
	policyInput := awsiot.DeletePolicyInput{
		PolicyName: aws.String(policyName),
	}
	_, err3 := cfg.Client.DeletePolicy(&policyInput)
	if isErrorForDelete(err3) {
		log.Printf("[IOT][DELETE] Error during DeletePolicy:%+v", err3)
		return err3
	}
	//4-Delete Certificate
	//To delete cretificate we must first deactivate it
	certifDeactivateInput := awsiot.UpdateCertificateInput{
		CertificateId: aws.String(device.CertificateId),
		NewStatus:     aws.String("INACTIVE"),
	}
	_, err41 := cfg.Client.UpdateCertificate(&certifDeactivateInput)
	if err41 != nil {
		//non blocking error since if it can't deactivate. delete certificate will fail
		log.Printf("[IOT][DELETE] Error during Deactivating certificate: %+v", err41)
	}
	credsInput := awsiot.DeleteCertificateInput{
		CertificateId: aws.String(device.CertificateId),
		ForceDelete:   aws.Bool(true),
	}
	_, err42 := cfg.Client.DeleteCertificate(&credsInput)
	if isErrorForDelete(err42) {
		log.Printf("[IOT][DELETE] Error during DeleteCertificate: %+v", err42)
		return err42
	}
	//5-Delete Thing
	thingINput := awsiot.DeleteThingInput{
		ThingName: aws.String(device.Name),
	}
	_, err5 := cfg.Client.DeleteThing(&thingINput)
	if isErrorForDelete(err5) {
		log.Printf("[IOT][DELETE] Error during DeleteThingInput: %+v", err5)
		return err5
	}
	return nil
}

/////////////////////////////////////////////////
// Create Device
////////////////////////////////////////////////
// This functtion creates all resources necessary to configure a device
// to be able to publish to an IOT core topic

func (cfg Config) CreateDevice(name string) (Device, error) {
	//1-Create an IOT Thing with the provided Device name
	thingINput := awsiot.CreateThingInput{
		ThingName: aws.String(name),
	}
	createThingOutput, err0 := cfg.Client.CreateThing(&thingINput)
	if err0 != nil {
		log.Printf("[IOT] Error while creating device: %+v", err0)
		return Device{}, err0
	}
	log.Printf("[IOT] Successfuly Created Thing %+v", createThingOutput)
	//2-Create associated credentals (certificate + key pair)
	credsInput := awsiot.CreateKeysAndCertificateInput{
		SetAsActive: aws.Bool(true),
	}
	createKeysAndCertificateOutput, err2 := cfg.Client.CreateKeysAndCertificate(&credsInput)
	if err2 != nil {
		log.Printf("[IOT] Error while creating device: %+v", err2)
		return Device{}, err2
	}
	log.Printf("[IOT] Successfuly Created Keys and Certificate: %+v", createKeysAndCertificateOutput)
	//4-Create Policy allowing the device to connect and publish to the configured topic
	//TODO: add aitional topics
	region, account, err3 := cfg.GetRegionAccount()
	if err3 != nil {
		log.Printf("[IOT] Error while creating device: %+v", err3)
		return Device{}, err3
	}
	policyContent := buildPolicy(cfg.Topic, region, account)
	policyInput := awsiot.CreatePolicyInput{
		PolicyDocument: aws.String(policyContent),
		PolicyName:     aws.String(name + "Policy"),
	}
	policyOutput, err4 := cfg.Client.CreatePolicy(&policyInput)
	if err4 != nil {
		log.Printf("[IOT] Error while creating device policy with content: %v. Error Message:%+v", policyContent, err4)
		return Device{}, err4
	}
	log.Printf("[IOT] Successfuly Created Policy: %+v", policyOutput)
	//5-Attach Thing to principal allowing the device to authenticate using the proviided certificate
	attachmentInput := awsiot.AttachThingPrincipalInput{
		Principal: createKeysAndCertificateOutput.CertificateArn,
		ThingName: aws.String(name),
	}
	_, err5 := cfg.Client.AttachThingPrincipal(&attachmentInput)
	if err5 != nil {
		log.Printf("[IOT] Error while creating device: %+v", err5)
		return Device{}, err5
	}
	log.Printf("[IOT] Successfuly attached policy to principal")
	//6-Attach Policy to certificate
	policyAttachementInput := awsiot.AttachPolicyInput{
		PolicyName: policyInput.PolicyName,
		Target:     createKeysAndCertificateOutput.CertificateArn,
	}
	_, err6 := cfg.Client.AttachPolicy(&policyAttachementInput)
	if err6 != nil {
		log.Printf("[IOT] Error while creating device: %+v", err6)
		return Device{}, err6
	}
	//7-returns Device Struct with required data to onboard the device
	return Device{
		ID:             *createThingOutput.ThingId,
		Name:           *createThingOutput.ThingName,
		CertificateArn: *createKeysAndCertificateOutput.CertificateArn,
		CertificateId:  *createKeysAndCertificateOutput.CertificateId,
		CertificatePem: *createKeysAndCertificateOutput.CertificatePem,
		PrivateKey:     *createKeysAndCertificateOutput.KeyPair.PrivateKey,
		PublicKey:      *createKeysAndCertificateOutput.KeyPair.PublicKey,
		CaCerts: []string{
			"https://www.amazontrust.com/repository/AmazonRootCA1.pem",
			"https://www.amazontrust.com/repository/AmazonRootCA2.pem",
			"https://www.amazontrust.com/repository/AmazonRootCA3.pem",
			"https://www.amazontrust.com/repository/AmazonRootCA4.pem",
			"https://www.amazontrust.com/repository/G2-RootCA1.pem",
			"https://www.amazontrust.com/repository/G2-RootCA2.pem",
			"https://www.amazontrust.com/repository/G2-RootCA3.pem",
			"https://www.amazontrust.com/repository/G2-RootCA4.pem",
			"https://www.amazontrust.com/repository/SFSRootCAG2.pem"},
	}, nil
}

//This function creates a policy for the device to be allowed connectino and publish on teh configured topic.
//TODO: we should probably used structs + json marshaller, an ARN builder and allow multiple topics in the long run
func buildPolicy(topic string, region string, account string) string {
	return `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "iot:Connect"
            ],
            "Resource": [
                "arn:aws:iot:` + region + `:` + account + `:client/${iot:Connection.Thing.ThingName}"
            ]
        },
        {
            "Effect": "Allow",
            "Action": [
                "iot:Publish"
            ],
            "Resource": [
                "arn:aws:iot:` + region + `:` + account + `:topic/` + topic + `/*"
            ]
        }
    ]
}`
}

//This function returns true if the error return is non nil and is not a ErrCodeResourceNotFoundException.
//this is ment to be use to assess the validity of resource deletion operation
func isErrorForDelete(err error) bool {
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() != iot.ErrCodeResourceNotFoundException {
				return true
			}
		}
	}
	return false
}
