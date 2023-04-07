package s3

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"net"

	"github.com/aws/aws-sdk-go/aws"
	"golang.org/x/net/http2"

	//"github.com/aws/aws-sdk-go/aws/awsutil"
	//"github.com/aws/aws-sdk-go/aws/credentials"
	"encoding/json"
	"errors"
	_ "image/jpeg"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"sync"
	core "tah/core/core"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const S3_HOST = "https://s3.amazonaws.com"
const LIST_PART_MAX_PAGE = 1000
const DEFAULT_BATCH_SIZE = 100

type IS3Config interface {
	SetTx(tx core.Transaction) error
	SaveJson(path string, id string, data []byte) error
	GetBucket() string
	GetRegion() string
}

type S3Config struct {
	Svc       *s3.S3
	Region    string
	Bucket    string
	AccessKey string
	Secret    string
	Token     string
	Path      string
	Tx        core.Transaction
}

type S3MultipartConfig struct {
	S3Config          S3Config
	MultipartUploadId string
	S3Key             string
	ChunkSize         int64
	NThreads          int64
}

type BucketPolicyDocument struct {
	Version   string
	Id        string
	Statement []BucketPolicyStatementEntry
}
type BucketPolicyStatementEntry struct {
	Sid       string
	Effect    string
	Action    []string
	Principal map[string][]string
	Resource  []string
}

func InitRegion(region string) S3Config {
	cfg := aws.NewConfig().WithRegion(region)
	svc := s3.New(session.New(), cfg)
	return S3Config{
		Svc:    svc,
		Region: region,
	}
}

func InitWithRandBucket(prefix string, path string, region string) (S3Config, error) {
	s3c := InitRegion(region)
	name, err := s3c.CreateRandomBucket(prefix)
	return Init(name, path, region), err
}

func Init(bucket string, path string, region string) S3Config {
	//we overide the default HTTP client to allo massive parallel upload to S3
	httpClient, err := NewHTTPClientWithSettings(HTTPClientSettings{
		Connect:          5 * time.Second,
		ExpectContinue:   1 * time.Second,
		IdleConn:         90 * time.Second,
		ConnKeepAlive:    30 * time.Second,
		MaxAllIdleConns:  1000,
		MaxHostIdleConns: 1000,
		ResponseHeader:   5 * time.Second,
		TLSHandshake:     5 * time.Second,
	})
	if err != nil {
		log.Printf("Got an error creating custom HTTP client: %v", err)
	}
	cfg := aws.NewConfig().WithRegion(region)
	svc := s3.New(session.New(&aws.Config{
		HTTPClient: httpClient,
	}), cfg)
	return S3Config{
		Svc:    svc,
		Bucket: bucket,
		Region: region,
		Path:   path}
}

type HTTPClientSettings struct {
	Connect          time.Duration
	ConnKeepAlive    time.Duration
	ExpectContinue   time.Duration
	IdleConn         time.Duration
	MaxAllIdleConns  int
	MaxHostIdleConns int
	ResponseHeader   time.Duration
	TLSHandshake     time.Duration
}

func NewHTTPClientWithSettings(httpSettings HTTPClientSettings) (*http.Client, error) {
	var client http.Client
	tr := &http.Transport{
		ResponseHeaderTimeout: httpSettings.ResponseHeader,
		Proxy:                 http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			KeepAlive: httpSettings.ConnKeepAlive,
			DualStack: true,
			Timeout:   httpSettings.Connect,
		}).DialContext,
		MaxIdleConns:          httpSettings.MaxAllIdleConns,
		IdleConnTimeout:       httpSettings.IdleConn,
		TLSHandshakeTimeout:   httpSettings.TLSHandshake,
		MaxIdleConnsPerHost:   httpSettings.MaxHostIdleConns,
		ExpectContinueTimeout: httpSettings.ExpectContinue,
	}
	// So client makes HTTP/2 requests
	err := http2.ConfigureTransport(tr)
	if err != nil {
		return &client, err
	}

	return &http.Client{
		Transport: tr,
	}, nil
}

func (s3c *S3Config) SetTx(tx core.Transaction) error {
	tx.LogPrefix = "S3"
	s3c.Tx = tx
	return nil
}

func (s3c S3Config) Url() string {
	return S3_HOST + "/" + s3c.Bucket + "/" + s3c.Path
}
func (s3c S3Config) GetBucket() string {
	return s3c.Bucket
}
func (s3c S3Config) GetRegion() string {
	return s3c.Region
}

func (s3c S3Config) Key(key string) string {
	if key == "" {
		panic("Must provide a key to this function")
	}
	if s3c.Path == "" {
		return key
	}
	return s3c.Path + "/" + key
}

func (mpc S3MultipartConfig) Key() string {
	return mpc.S3Config.Key(mpc.S3Key)
}

func (s3c *S3Config) CreateBucket(name string) error {
	s3c.Tx.Log("Creating Bucket %s", name)
	input := &s3.CreateBucketInput{
		Bucket: aws.String(name)}
	_, err := s3c.Svc.CreateBucket(input)
	return err
}

func (s3c S3Config) CreateRandomBucket(prefix string) (string, error) {
	name := prefix + "-" + strings.ToLower(core.GeneratUniqueId())
	err := s3c.CreateBucket(name)
	return name, err
}

func (s3c S3Config) DeleteBucket(name string) error {
	s3c.Tx.Log("Deleting Bucket %s", s3c.Bucket)
	input := &s3.DeleteBucketInput{
		Bucket: aws.String(name)}
	_, err := s3c.Svc.DeleteBucket(input)
	return err
}

func (s3c S3Config) EmptyBucket() error {
	s3c.Tx.Log("Emptying Bucket %s", s3c.Bucket)
	// Setup BatchDeleteIterator to iterate through a list of objects.
	iter := s3manager.NewDeleteListIterator(s3c.Svc, &s3.ListObjectsInput{
		Bucket: aws.String(s3c.Bucket),
	})
	// Traverse iterator deleting each object
	if err := s3manager.NewBatchDeleteWithClient(s3c.Svc).Delete(aws.BackgroundContext(), iter); err != nil {
		s3c.Tx.Log("Unable to delete objects from bucket %q, %v", s3c.Bucket, err)
		return err
	}
	return nil
}

//TODO: seem to be a bug when bucket is empty. to check
func (s3c S3Config) EmptyAndDelete() error {
	s3c.Tx.Log("Emptying and deleting bucket %s", s3c.Bucket)
	err := s3c.EmptyBucket()
	if err != nil {
		s3c.Tx.Log("Could not empty")
		return err
	}
	err = s3c.DeleteBucket(s3c.Bucket)
	if err != nil {
		s3c.Tx.Log("Could not delete")
		return err
	}
	return nil
}

func (s3c S3Config) AddPolicy(name string, resources []string, actions []string, principals map[string][]string) error {
	s3c.Tx.Log("Adding Bucket policy %s", s3c.Bucket)
	input := &s3.PutBucketPolicyInput{
		Bucket: aws.String(name),
		Policy: aws.String(s3c.buildBucketPolicy(resources, actions, principals)),
	}
	_, err := s3c.Svc.PutBucketPolicy(input)
	return err
}

func (s3c S3Config) buildBucketPolicy(resources []string, actions []string, principals map[string][]string) string {
	doc := BucketPolicyDocument{
		Version: "2012-10-17",
		Id:      "BucketPolicy",
		Statement: []BucketPolicyStatementEntry{
			BucketPolicyStatementEntry{
				Effect:    "Allow",
				Action:    actions,
				Resource:  resources,
				Principal: principals,
			},
		},
	}
	policyJson, _ := json.Marshal(doc)
	s3c.Tx.Log("Policy Created:  %s", string(policyJson))
	return string(policyJson)
}

func (s3c S3Config) UploadJpegToS3(path string, id string, rawData string) (string, error) {
	//decoding
	log.Printf("[S3][UploadJpegToS3] uploading picture with content %+v", rawData)
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(rawData))
	byteArray, err0 := ioutil.ReadAll(reader)
	if err0 != nil {
		log.Printf("[S3][UploadJpegToS3] error during base64 decoding: %+v", err0)
		return "", err0
	}
	log.Printf("[S3][UploadJpegToS3] Byte Array: %+v", byteArray)
	fileType := http.DetectContentType(byteArray)
	log.Printf("[S3][UploadJpegToS3] Content Type: %+v", fileType)
	fileBytes := bytes.NewReader(byteArray)
	log.Printf("[S3][UploadJpegToS3] Reader: %+v", fileBytes)
	params := &s3.PutObjectInput{
		Bucket:        aws.String(s3c.Bucket),
		Key:           aws.String(s3c.Path + "/" + path + "/" + id),
		Body:          fileBytes,
		ContentLength: aws.Int64(int64(len(byteArray))),
		ContentType:   aws.String(fileType)}
	log.Printf("[S3][UploadJpegToS3] upload request: %+v", s3c.Bucket+"/"+s3c.Path+"/"+path+"/"+id)
	_, err := s3c.Svc.PutObject(params)
	if err != nil {
		log.Printf("[S3][UploadJpegToS3] error durinig upload: %+v", err)
	}
	return s3c.Url() + "/" + path + "/" + id, err
}

func (s3c S3Config) SaveJson(path string, id string, data []byte) error {
	return s3c.Save(path, id, data, "application/json")
}
func (s3c S3Config) SaveXml(path string, id string, data []byte) error {
	return s3c.Save(path, id, data, "application/xml")
}

func (s3c S3Config) SaveManyJson(paths []string, data []string) []error {
	return s3c.SaveMany(paths, data, "application/json", DEFAULT_BATCH_SIZE)
}

func (s3c S3Config) SaveMany(paths []string, data []string, contentType string, batchSize int) []error {
	if len(paths) != len(data) {
		s3c.Tx.Log("[SaveMany] Length of S3 paths and data does not match (%v versus %v)\n", len(paths), len(data))
		return []error{errors.New("Length of S3 paths and data does not match")}
	}
	var wg sync.WaitGroup
	var errs []error

	batchedData := core.Chunk(core.InterfaceSlice(data), batchSize)
	batchedPaths := core.Chunk(core.InterfaceSlice(paths), batchSize)
	s3c.Tx.Log("[SaveMany] splitting the content in \n", len(data))
	mu := &sync.Mutex{}
	for i, batch := range batchedData {
		log.Printf("[SaveMany] Adding %+v thead to waitgroup.\n", len(batch))
		wg.Add(len(batch))
		for j, _ := range batch {
			go func(index int) {
				b := []byte(batch[index].(string))
				err := s3c.Save("", batchedPaths[i][index].(string), b, contentType)
				if err != nil {
					//this ok to lock here given that it is only in error cases
					mu.Lock()
					errs = append(errs, err)
					mu.Unlock()
				}
				wg.Done()
			}(j)
		}
		wg.Wait()
		if len(errs) > 0 {
			return errs
		}
	}
	return errs
}

func (s3c S3Config) DeleteMany(paths []string, batchSize int) []error {
	var wg sync.WaitGroup
	var errs []error
	batchedPaths := core.Chunk(core.InterfaceSlice(paths), batchSize)
	s3c.Tx.Log("[DeleteMany] splitting the content in \n", len(paths))
	mu := &sync.Mutex{}
	for i, batch := range batchedPaths {
		log.Printf("[DeleteMany] Adding %+v thead to waitgroup.\n", len(batch))
		wg.Add(len(batch))
		for j, _ := range batch {
			go func(index int) {
				err := s3c.Delete("", batchedPaths[i][index].(string))
				if err != nil {
					//this ok to lock here given that it is only in error cases
					mu.Lock()
					errs = append(errs, err)
					mu.Unlock()
				}
				wg.Done()
			}(j)
		}
		wg.Wait()
		if len(errs) > 0 {
			return errs
		}
	}
	return errs
}

func (s3c S3Config) Save(path string, id string, data []byte, contentType string) error {
	dataBytes := bytes.NewReader(data)
	key := s3c.Path + "/" + path + "/" + id
	if path == "" {
		key = s3c.Path + "/" + id
	}
	params := &s3.PutObjectInput{
		Bucket:      aws.String(s3c.Bucket),
		Key:         aws.String(key),
		Body:        dataBytes,
		ContentType: aws.String(contentType)}
	s3c.Tx.Log("Uploading data to S3 %+v", s3c.Bucket+"/"+key)
	result, err := s3c.Svc.PutObject(params)
	if err == nil {
		s3c.Tx.Log("Result of upload to S3: %+v", result)
	} else {
		s3c.Tx.Log("Error while uploading to S3: %+v", err)
	}
	return err
}

func (s3c S3Config) CreatePresignedUrl(key string, dur time.Duration) (string, error) {
	if s3c.Path != "" {
		key = s3c.Path + "/" + key
	}
	req, _ := s3c.Svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s3c.Bucket),
		Key:    aws.String(key),
	})
	urlStr, err := req.Presign(dur)
	s3c.Tx.Log("[S3] Error for CreatePresignedUrl %+v", err)
	return urlStr, err
}

//Function to get JSON encode object and unmarshall it
func (s3c S3Config) Get(key string, object interface{}) error {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(s3c.Region)},
	)
	downloader := s3manager.NewDownloader(sess)
	buf := aws.NewWriteAtBuffer([]byte{})
	s3c.Tx.Log("DOWNLOADING FROM S3 bucket %s at key %s", s3c.Bucket, s3c.Path+"/"+key)
	numBytes, err := downloader.Download(buf,
		&s3.GetObjectInput{
			Bucket: aws.String(s3c.Bucket),
			Key:    aws.String(s3c.Path + "/" + key),
		})
	s3c.Tx.Log("Downloaded %v Bytes", numBytes)
	s3c.Tx.Log("[S3] Object Downloaded %s", buf.Bytes())
	json.Unmarshal(buf.Bytes(), object)
	if err != nil {
		s3c.Tx.Log("[S3] ERROR while downloading object %+v", err)
	}
	return err
}

func (s3c S3Config) Delete(path string, id string) error {
	key := s3c.Path + "/" + path + "/" + id
	if path == "" {
		key = s3c.Path + "/" + id
	}
	s3c.Tx.Log("[S3] Deleting object %+v", key)
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s3c.Bucket),
		Key:    aws.String(key),
	}
	_, err := s3c.Svc.DeleteObject(input)
	if err != nil {
		s3c.Tx.Log("[S3] ERROR while deletin object %+v", err)
	}
	return err
}

func (s3c S3Config) GetTextObj(key string) (string, error) {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(s3c.Region)},
	)
	downloader := s3manager.NewDownloader(sess)
	buf := aws.NewWriteAtBuffer([]byte{})
	log.Printf("DOWNLOADING FROM S3 bucket %s at key %s", s3c.Bucket, s3c.Path+"/"+key)
	numBytes, err := downloader.Download(buf,
		&s3.GetObjectInput{
			Bucket: aws.String(s3c.Bucket),
			Key:    aws.String(s3c.Path + "/" + key),
		})
	log.Printf("Downloaded %v Bytes", numBytes)
	//log.Printf("[S3] Object Downloaded %s", buf.Bytes())
	if err != nil {
		log.Printf("[S3] ERROR while downloading object %+v", err)
	}
	return string(buf.Bytes()), err
}

//Parse a CSV file in S3 (to be tested at large scale)
func (s3c S3Config) ParseCsvFromS3(fileName string) ([][]string, error) {
	csvContent, err := s3c.GetTextObj(fileName)
	if err != nil {
		s3c.Tx.Log("Could not fetch CSV file form S3: %v", err)
		return [][]string{}, err
	}
	strReader := strings.NewReader(csvContent)
	csvReader := csv.NewReader(strReader)
	csvData, err2 := csvReader.ReadAll()
	return csvData, err2
}

func (s3c S3Config) GetManyAsText(keys []string) ([]string, error) {
	log.Printf("[S3][GetManyAsText] GetManyAsText from S3 bucket %s", s3c.Bucket)
	var wg sync.WaitGroup
	responses := make([]string, len(keys), len(keys))
	wg.Add(len(keys))
	var lastErr error
	for ind, _ := range keys {
		go func(ind int, responses *[]string, lastErr *error) {
			log.Printf("[S3][GetManyAsText]] getting object: %s\n", keys[ind])
			defer wg.Done()
			res, err := s3c.GetTextObj(keys[ind])
			if err == nil {
				(*responses)[ind] = res
			} else {
				*lastErr = err
			}
		}(ind, &responses, &lastErr)
	}
	wg.Wait()
	return responses, lastErr
}

//List all obbjects in a S3 bucket. thsi function could run for a long time if the bucket has many objects
func (s3c S3Config) Search(prefix string, maxRes int) ([]string, error) {
	s3c.Tx.Log("[S3] listing object with prefix %+v", prefix)
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(s3c.Region)},
	)
	svc := s3.New(sess)
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s3c.Bucket),
		Prefix: aws.String(prefix),
	}
	result, err := svc.ListObjectsV2(input)
	if err != nil {
		s3c.Tx.Log("[S3] ERROR while listing object %+v", err)
	}
	res := []string{}
	for _, obj := range result.Contents {
		res = append(res, *obj.Key)
	}
	if len(res) >= maxRes {
		return res[0:maxRes], err
	}
	for result.IsTruncated != nil && *result.IsTruncated {
		s3c.Tx.Log("[S3] List is truncated getting next page")
		input.ContinuationToken = result.NextContinuationToken
		result, err = svc.ListObjectsV2(input)
		if err != nil {
			s3c.Tx.Log("[S3] ERROR while listing object %+v", err)
		}
		for _, obj := range result.Contents {
			res = append(res, *obj.Key)
		}
	}
	if len(res) >= maxRes {
		return res[0:maxRes], err
	}
	return res, err
}

func (s3c S3Config) UploadFile(key string, filePath string) error {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(s3c.Region)},
	)
	// Create an uploader with the session and custom options
	uploader := s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
		u.PartSize = 5 * 1024 * 1024 // The minimum/default allowed part size is 5MB
		u.Concurrency = 2            // default is 5
	})

	//open the file
	f, err := os.Open(filePath)
	if err != nil {
		log.Printf("failed to open file %q, %v", filePath, err)
		return err
	}

	// Upload the file to S3.
	_, err2 := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s3c.Bucket),
		Key:    aws.String(s3c.Path + "/" + key),
		Body:   f,
	})

	if err2 != nil {
		log.Printf("File upload faile with error %+v", err)
		return err2
	}

	return nil
}

//TODO: compute the chunksize from the filezise to avoid the 10000 parts limit
//TODO: get default Nthread from system
func (s3c S3Config) StartMPUpload(key string) (S3MultipartConfig, error) {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(s3c.Region)},
	)
	svc := s3.New(sess)
	mpConfig := S3MultipartConfig{S3Config: s3c, S3Key: key, ChunkSize: 5, NThreads: 2}

	params := &s3.CreateMultipartUploadInput{
		Bucket: aws.String(s3c.Bucket),   // Required
		Key:    aws.String(s3c.Key(key)), // Required
	}
	resp, err := svc.CreateMultipartUpload(params)

	if err != nil {
		log.Printf("Failed to create new Multipart Upload: %+v", err)
		return mpConfig, err
	}
	mpConfig.MultipartUploadId = *resp.UploadId
	return mpConfig, nil
}

func (s3c S3Config) ListMpUploads() ([]S3MultipartConfig, error) {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(s3c.Region)},
	)
	svc := s3.New(sess)

	uploads := []S3MultipartConfig{}

	params := &s3.ListMultipartUploadsInput{
		Bucket: aws.String(s3c.Bucket), // Required
	}
	res, err := svc.ListMultipartUploads(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		log.Printf("Failed to list Multipart Uploads: %+v", err)
		return uploads, err
	}

	for _, upload := range res.Uploads {
		uploads = append(uploads, S3MultipartConfig{S3Config: s3c,
			MultipartUploadId: *upload.UploadId,
			S3Key:             *upload.Key,
		})
	}

	return uploads, nil
}

func (mpc S3MultipartConfig) Abort() error {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(mpc.S3Config.Region)},
	)
	svc := s3.New(sess)

	params := &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(mpc.S3Config.Bucket),   // Required
		Key:      aws.String(mpc.Key()),             // Required
		UploadId: aws.String(mpc.MultipartUploadId), // Required
	}
	_, err := svc.AbortMultipartUpload(params)

	if err != nil {
		log.Printf("Failed to abort Multipart Upload: %+v", err)
		return err
	}
	return nil
}

func (mpc S3MultipartConfig) Parts() ([]*s3.CompletedPart, error) {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(mpc.S3Config.Region)},
	)
	svc := s3.New(sess)

	log.Printf("[Parts] Listing parts for Multipart Upload %+v", mpc.MultipartUploadId)

	parts := []*s3.CompletedPart{}

	params := &s3.ListPartsInput{
		Bucket:   aws.String(mpc.S3Config.Bucket),   // Required
		Key:      aws.String(mpc.Key()),             // Required
		UploadId: aws.String(mpc.MultipartUploadId), // Required
	}
	resp, err := svc.ListParts(params)
	page := 0
	for *resp.IsTruncated && err == nil {
		for _, part := range resp.Parts {
			parts = append(parts, &s3.CompletedPart{ETag: part.ETag, PartNumber: part.PartNumber})
		}
		page++
		log.Printf("[Parts] List is paginated. Getting page %v", page)
		params.PartNumberMarker = resp.NextPartNumberMarker
		resp, err = svc.ListParts(params)

		if page > LIST_PART_MAX_PAGE {
			log.Printf("[Parts] Max number of pages (%v) exeeded. exiting to avoid infinite loop", LIST_PART_MAX_PAGE)
			return parts, errors.New("Max number of pages exeeded. exiting to avoid infinite loop")
		}
	}
	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		log.Printf("Failed to list Uploaded Parts: %+v", err)
		return parts, err
	}
	for _, part := range resp.Parts {
		parts = append(parts, &s3.CompletedPart{ETag: part.ETag, PartNumber: part.PartNumber})
	}

	log.Printf("[Parts] Found %v parts", len(parts))
	return parts, nil
}

func (mpc S3MultipartConfig) Done() error {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(mpc.S3Config.Region)},
	)
	svc := s3.New(sess)

	log.Printf("[Done] Completing Multipart Upload %+v", mpc.MultipartUploadId)

	uploadedParts, err := mpc.Parts()

	log.Printf("[Done] Found %+v uploaded parts", len(uploadedParts))

	if err != nil {
		log.Printf("Failed to complete Multipart Upload since the listPart operation failed with error: %+v", err)
		return err
	}

	if len(uploadedParts) == 0 {
		log.Printf("Failed to complete Multipart upload: no part uploaded. please abort")
		return errors.New("No part uploaded")
	}

	params := &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(mpc.S3Config.Bucket),   // Required
		Key:      aws.String(mpc.Key()),             // Required
		UploadId: aws.String(mpc.MultipartUploadId), // Required
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: uploadedParts,
		},
	}
	_, err2 := svc.CompleteMultipartUpload(params)

	if err2 != nil {
		log.Printf("Failed to complete Multipart Upload: %+v", err2)
		return err2
	}
	return nil
}

func (mpc S3MultipartConfig) Send(partNumber int64, payload []byte) error {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(mpc.S3Config.Region)},
	)
	svc := s3.New(sess)

	params := &s3.UploadPartInput{
		Bucket:     aws.String(mpc.S3Config.Bucket),   // Required
		Key:        aws.String(mpc.Key()),             // Required
		UploadId:   aws.String(mpc.MultipartUploadId), // Required
		PartNumber: aws.Int64(partNumber),             // Required
		Body:       bytes.NewReader(payload),
	}
	_, err := svc.UploadPart(params)

	if err != nil {
		log.Printf("Failed to send part number %v, error: %+v", partNumber, err)
		return err
	}

	return nil
}

func (mpc S3MultipartConfig) Upload(fileName string) {
	log.Printf("[Upload] Starting Upload of file %s", fileName)
	//measuring time
	start := time.Now()
	//0-Opening File
	file, err := os.Open(fileName)

	if err != nil {
		log.Printf("[Upload] error file opening the file: %+v", err)
		os.Exit(1)
	}
	defer file.Close()
	//0-Opening File
	fileInfo, _ := file.Stat()
	var fileSize int64 = fileInfo.Size()
	log.Printf("[Upload] Starting Upload of file %s of size %v", fileName, fileSize)
	fileChunk := mpc.ChunkSize * (1 << 20) // 1 MB, change this to your requirement
	// calculate total number of parts the file will be chunked into
	totalPartsNum := int64(math.Ceil(float64(fileSize) / float64(fileChunk)))
	chunked := core.ChunkArray(totalPartsNum, int64(math.Ceil(float64(totalPartsNum)/float64(mpc.NThreads))))
	log.Printf("[Upload] Splitting to %d pieces: %+v.\n", totalPartsNum, chunked)
	var wg sync.WaitGroup
	log.Printf("[Upload] Adding %+v thead to waitgroup.\n", len(chunked))
	wg.Add(len(chunked))
	for _, parts := range chunked {
		//async
		go func(parts []interface{}) {
			for _, part := range parts {
				partSize := int(math.Min(float64(fileChunk), float64(fileSize-int64(part.(int64)*fileChunk))))
				log.Printf("[Upload] Uploading part %v of size %v bytes", part, partSize)
				partBuffer := make([]byte, partSize)
				file.Read(partBuffer)
				//part number must be greater than 0
				mpc.Send(part.(int64)+1, partBuffer)
				log.Printf("[Upload] Uploading part %v Completed", part)
			}
			wg.Done()
		}(parts)

	}
	wg.Wait()
	elapsed := time.Since(start)
	log.Printf("[Upload] Upload of file %s (size %v) completed in %s", fileName, fileSize, elapsed)

}
