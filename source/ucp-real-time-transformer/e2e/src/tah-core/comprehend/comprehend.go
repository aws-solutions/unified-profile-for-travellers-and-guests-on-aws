package kendra

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsComprehend "github.com/aws/aws-sdk-go/service/comprehend"
)

type ComprehendCfg struct {
	Client     *awsComprehend.Comprehend
	EnpointArn string
	Region     string
}
type Result struct {
	Class      string
	Confidence float64
}

func Init(endpoint string, region string) ComprehendCfg {
	mySession := session.Must(session.NewSession())
	dbSession := awsComprehend.New(mySession, aws.NewConfig().WithRegion(region))
	// Create DynamoDB client
	return ComprehendCfg{
		Client:     dbSession,
		EnpointArn: endpoint,
		Region:     region,
	}

}

func (cph ComprehendCfg) GetBestClass(text string) (Result, error) {
	input := &awsComprehend.ClassifyDocumentInput{
		Text:        aws.String(text),
		EndpointArn: aws.String(cph.EnpointArn),
	}
	log.Printf("[comprehend][classify] input %+v", input)
	output, err := cph.Client.ClassifyDocument(input)
	if err != nil {
		log.Printf("[comprehend][Search] Errror during query: %+v", err)
		return Result{}, err
	}
	return Result{
		Class:      *output.Classes[0].Name,
		Confidence: *output.Classes[0].Score,
	}, nil

}
