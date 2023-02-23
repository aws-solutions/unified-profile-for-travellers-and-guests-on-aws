    
// Package sigv4 signs HTTP requests as prescribed in
// http://docs.amazonwebservices.com/general/latest/gr/signature-version-4.html
//courtesy of https://github.com/bmizerany/aws4
package sigv4

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"os"
	"log"
)

const iSO8601BasicFormat = "20060102T150405Z"
const iSO8601BasicFormatShort = "20060102"

var lf = []byte{'\n'}

func Post(urlVal string, bodyType string, body io.Reader) (resp *http.Response, err error){
	req, err := http.NewRequest("POST", urlVal, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", bodyType)
	req.Header.Set("Host", "a4dqpl7ml7.execute-api.us-east-1.amazonaws.com")
	Sign(req)
	log.Printf("[sigv4] signed request %v",req)
	return http.DefaultClient.Do(req)
}

//testing sigv4 algo
func Test(){
	sv := new(Service)
	sv.Name = "service"
	sv.Region = "us-east-1"
	req, _ := http.NewRequest("GET", "/?Param2=value2&Param1=value1", nil)
	req.Header.Set("Host", "example.amazonaws.com")
	req.Header.Set("X-Amz-Date", "20150830T123600Z")
	key := Keys{
		AccessKey : "AKIDEXAMPLE",
		SecretKey : "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY",
	}
	sv.Sign(&key,req)
	log.Printf("[sigv4] TEST signed request %+v. Expected: %v",req,"b97d918cfa904a5beff61c982a1b6f458b799221646efd99d3219ec94cdf2500")
}

func KeysFromEnvironment() *Keys {
	return &Keys{
		AccessKey: os.Getenv("AWS_ACCESS_KEY"),
		SecretKey: os.Getenv("AWS_SECRET_KEY"),
	}
}


// Keys holds a set of Amazon Security Credentials.
type Keys struct {
	AccessKey string
	SecretKey string
}

func (k *Keys) sign(s *Service, t time.Time) []byte {
	h := ghmac([]byte("AWS4"+k.SecretKey), []byte(t.Format(iSO8601BasicFormatShort)))
	h = ghmac(h, []byte(s.Region))
	h = ghmac(h, []byte(s.Name))
	h = ghmac(h, []byte("aws4_request"))
	log.Printf("[sigv4] Signing Key %x",h )
	return h
}

// Service represents an AWS-compatible service.
type Service struct {
	// Name is the name of the service being used (i.e. iam, etc)
	Name string

	// Region is the region you want to communicate with the service through. (i.e. us-east-1)
	Region string
}

// Sign signs a request with a Service derived from r.Host
func Sign(r *http.Request) error {
	parts := strings.Split(r.Host, ".")
	if len(parts) < 4 {
		return fmt.Errorf("Invalid AWS Endpoint: %s", r.Host)
	}
	sv := new(Service)
	sv.Name = parts[1]
	sv.Region = parts[2]
	sv.Sign(KeysFromEnvironment(), r)
	return nil
}

// Sign signs an HTTP request with the given AWS keys for use on service s.
func (s *Service) Sign(keys *Keys, r *http.Request) error {
	log.Printf("[sigv4] keys %+v",keys)

	date := r.Header.Get("X-Amz-Date")
	t := time.Now().UTC()
	if date != "" {
		var err error
		t, err = time.Parse("20060102T150405Z", date)
		if err != nil {
			log.Printf("[sigv4] invalid time format in header X-Amz-Date : %v with error : %+v",date, err)
			return err
		}
	} else {
		r.Header.Set("X-Amz-Date", t.Format(iSO8601BasicFormat))
	}

	k := keys.sign(s, t)
	h := hmac.New(sha256.New, k)
	s.writeStringToSign(h, t, r)

	auth := bytes.NewBufferString("AWS4-HMAC-SHA256 ")
	auth.Write([]byte("Credential=" + keys.AccessKey + "/" + s.creds(t)))
	auth.Write([]byte{',', ' '})
	auth.Write([]byte("SignedHeaders="))
	s.writeHeaderList(auth, r)
	auth.Write([]byte{',', ' '})
	auth.Write([]byte("Signature=" + fmt.Sprintf("%x", h.Sum(nil))))

	r.Header.Set("Authorization", auth.String())

	return nil
}

func (s *Service) writeQuery(w io.Writer, r *http.Request) {
	var a []string
	for k, vs := range r.URL.Query() {
		k = url.QueryEscape(k)
		for _, v := range vs {
			if v == "" {
				a = append(a, k)
			} else {
				v = url.QueryEscape(v)
				a = append(a, k+"="+v)
			}
		}
	}
	sort.Strings(a)
	b := make([]byte,0,0)
	for i, s := range a {
		if i > 0 {
			b = append(b,[]byte{'&'}...)
		}
		b = append(b,[]byte(s)...)
	}
	log.Printf("[sigv4] writeQuery:  %s",b)
	w.Write(b)
}

func (s *Service) writeHeader(w io.Writer, r *http.Request) {
	i, a := 0, make([]string, len(r.Header))
	for k, v := range r.Header {
		sort.Strings(v)
		a[i] = strings.ToLower(k) + ":" + strings.Join(v, ",")
		i++
	}
	sort.Strings(a)
	b := make([]byte,0,0)
	for i, s := range a {
		if i > 0 {
			b = append(b,lf...)
		}
		b = append(b,[]byte(s)...)
	}
	log.Printf("[sigv4] writeHeader:  %s",b)
	w.Write(b)
}

func (s *Service) writeHeaderList(w io.Writer, r *http.Request) {
	i, a := 0, make([]string, len(r.Header))
	for k, _ := range r.Header {
		a[i] = strings.ToLower(k)
		i++
	}
	sort.Strings(a)
	b := make([]byte,0,0)
	for i, s := range a {
		if i > 0 {
			b = append(b,[]byte{';'}...)
		}
		b = append(b,[]byte(s)...)
	}
	log.Printf("[sigv4] writeHeaderList  %s",b)
	w.Write(b)
}

func (s *Service) writeBody(w io.Writer, r *http.Request) {
	var b []byte
	// If the payload is empty, use the empty string as the input to the SHA256 function
	// http://docs.amazonwebservices.com/general/latest/gr/sigv4-create-canonical-request.html
	if r.Body == nil {
		b = []byte("")
	} else {
		var err error
		b, err = ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		r.Body = ioutil.NopCloser(bytes.NewBuffer(b))
	}
	log.Printf("[sigv4] writeBody   %s",b)
	h := sha256.New()
	h.Write(b)
	log.Printf("[sigv4] writeBody hashed   %x",h.Sum(nil))
	fmt.Fprintf(w, "%x", h.Sum(nil))
}

func (s *Service) writeURI(w io.Writer, r *http.Request) {
	path := r.URL.RequestURI()
	if r.URL.RawQuery != "" {
		path = path[:len(path)-len(r.URL.RawQuery)-1]
	}
	slash := strings.HasSuffix(path, "/")
	path = filepath.Clean(path)
	if path != "/" && slash {
		path += "/"
	}
	log.Printf("[sigv4] writeURI   %s",path)
	w.Write([]byte(path))
}

func (s *Service) writeRequest(w io.Writer, r *http.Request) {
	b :=[]byte(r.Method)
	b = append(b,lf...)
	log.Printf("[sigv4] method  %s",b)
	w.Write([]byte(r.Method))
	w.Write(lf)
	s.writeURI(w, r)
	w.Write(lf)
	s.writeQuery(w, r)
	w.Write(lf)
	s.writeHeader(w, r)
	w.Write(lf)
	w.Write(lf)
	s.writeHeaderList(w, r)
	w.Write(lf)
	s.writeBody(w, r)
}

func (s *Service) writeStringToSign(w io.Writer, t time.Time, r *http.Request) {
	b :=[]byte("AWS4-HMAC-SHA256")
	b = append(b,lf...)
	b = append(b,[]byte(t.Format(iSO8601BasicFormat))...)
	b = append(b,lf...)
	b = append(b,[]byte(s.creds(t))...)
	b = append(b,lf...)
	log.Printf("[sigv4] writeStringToSign   %s",b)
	w.Write(b)
	/*w.Write([]byte("AWS4-HMAC-SHA256"))
	w.Write(lf)
	w.Write([]byte(t.Format(iSO8601BasicFormat)))
	w.Write(lf)
	w.Write([]byte(s.creds(t)))
	w.Write(lf)*/

	h := sha256.New()
	s.writeRequest(h, r)
	log.Printf("[sigv4] Canonical request hash   %x",h.Sum(nil))
	fmt.Fprintf(w, "%x", h.Sum(nil))
}

func (s *Service) creds(t time.Time) string {
	return t.Format(iSO8601BasicFormatShort) + "/" + s.Region + "/" + s.Name + "/aws4_request"
}

func ghmac(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}