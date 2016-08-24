package common

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
)

// HttpLog satisfies the http.RoundTripper interface and is used to
// customize the default Gophercloud RoundTripper to allow for logging.
type HttpLog struct {
	Logger *logrus.Logger
	rt     http.RoundTripper
}

const httpTimeout = 15 * time.Second

// NewHTTPClient return a custom HTTP client that allows for logging relevant
// information before and after the HTTP request.
func NewHTTPClient() *http.Client {
	return &http.Client{
		Timeout: httpTimeout,
		Transport: &HttpLog{
			rt:     http.DefaultTransport,
			Logger: Log.Logger,
		},
	}
}

// RoundTrip performs a round-trip HTTP request and logs relevant information about it.
func (hl *HttpLog) RoundTrip(request *http.Request) (*http.Response, error) {
	defer func() {
		if request.Body != nil {
			request.Body.Close()
		}
	}()

	var err error

	if hl.Logger.Level == logrus.DebugLevel && request.Body != nil {
		request.Body, err = hl.logRequestBody(request.Body, request.Header)
		if err != nil {
			return nil, err
		}
	}

	hl.Logger.Debugf("Request: %s %s", request.Method, request.URL)

	response, err := hl.rt.RoundTrip(request)
	if response == nil {
		return nil, err
	}

	responseBody, _ := hl.logResponseBody(response.Body, response.Header)
	response.Body = responseBody

	if response.StatusCode >= 400 {
		buf := bytes.NewBuffer([]byte{})
		body, _ := ioutil.ReadAll(io.TeeReader(response.Body, buf))
		hl.Logger.Infof("Response Error: %+v", string(body))
		bufWithClose := ioutil.NopCloser(buf)
		response.Body = bufWithClose
	}

	return response, err
}

func (hl *HttpLog) logRequestBody(original io.ReadCloser, headers http.Header) (io.ReadCloser, error) {
	defer original.Close()

	var bs bytes.Buffer
	_, err := io.Copy(&bs, original)
	if err != nil {
		return nil, err
	}

	contentType := headers.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		debugInfo := hl.formatJSON(bs.Bytes())
		hl.Logger.Debugf("Request Options: %s", debugInfo)
	} else {
		hl.Logger.Debugf("Request Options: %s", bs.String())
	}

	return ioutil.NopCloser(strings.NewReader(bs.String())), nil
}

func (hl *HttpLog) logResponseBody(original io.ReadCloser, headers http.Header) (io.ReadCloser, error) {
	defer original.Close()

	var bs bytes.Buffer
	_, err := io.Copy(&bs, original)
	if err != nil {
		return nil, err
	}

	contentType := headers.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		debugInfo := hl.formatJSON(bs.Bytes())
		if debugInfo != "" {
			hl.Logger.Debugf("Response Body: %s", debugInfo)
		}
	} else {
		hl.Logger.Debugf("Not logging because response body isn't JSON")
	}

	return ioutil.NopCloser(strings.NewReader(bs.String())), nil
}

func (hl *HttpLog) formatJSON(raw []byte) string {
	var data map[string]interface{}

	err := json.Unmarshal(raw, &data)
	if err != nil {
		return string(raw)
	}

	pretty, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return string(raw)
	}

	return string(pretty)
}
