package cloudfront

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsCloudfront "github.com/aws/aws-sdk-go/service/cloudfront"
)

type Config struct {
	Client         *awsCloudfront.CloudFront
	DistributionID string
}

func Init(distributionId string) Config {
	mySession := session.Must(session.NewSession())
	// Create a Athena client from just a session.
	svc := awsCloudfront.New(mySession)
	return Config{
		Client:         svc,
		DistributionID: distributionId,
	}
}

func (c Config) Invalidate() error {
	in := &awsCloudfront.CreateInvalidationInput{
		DistributionId: aws.String(c.DistributionID),
		InvalidationBatch: &awsCloudfront.InvalidationBatch{
			Paths: &awsCloudfront.Paths{
				Items:    []*string{aws.String("/*")},
				Quantity: aws.Int64(1),
			},
			CallerReference: aws.String(time.Now().String()),
		},
	}
	_, err := c.Client.CreateInvalidation(in)
	return err
}
