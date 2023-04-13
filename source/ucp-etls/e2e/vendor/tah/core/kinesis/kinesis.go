package kinesis

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"tah/core/core"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	awskinesis "github.com/aws/aws-sdk-go/service/kinesis"
)

var DEFAULT_BATCH_SIZE = 10
var WAIT_DELAY = 5
var ITERATOR_TYPE_AT_SEQUENCE_NUMBER = "AT_SEQUENCE_NUMBER"
var ITERATOR_TYPE_AFTER_SEQUENCE_NUMBER = "AFTER_SEQUENCE_NUMBER"
var ITERATOR_TYPE_TRIM_HORIZON = "TRIM_HORIZON"
var ITERATOR_TYPE_LATEST = "LATEST"
var ITERATOR_TYPE_AT_TIMESTAMP = "AT_TIMESTAMP"

type Config struct {
	Client         *awskinesis.Kinesis
	Region         string
	Stream         string
	BatchSize      int
	ShardCount     int64
	ShardIterators map[string]*string
}

type Record struct {
	Pk   string
	Data string
}

type IngestionError struct {
	ErrorCore      string
	ErrorMessage   string
	SequenceNumber string
	ShardId        string
}

type Stream struct {
	Shards    []Shard
	Arn       string
	CreatedOn time.Time
	Mode      string
	Status    string
}

type Shard struct {
	ID string
}

type IConfig interface {
	Create(streamName string) error
	Delete(streamName string) error
	WaitForStreamCreation(timeoutSeconds int) (string, error)
	Describe() (Stream, error)
	PutRecords(recs []Record) (error, []IngestionError)
	FetchRecords(timeoutSeconds int) ([]Record, error)
}

func Init(streamName, region string) Config {
	mySession := session.Must(session.NewSession())
	cfg := aws.NewConfig().WithRegion(region).WithMaxRetries(0)
	svc := awskinesis.New(mySession, cfg)
	return Config{
		Client:    svc,
		Region:    region,
		Stream:    streamName,
		BatchSize: DEFAULT_BATCH_SIZE,
	}
}

func InitAndCreate(streamName, region string) (Config, error) {
	cfg := Init(streamName, region)
	err := cfg.Create(streamName)
	return cfg, err
}

func (c *Config) Create(streamName string) error {
	in := &awskinesis.CreateStreamInput{
		StreamName: aws.String(c.Stream),
	}
	_, err := c.Client.CreateStream(in)
	if err == nil {
		c.Stream = streamName
	}
	return err
}

func (c *Config) WaitForStreamCreation(timeoutSeconds int) (string, error) {
	log.Printf("Waiting for crawler run to complete")
	stream, err := c.Describe()
	if err != nil {
		return "", err
	}
	it := 0
	for stream.Status != awskinesis.StreamStatusActive {
		log.Printf("Stream State: %v Waiting %v seconds before checking again", stream.Status, WAIT_DELAY)
		time.Sleep(time.Duration(WAIT_DELAY) * time.Second)
		stream, err = c.Describe()
		if err != nil {
			return "", err
		}
		it += 1
		if it*WAIT_DELAY >= timeoutSeconds {
			return "", errors.New(fmt.Sprintf("Stream wait timed out after %v secconds", it*WAIT_DELAY))
		}
	}
	log.Printf("Crawler State: %v. Completed", stream.Status)
	return stream.Status, err
}

func (c *Config) Describe() (Stream, error) {
	in := &awskinesis.DescribeStreamInput{
		StreamName: aws.String(c.Stream),
	}
	out, err := c.Client.DescribeStream(in)
	log.Printf("[Describe] DescribeStream output: %+v", out)
	return Stream{
		Shards:    parseShards(out.StreamDescription.Shards),
		Arn:       *out.StreamDescription.StreamARN,
		CreatedOn: *out.StreamDescription.StreamCreationTimestamp,
		Mode:      *out.StreamDescription.StreamModeDetails.StreamMode,
		Status:    *out.StreamDescription.StreamStatus,
	}, err
}

func parseShards(awsshards []*awskinesis.Shard) []Shard {
	shards := []Shard{}
	for _, awsShard := range awsshards {
		shards = append(shards, Shard{
			ID: *awsShard.ShardId,
		})
	}
	return shards
}

func (c *Config) Delete(streamName string) error {
	if c.Stream == "" {
		streamName = c.Stream
	}
	in := &awskinesis.DeleteStreamInput{
		StreamName: aws.String(c.Stream),
	}
	_, err := c.Client.DeleteStream(in)
	return err
}

func (c *Config) PutRecords(recs []Record) (error, []IngestionError) {
	log.Printf("[kinesis] putting %v records to Kinesis dara stream %v by batch of %v", len(recs), c.Stream, c.BatchSize)
	batchedRecs := core.Chunk(core.InterfaceSlice(recs), c.BatchSize)
	errs := []IngestionError{}
	for _, batch := range batchedRecs {
		in := &awskinesis.PutRecordsInput{
			StreamName: aws.String(c.Stream),
			Records:    []*awskinesis.PutRecordsRequestEntry{},
		}
		for _, record := range batch {
			in.Records = append(in.Records, &awskinesis.PutRecordsRequestEntry{
				Data:         []byte(record.(Record).Data),
				PartitionKey: aws.String(record.(Record).Pk),
			})
		}
		out, err := c.Client.PutRecords(in)
		if err != nil {
			log.Printf("[kinesis] error putting records to Kinesis dara stream %v: %v", c.Stream, err)
		}
		for _, failedRec := range out.Records {
			if failedRec.ErrorCode != nil {
				errs = append(errs, IngestionError{
					ErrorCore:      *failedRec.ErrorCode,
					ErrorMessage:   *failedRec.ErrorMessage,
					SequenceNumber: *failedRec.SequenceNumber,
					ShardId:        *failedRec.ShardId,
				})
			}
		}
	}
	if len(errs) > 0 {
		log.Printf("[kinesis] failed to ingest %v: records", len(errs))
		return errors.New("Failed to ingests " + strconv.Itoa(len(errs)) + " records"), errs
	}
	return nil, []IngestionError{}
}

func (c *Config) InitConsumer(iteratorType string) error {
	stream, err := c.Describe()
	if err != nil {
		return err
	}
	c.ShardIterators = map[string]*string{}
	for _, shard := range stream.Shards {
		out, err1 := c.Client.GetShardIterator(&awskinesis.GetShardIteratorInput{
			ShardId:           aws.String(shard.ID),
			ShardIteratorType: aws.String(iteratorType),
			StreamName:        aws.String(c.Stream),
		})
		if err1 != nil {
			return err1
		}
		c.ShardIterators[shard.ID] = out.ShardIterator
	}
	return nil
}

//do not use as a real consumer. Call InitConsumer() prior to writing to the stream and this function after
func (c Config) FetchRecords(timeoutSeconds int) ([]Record, error) {
	stream, err := c.Describe()
	recs := []Record{}
	if err != nil {
		return []Record{}, err
	}
	log.Printf("Iterating through all shards ")
	//TODO: parallelize this
	it := 0
	for it < timeoutSeconds {
		for _, shard := range stream.Shards {
			log.Printf("Reading Shard %+v", shard)
			records, err := c.Client.GetRecords(&kinesis.GetRecordsInput{
				ShardIterator: c.ShardIterators[shard.ID],
			})
			log.Printf("GetRecords output %+v", records)
			if err != nil {
				log.Printf("GetRecords error: %+v", err)
				return recs, err
			}
			for _, r := range records.Records {
				recs = append(recs, Record{
					Data: string(r.Data),
				})
			}
			c.ShardIterators[shard.ID] = records.NextShardIterator
		}
		time.Sleep(1 * time.Second)
		it++
	}
	log.Printf("Timeout of %+v seconds expired. returning %v records", timeoutSeconds, len(recs))
	return recs, nil
}
