package s3

import (
	"log"
	"os"
	"strconv"
	"testing"
	"time"
)

var env = os.Getenv("TAH_CORE_ENV")
var TAH_CORE_REGION = os.Getenv("TAH_CORE_REGION")

/*******
* S3
*****/

func TestCreateDeleteBucket(t *testing.T) {
	s3c := InitRegion(TAH_CORE_REGION)
	name, err := s3c.CreateRandomBucket("tah-core-test-bucket")
	if err != nil {
		t.Errorf("[TestCreateDeleteBucket] Failed create %+v", err)
	}
	resource := []string{"arn:aws:s3:::" + name, "arn:aws:s3:::" + name + "/*"}
	actions := []string{"s3:PutObject", "s3:ListBucket", "s3:GetObject", "s3:GetBucketLocation", "s3:GetBucketPolicy"}
	principal := map[string][]string{"Service": []string{"appflow.amazonaws.com"}}
	err = s3c.AddPolicy(name, resource, actions, principal)
	if err != nil {
		t.Errorf("[TestCreateDeleteBucket] error adding bucket policy %+v", err)
	}
	err = s3c.DeleteBucket(name)
	if err != nil {
		t.Errorf("[TestCreateDeleteBucket] Failed delete %+v", err)
	}
}

func TestParseCSV(t *testing.T) {
	log.Printf("Testing CSV parsing")
	s3c, err := InitWithRandBucket("tah-core-test-bucket-csv", "test_csv", TAH_CORE_REGION)
	if err != nil {
		t.Errorf("Could not initialize random bucket %+v", err)
	}
	log.Printf("Uploading CSV ../../test_assets/csv1.csv")
	err = s3c.UploadFile("csv1.csv", "../../test_assets/csv1.csv")
	if err != nil {
		t.Errorf("[TestParseCSV] Failed to upload CSV file: %+v", err)
	}
	log.Printf("Waiting 5 seconds")
	time.Sleep(5 * time.Second)
	log.Printf("Downloading and parsing file")
	data, err1 := s3c.ParseCsvFromS3("csv1.csv")
	if err1 != nil {
		t.Errorf("[TestParseCSV] Failed to fetch CSV file: %+v", err1)
	}
	if len(data) == 0 {
		t.Errorf("[TestParseCSV] Failed to fetch CSV file: file should have 9 rows")
	}
	err = s3c.EmptyAndDelete()
	if err != nil {
		t.Errorf("[TestParseCSV] Failed to empty and delete bucket %+v", err)
	}
}

func TestList(t *testing.T) {
	log.Printf("Uploading CSV ../../test_assets/csv1.csv")
	s3c, err := InitWithRandBucket("tah-core-test-bucket-csv", "test_csv", TAH_CORE_REGION)
	if err != nil {
		t.Errorf("Could not initialize random bucket %+v", err)
	}
	err = s3c.UploadFile("csv1.csv", "../../test_assets/csv1.csv")
	if err != nil {
		t.Errorf("[TestParseCSV] Failed to upload CSV file: %+v", err)
	}
	log.Printf("Searching bucket content")
	res, err := s3c.Search("", 10)
	log.Printf("bucket content: %+v", res)
	if err != nil {
		t.Errorf("[TestList] List objects in bucket: %+v", err)
	}
	if len(res) == 0 {
		t.Errorf("[TestList] Bucket should have at least one object")
	}
	hasIt := false
	for _, val := range res {
		if val == "test_csv/csv1.csv" {
			hasIt = true
		}
	}
	if !hasIt {
		t.Errorf("[TestList] bucket should have object test_csv/csv1.csv")
	}
	err = s3c.EmptyAndDelete()
	if err != nil {
		t.Errorf("[TestParseCSV] Failed to empty and delete bucket %+v", err)
	}

}

func TestSaveMany(t *testing.T) {
	log.Printf("TestSaveMany")
	now := "_" + time.Now().Format("2006-01-02-15-04-05")

	s3c, err := InitWithRandBucket("tah-core-test-bucket-csv", "test_mass_upload", TAH_CORE_REGION)
	if err != nil {
		t.Errorf("Could not initialize random bucket %+v", err)
	}

	nItems := 10000
	paths := make([]string, nItems, nItems)
	data := make([]string, nItems, nItems)
	for i := 0; i < nItems; i++ {
		iStr := strconv.Itoa(i)
		paths[i] = now + "-" + iStr + ".json"
		data[i] = "{\"id\" : " + iStr + "}"
	}

	errs := s3c.SaveMany(paths, data, "application/json", 200)
	if len(errs) > 0 {
		t.Errorf("[TestSaveMany] multiple errors: %+v", errs)
	}
	//testing search function
	res, err1 := s3c.Search("", 100)
	if err1 != nil {
		t.Errorf("[TestSaveMany] error listing objects in bucket %+v", err1)
	}
	if len(res) != 100 {
		t.Errorf("[TestSaveMany] Search funvtino should return %+v ojects and not %v", 100, len(res))
	}
	errs = s3c.DeleteMany(paths, 200)
	if len(errs) > 0 {
		t.Errorf("[TestDeleteMany] multiple errors: %+v", errs)
	}

	err = s3c.EmptyAndDelete()
	if err != nil {
		t.Errorf("[TestParseCSV] Failed to empty and delete bucket %+v", err)
	}
}

func TestS3MultiPart(t *testing.T) {
	PATH := "test_multipart"
	s3Config, err0 := InitWithRandBucket("tah-core-test-bucket-csv", PATH, TAH_CORE_REGION)
	if err0 != nil {
		t.Errorf("Could not initialize random bucket %+v", err0)
	}

	key := "mpUploadTest"
	mpc, err := s3Config.StartMPUpload(key)
	if err != nil {
		t.Errorf("[StartMpUpload] Failed to start multipart Upload with erro: %+v", err)
	}
	if mpc.MultipartUploadId == "" {
		t.Errorf("[StartMpUpload]Should return Upload Id ")
	}
	if mpc.Key() != PATH+"/"+key {
		t.Errorf("[Key]Should return %v and NOT %v", PATH+"/"+key, mpc.Key())
	}
	//LISTING MULTIPARTID
	mpcs, err := s3Config.ListMpUploads()
	if err != nil {
		t.Errorf("[ListMpUploads]Error while listing multipart uploads: %+v", err)
	}
	exists := false
	for _, upload := range mpcs {
		if upload.MultipartUploadId == mpc.MultipartUploadId {
			exists = true
		}
	}
	if !exists {
		t.Errorf("[ListMpUploads]Should find %v in list: %+v", mpc.MultipartUploadId, mpcs)
	}
	//Adding Parts
	partNum := int64(990)
	mpc.Send(partNum, []byte("testOfUpload"))
	parts, err := mpc.Parts()
	if err != nil {
		t.Errorf("[Parts]Error while listing parts for upload %+v: %+v", mpc, err)
	}
	exists = false
	for _, part := range parts {
		if *part.PartNumber == partNum {
			exists = true
		}
	}
	if !exists {
		t.Errorf("[Parts]Part with nnumber %v no found after upload %+v ", partNum, mpc)
	}

	//Comleting Multipart Upload
	mpc.Done()
	mpcs, err = s3Config.ListMpUploads()
	if err != nil {
		t.Errorf("[ListMpUploads]Error while listing multipart uploads: %+v", err)
	}
	exists = false
	for _, upload := range mpcs {
		if upload.MultipartUploadId == mpc.MultipartUploadId {
			exists = true
		}
	}
	if exists {
		t.Errorf("[ListMpUploads]Should NOT find %v in list: %+v as it has been completed", mpc.MultipartUploadId, mpcs)
	}

	err = s3Config.EmptyAndDelete()
	if err != nil {
		t.Errorf("[TestParseCSV] Failed to empty and delete bucket %+v", err)
	}
}

func TestPresignUrl(t *testing.T) {
	s3c, err := InitWithRandBucket("tah-core-test-bucket-csv", "test_presign", TAH_CORE_REGION)
	if err != nil {
		t.Errorf("Could not initialize random bucket %+v", err)
	}
	key := "mpUploadTest"
	url, err := s3c.CreatePresignedUrl(key, 1)
	if err != nil {
		t.Errorf("[TestPresignUrl]Error while creating presign URL: %+v", err)
	}
	if url == "" {
		t.Errorf("[TestPresignUrl] Failed to create presigned URL, no URL returned")
	}
	err = s3c.EmptyAndDelete()
	if err != nil {
		t.Errorf("[TestParseCSV] Failed to empty and delete bucket %+v", err)
	}
}
