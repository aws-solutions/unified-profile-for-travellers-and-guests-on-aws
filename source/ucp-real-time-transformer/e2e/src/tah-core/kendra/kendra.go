package kendra

import (
	"log"
	"strings"

	s3 "tah/core/s3"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsKendra "github.com/aws/aws-sdk-go/service/kendra"
)

const KENDRA_ADDITIONAL_ATTRIBUTES_KEY_ANSWER_TEXT = "AnswerText"
const KENDRA_ANSWER_TYPE_ANSWER = "answer"
const KENDRA_ANSWER_TYPE_DOCUMENT = "document"
const KENDRA_ANSWER_TYPE_QUESTION_ANSWER = "questionAndAnswer"

type KendraCfg struct {
	Client  *awsKendra.Kendra
	IndexId string
	Region  string
}
type Document struct {
	Excerpt    string
	Id         string
	Title      string
	Url        string
	IsAnswer   bool
	AnswerText string
	S3Bucket   string
	S3Path     string
	FullText   string
}

func Init(indexId string, region string) KendraCfg {
	mySession := session.Must(session.NewSession())
	dbSession := awsKendra.New(mySession, aws.NewConfig().WithRegion(region))
	// Create DynamoDB client
	return KendraCfg{
		Client:  dbSession,
		IndexId: indexId,
		Region:  region,
	}

}

func (ken KendraCfg) Search(query string, answerType string, all bool, getBody bool) ([]Document, error) {
	input := &awsKendra.QueryInput{
		IndexId:   aws.String(ken.IndexId),
		QueryText: aws.String(query),
		PageSize:  aws.Int64(100),
	}
	log.Printf("[kendra][Search] QueryInput %+v", input)
	output, err := ken.Client.Query(input)
	if err != nil {
		log.Printf("[kendra][Search] Errror during query: %+v", err)
		return []Document{}, err
	}
	log.Printf("[kendra][Search] QueryRes %+v", output)

	page := 0
	results := toResults(output.ResultItems)
	log.Printf("[kendra][Search] Parsed results %+v", results)
	for page < int(float64(*output.TotalNumberOfResults)/100.0) && all {
		page++
		log.Printf("[kendra][Search] Pagination ON getting page %+v", page)
		input = &awsKendra.QueryInput{
			IndexId:    aws.String(ken.IndexId),
			QueryText:  aws.String(query),
			PageNumber: aws.Int64(int64(page)),
			PageSize:   aws.Int64(100),
		}
		output, err := ken.Client.Query(input)
		if err != nil {
			log.Printf("[kendra][Search] Errror during query: %+v", err)
			return []Document{}, err
		}
		for _, queryResItem := range output.ResultItems {
			results = append(results, toResult(queryResItem))
		}
	}
	log.Printf("[kendra][Search] Got all results. NRes =  %+v", len(results))
	if answerType == KENDRA_ANSWER_TYPE_ANSWER {
		log.Printf("[kendra][Search]requested answer type =  %s", KENDRA_ANSWER_TYPE_ANSWER)
		filteredRes := []Document{}
		for _, result := range results {
			if result.IsAnswer {
				filteredRes = append(filteredRes, result)
			}
		}
		log.Printf("[kendra][Search] Filtered answers only. NRes =  %+v", len(filteredRes))
		results = filteredRes
	}
	if getBody && len(results) > 0 {
		log.Printf("[kendra][Search] Getting  body content of kendra answer (S3 source only)")

		s3c := s3.Init(results[0].S3Bucket, "", ken.Region)
		paths := []string{}
		for _, item := range results {
			paths = append(paths, item.S3Path)
		}
		objects, err := s3c.GetManyAsText(paths)
		if err != nil {
			log.Printf("[kendra][Search] Error getting Documnet body %+v", err)
			return []Document{}, err
		}
		log.Printf("[kendra][Search] Got %+v objects from Amazon S3", len(objects))
		for i, _ := range results {
			results[i].FullText = objects[i]
		}
	}
	return results, nil
}

//Only Returns answers
func (ken KendraCfg) SearchAnswers(query string) ([]Document, error) {
	return ken.Search(query, KENDRA_ANSWER_TYPE_ANSWER, false, true)
}

//Only Returns answers without body
func (ken KendraCfg) SearchAnswersWithoutBody(query string) ([]Document, error) {
	return ken.Search(query, KENDRA_ANSWER_TYPE_ANSWER, false, false)
}

//Returns all without body
func (ken KendraCfg) SearchAllWithoutBody(query string) ([]Document, error) {
	return ken.Search(query, "", false, false)
}

func toResult(queryResItem *awsKendra.QueryResultItem) Document {
	//check if this is returned as a kendra answer
	kendraAnswer := parseAnswerText(queryResItem.AdditionalAttributes)
	bucket, path := parseS3info(*queryResItem.DocumentId)
	titleText := ""
	if queryResItem.DocumentTitle != nil && (*queryResItem.DocumentTitle).Text != nil {
		titleText = *(*queryResItem.DocumentTitle).Text
	}

	return Document{
		Excerpt: *queryResItem.DocumentExcerpt.Text,
		//for S3, should be s s3://<bucketName>/<path>
		Id:         *queryResItem.DocumentId,
		Title:      titleText,
		Url:        *queryResItem.DocumentURI,
		IsAnswer:   kendraAnswer != "",
		AnswerText: kendraAnswer,
		S3Bucket:   bucket,
		S3Path:     path,
	}
}

func parseAnswerText(additionalResults []*awsKendra.AdditionalResultAttribute) string {
	for _, res := range additionalResults {
		if *res.Key == KENDRA_ADDITIONAL_ATTRIBUTES_KEY_ANSWER_TEXT {
			return *res.Value.TextWithHighlightsValue.Text
		}
	}
	return ""
}

//parses s3 bucket name from URL (is correct), returns empty string if not an S3 url
func parseS3info(docId string) (string, string) {
	split := strings.Split(docId, "/")
	if len(split) > 3 && split[0] == "s3:" {
		return split[2], docId[len("s3://"+split[2]+"/"):]
	}
	return "", ""
}

func toResults(kendraRes []*awsKendra.QueryResultItem) []Document {
	res := []Document{}
	for _, queryResItem := range kendraRes {
		res = append(res, toResult(queryResItem))
	}
	return res
}
