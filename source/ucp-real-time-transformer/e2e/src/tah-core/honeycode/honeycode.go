package honeycode

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsHoneyCode "github.com/aws/aws-sdk-go/service/honeycode"
)

type HoneyCodeConfig struct {
	Client     *awsHoneyCode.Honeycode
	WorkbookId string
	Region     string
}

func Init(workbookId string, region string) HoneyCodeConfig {
	mySession := session.Must(session.NewSession())
	client := awsHoneyCode.New(mySession, aws.NewConfig().WithRegion(region))
	// Create DynamoDB client
	return HoneyCodeConfig{
		Client:     client,
		WorkbookId: workbookId,
		Region:     region,
	}

}

func (hon HoneyCodeConfig) GetListData(appId string, screenId string, listName string) ([][]string, error) {
	input := &awsHoneyCode.GetScreenDataInput{
		AppId:      aws.String(appId),
		ScreenId:   aws.String(screenId),
		WorkbookId: aws.String(hon.WorkbookId),
	}
	log.Printf("[HONEYCODE] Honeycode request %+v", input)
	out, err := hon.Client.GetScreenData(input)
	log.Printf("[HONEYCODE] Honeycode response %+v", out)
	if err != nil {
		log.Printf("[HONEYCODE] Honeycode response ERROR %+v", err)
		return [][]string{}, fmt.Errorf("List not found on app screen: error: %v", err)
	}
	for key, resultSet := range out.Results {
		if key == listName {
			log.Printf("[HONEYCODE] found list: %s", listName)
			response := [][]string{}
			for _, row := range resultSet.Rows {
				item := []string{}
				for _, col := range row.DataItems {
					item = append(item, *col.RawValue)
				}
				response = append(response, item)
			}
			return response, nil
		}
	}
	return [][]string{}, fmt.Errorf("List %s not found on app screen", listName)
}

func (hon HoneyCodeConfig) GetWorkbookData(tableID string) ([][]string, error) {
	input := &awsHoneyCode.ListTableRowsInput{
		TableId:    aws.String(tableID),
		WorkbookId: aws.String(hon.WorkbookId),
		MaxResults: aws.Int64(100),
	}
	log.Printf("[HONEYCODE] Honeycode request %+v", input)
	out, err := hon.Client.ListTableRows(input)
	log.Printf("[HONEYCODE] Honeycode response %+v", out)
	if err != nil {
		log.Printf("[HONEYCODE] Honeycode response ERROR %+v", err)
		return [][]string{}, fmt.Errorf("Data not double in workbook: error: %v", err)
	}
	response := [][]string{}
	for _, row := range out.Rows {
		if row != nil {
			item := []string{}
			for _, cell := range row.Cells {
				log.Printf("[HONEYCODE] Honeycode response cell: %+v", cell)
				if cell != nil && cell.RawValue != nil {
					item = append(item, *cell.RawValue)
				}
			}
			response = append(response, item)
		}
	}
	return response, nil
}
