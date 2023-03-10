package honeycode

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsQuickSight "github.com/aws/aws-sdk-go/service/quicksight"
)

type QuickSightConfig struct {
	Client    *awsQuickSight.QuickSight
	AccountID string
	UserArn   string
}
type QuickSightDashboardConfig struct {
	Url string
}

func Init(region string, accountID string, userArn string) QuickSightConfig {
	mySession := session.Must(session.NewSession())
	client := awsQuickSight.New(mySession, aws.NewConfig().WithRegion(region))
	log.Printf("[QUICKSIGHT] new client: %+v", client)

	// Create DynamoDB client
	return QuickSightConfig{
		Client:    client,
		AccountID: accountID,
		UserArn:   userArn,
	}

}

func (quick QuickSightConfig) GetDashboardUrl(dashboardId string) (QuickSightDashboardConfig, error) {
	input := &awsQuickSight.GetDashboardEmbedUrlInput{
		AwsAccountId: aws.String(quick.AccountID),
		DashboardId:  aws.String(dashboardId),
		IdentityType: aws.String("QUICKSIGHT"),
		UserArn:      aws.String(quick.UserArn),
	}
	log.Printf("[QUICKSIGHT][GetDashboardUrl] request %+v", input)
	out, err := quick.Client.GetDashboardEmbedUrl(input)
	if err != nil {
		log.Printf("[QUICKSIGHT][GetDashboardUrl] error %+v", err)
		return QuickSightDashboardConfig{}, err
	}
	log.Printf("[QUICKSIGHT][GetDashboardUrl] response %+v", out)
	return QuickSightDashboardConfig{
		Url: *out.EmbedUrl,
	}, err

}
