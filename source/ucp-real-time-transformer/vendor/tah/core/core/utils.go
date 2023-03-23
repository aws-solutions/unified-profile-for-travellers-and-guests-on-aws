package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"hash/fnv"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"

	geohash "github.com/mmcloughlin/geohash"
)

var CLOUDRACK_DATEFORMAT string = "20060102"

func Hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func Chunk(array []interface{}, chunkSize int) [][]interface{} {
	var divided [][]interface{}
	for i := 0; i < len(array); i += chunkSize {
		end := i + chunkSize
		if end > len(array) {
			end = len(array)
		}
		divided = append(divided, array[i:end])
	}
	return divided
}

//Buld a double array of integer from a max value and a chunkSize
func ChunkArray(size int64, chunkSize int64) [][]interface{} {
	toChunk := make([]interface{}, 0, 0)
	for i := int64(0); i < size; i++ {
		toChunk = append(toChunk, i)
	}
	return Chunk(toChunk, int(chunkSize))
}

//Generats a usinque ID based timestamp
func GeneratUniqueId() string {
	return strconv.FormatInt(int64(Hash(time.Now().Format("20060102150405.000"))), 10)
}

func ParseEpochMs(ms string) (time.Time, error) {
	msInt, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(0, msInt*int64(time.Millisecond)), nil
}

//TODO: to evaluate th value of this tradeoff: this is probably a little slow but abstract the complexity for all uses of
//the save many function(and actually any core operation on array of interface)
func InterfaceSlice(slice interface{}) []interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		panic("InterfaceSlice() given a non-slice type")
	}

	ret := make([]interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret
}

/**********
*Geo Hasshing
**************/
func GeoHash(lat, lng float64) string {
	return geohash.Encode(lat, lng)
}
func GeoHashWithPrecision(lat, lng, radius float64) string {
	prec := builCaracterLength(radius)
	return geohash.EncodeWithPrecision(lat, lng, prec)
}

//geohasg precision: https://en.wikipedia.org/wiki/Geohash
//1   ±2500
//2   ±630
//3   ±78
//4   ±20
//5   ±2.4
//6   ±0.61
//7   ±0.076
//8   ±0.019
func builCaracterLength(radius float64) uint {
	precisions := []float64{2500, 630, 78, 20, 2.4, 0.61, 0.076, 0.019}
	ind := 0
	if radius >= precisions[ind] {
		return uint(ind + 1)
	}
	ind = ind + 1
	for ind < len(precisions)-2 {
		if radius >= precisions[ind+1] && radius < precisions[ind] {
			return uint(ind + 1)
		}
		ind = ind + 1
	}
	log.Printf("[UTILS] Geohash precision for %f KM is  %v \n", radius, ind)
	return uint(ind)
}

func GeterateDateRange(startDate string, endDate string) []string {
	start, _ := time.Parse(CLOUDRACK_DATEFORMAT, startDate)
	end, _ := time.Parse(CLOUDRACK_DATEFORMAT, endDate)
	diff := int(end.Sub(start).Hours() / 24.0)
	log.Printf("[UTILS] Diff between %v and %v => %+v \n", start, end, diff)
	dates := make([]string, 0, 0)
	for i := 0; i < diff; i++ {
		t := start.AddDate(0, 0, i)
		dates = append(dates, t.Format(CLOUDRACK_DATEFORMAT))
	}
	log.Printf("[UTILS] Generated date range from %s to %s => %+v \n", startDate, endDate, dates)
	return dates
}

/****************
* HTTP WRAPPERS
******************/

type HttpService struct {
	Endpoint string
	Tx       Transaction
	Headers  map[string]string
	Region   string
}

type RestOptions struct {
	SubEndpoint  string
	Headers      map[string]string
	SigV4Service string
	Debug        bool
}

func HttpInit(endpoint string) HttpService {
	log.Printf("[HTTP] initializing HTTP client with endpoint %v\n", endpoint)
	return HttpService{Endpoint: endpoint, Tx: NewTransaction("CORE", ""), Headers: map[string]string{}}
}
func HttpInitWithRegion(endpoint string, region string) HttpService {
	log.Printf("[HTTP] initializing HTTP client with endpoint %v\n", endpoint)
	return HttpService{Endpoint: endpoint, Tx: NewTransaction("CORE", ""), Headers: map[string]string{}, Region: region}
}

func (s HttpService) SetHeader(key string, val string) {
	s.Headers[key] = val
}

func (s HttpService) SetTx(tx Transaction) {
	tx.LogPrefix = "CORE"
	s.Tx = tx
}

func processResponse(resp *http.Response) (*http.Response, error) {
	if resp.StatusCode >= 400 {
		return resp, errors.New(resp.Status)
	}
	return resp, nil
}

func (s HttpService) HttpPut(id string, object interface{}, tpl JSONObject, options RestOptions) (interface{}, error) {
	endpoint := s.Endpoint
	if id != "" {
		endpoint = endpoint + "/" + id
	}
	s.Tx.Log("[HTTP] PUT %v\n", endpoint)
	s.Tx.Log("[HTTP] Body: %+v\n", object)

	var resp *http.Response
	bytesRepresentation, err := json.Marshal(object)
	if err != nil {
		log.Fatalln(err)
	}
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer(bytesRepresentation))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add(TRANSACTION_ID_HEADER, s.Tx.TransactionID)
	for headerName, headerValue := range s.Headers {
		req.Header.Add(headerName, headerValue)
	}
	for headerName, headerValue := range options.Headers {
		req.Header.Add(headerName, headerValue)
	}
	resp, err = client.Do(req)
	s.Tx.Log("[HTTP] PUT RESPONSE %v\n", resp)
	if err == nil && resp != nil {
		resp, err = processResponse(resp)
		res, _ := s.ParseBody(resp.Body, tpl)
		s.Tx.Log("[HTTP] PUT RESPONSE DECODED %v\n", res)
		return res, err
	}
	return tpl, err
}

func (s HttpService) HttpPost(object interface{}, tpl JSONObject, options RestOptions) (interface{}, error) {
	endpoint := s.Endpoint
	if options.SubEndpoint != "" {
		endpoint = endpoint + "/" + options.SubEndpoint
	}
	s.Tx.Log("[HTTP] POST %v\n", endpoint)
	s.Tx.Log("[HTTP] Body: %+v\n", object)

	var resp *http.Response
	bytesRepresentation, err := json.Marshal(object)
	s.Tx.Log("[HTTP][DEBUG] POST REQUEST Body Json %+v\n", string(bytesRepresentation))
	if err != nil {
		log.Fatalln(err)
	}
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(bytesRepresentation))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add(TRANSACTION_ID_HEADER, s.Tx.TransactionID)
	for headerName, headerValue := range s.Headers {
		req.Header.Add(headerName, headerValue)
	}
	for headerName, headerValue := range options.Headers {
		req.Header.Set(headerName, headerValue)
	}
	s.Tx.Log("[HTTP][DEBUG] POST REQUEST Headers %+v\n", req.Header)

	if options.SigV4Service != "" {
		s.Tx.Log("[HTTP][SIGV4] Adding SigV4")
		creds := credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY"), os.Getenv("AWS_SECRET_KEY"), os.Getenv("AWS_SESSION_TOKEN"))
		s.Tx.Log("[HTTP][SIGV4] creds:  %+v\n", creds)
		signer := v4.NewSigner(creds)
		s.Tx.Log("[HTTP][SIGV4] signer:  %+v\n", signer)
		bodyPayload := strings.NewReader(string(bytesRepresentation))
		signer.Sign(req, bodyPayload, options.SigV4Service, s.Region, time.Now())
	}

	resp, err = client.Do(req)
	s.Tx.Log("[HTTP] POST RESPONSE %v\n", resp)
	if err == nil && resp != nil {
		if options.Debug {
			b, _ := ioutil.ReadAll(resp.Body)
			s.Tx.Log("[HTTP][DEBUG] POST RESPONSE Body %v\n", string(b))
			return tpl, err
		}
		resp, err = processResponse(resp)
		res, _ := s.ParseBody(resp.Body, tpl)
		s.Tx.Log("[HTTP] POST RESPONSE DECODED %v\n", res)
		return res, err
	}
	return tpl, err
}

func (s HttpService) HttpGet(params map[string]string, tpl JSONObject, options RestOptions) (interface{}, error) {
	endpoint := s.Endpoint
	if options.SubEndpoint != "" {
		endpoint = endpoint + "/" + options.SubEndpoint
	}
	s.Tx.Log("[HTTP] GET %v\n", s.Endpoint)
	var resp *http.Response
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	req.Header.Add(TRANSACTION_ID_HEADER, s.Tx.TransactionID)
	for headerName, headerValue := range s.Headers {
		req.Header.Add(headerName, headerValue)
	}
	for headerName, headerValue := range options.Headers {
		req.Header.Add(headerName, headerValue)
	}
	if err != nil {
		log.Fatalln(err)
	}
	//Adding params
	q := req.URL.Query()
	for key, val := range params {
		q.Add(key, val)
	}
	req.URL.RawQuery = q.Encode()
	//Workaround: uncode comma to support comma separated params
	req.URL.RawQuery = strings.ReplaceAll(req.URL.RawQuery, "%2C", ",")
	s.Tx.Log("[HTTP] GET REQUEST %v\n", req)
	resp, err = client.Do(req)
	s.Tx.Log("[HTTP] GET RESPONSE %v\n", resp)
	if err == nil {
		resp, err = processResponse(resp)

		res, _ := s.ParseBody(resp.Body, tpl)
		s.Tx.Log("[HTTP] GET RESPONSE DECODED %v\n", res)
		return res, err
	}
	return tpl, err
}

func (s HttpService) ParseBody(body io.ReadCloser, p JSONObject) (interface{}, error) {
	s.Tx.Log("[HTTP] parsing json body into %v", reflect.TypeOf(p))
	dec := json.NewDecoder(body)
	var err error = nil
	for {
		//if err = dec.Decode(&p); err == io.EOF {
		if err, p = p.Decode(*dec); err == io.EOF {
			break
		} else if err != nil {
			break
		}
	}
	if err == io.EOF {
		err = nil
	}
	return p, err
}

////////////
//Parsing HTTP Params
//////////////////
//to move to corre
func ParseIntParam(param string) int64 {
	i, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		log.Printf("[WARNING] invalid query sting parametter %v, should be parsable into integer", param)
	}
	return i
}

//type of object that is manageable by a micro service
type JSONObject interface {
	//ParseBody(body io.ReadCloser) (interface{}, error)
	Decode(dec json.Decoder) (error, JSONObject)
}

/********************
* Utility Functions
*********************/

func IsEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

/********************
* Pointer Converters
*********************/

func ToMapString(in map[string]*string) map[string]string {
	out := map[string]string{}
	for key, val := range in {
		out[key] = PtToString(val)
	}
	return out
}

func ToMapPtString(in map[string]string) map[string]*string {
	out := map[string]*string{}
	for key, val := range in {
		out[key] = &val
	}
	return out
}

func PtToString(in *string) string {
	if in != nil {
		return *in
	}
	return ""
}

func PtToInt64(in *int64) int64 {
	if in != nil {
		return *in
	}
	return 0
}

func PtToFloat64(in *float64) float64 {
	if in != nil {
		return *in
	}
	return float64(0.0)
}
