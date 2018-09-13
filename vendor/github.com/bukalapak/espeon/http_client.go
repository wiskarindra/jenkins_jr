// Package espeon provides HTTP client that implements Bukalapak SRE standards.
package espeon

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const (
	// DefaultMaxRetry is the default value for max retry.
	DefaultMaxRetry = 0
	// DefaultTimeout is the default value for connection timeout
	DefaultTimeout = 3 * time.Second
)

// Client is an interface for HTTP.
type Client interface {
	// Do does exactly as http.Client.Do() does.
	// It sends an HTTP request and returns an HTTP response and an error.
	Do(req *http.Request) (resp *http.Response, err error)
	// Get sends an HTTP GET request.
	// It needs an URL and a specific HTTP header to do its job.
	Get(url string, header http.Header) (resp *http.Response, err error)
	// Post sends an HTTP POST request.
	// It needs an URL, a specific HTTP header, and data to be sent.
	Post(url string, header http.Header, body io.Reader) (resp *http.Response, err error)
	// Put sends an HTTP PUT request.
	// It needs an URL, a specific HTTP header, and data to be sent.
	Put(url string, header http.Header, body io.Reader) (resp *http.Response, err error)
	// Patch sends an HTTP PATCH request.
	// It needs an URL, a specific HTTP header, and data to be sent.
	Patch(url string, header http.Header, body io.Reader) (resp *http.Response, err error)
	// Delete sends an HTTP DELETE request.
	// It needs an URL and a specific HTTP header to do its job.
	Delete(url string, header http.Header) (resp *http.Response, err error)
}

// HTTPClient is a struct that implements Client interface.
// It is used to connect to HTTP.
type HTTPClient struct {
	client *http.Client
	// Backoff is used to define the interval to retry request.
	Backoff Backoff
	// MaxRetry indicates how many times should a request be retried when fail.
	MaxRetry int
}

// NewDefaultHTTPClient creates an instance of HTTPClient with default values.
// Its timeout is DefaultTimeout.
// Default value for MaxRetry is DefaultMaxRetry.
// Default value for Backoff is ExponentialBackoff.
func NewDefaultHTTPClient() *HTTPClient {
	client := &http.Client{
		Timeout: DefaultTimeout,
	}

	return &HTTPClient{
		client:   client,
		Backoff:  NewExponentialBackoff(),
		MaxRetry: DefaultMaxRetry,
	}
}

// NewCustomHTTPClient creates an instance of HTTPClient with all value provided by caller.
func NewCustomHTTPClient(timeout time.Duration, backoff Backoff, maxRetry int) *HTTPClient {
	client := &http.Client{
		Timeout: timeout,
	}

	return &HTTPClient{
		client:   client,
		Backoff:  backoff,
		MaxRetry: maxRetry,
	}
}

// Do receives http.Request and sends it over HTTP then returns the response.
// It will retry sending request if HTTP response status code is 5xx.
func (h *HTTPClient) Do(req *http.Request) (resp *http.Response, err error) {
	req.Close = true

	// make sure header has request ID
	reqID := req.Header.Get("X-Request-ID")
	if reqID == "" {
		req.Header.Set("X-Request-ID", createRequestID())
	}

	for retry := 0; retry <= h.MaxRetry; retry++ {
		// set retry value in header
		// based on https://bukalapak.atlassian.net/wiki/spaces/INF/pages/119275619/RFC+Request+Context
		req.Header.Set("X-Retry", strconv.Itoa(retry))

		// why does it call time.Sleep() here?
		// note that the procedure should be like this:
		// 1. send request
		// 2. if transient error happens, retry request with non-zero backoff
		// 3. retry will be done if and only if retry count < max retry
		// so, when retry == 0, it should do step 1
		backoffTime := h.Backoff.NextInterval(retry)
		time.Sleep(backoffTime)

		resp, err = h.client.Do(req)
		if err != nil || resp.StatusCode >= 500 {
			continue
		}
		break
	}

	return
}

// Get receives an URL and http.Header then makes HTTP GET request.
// Basically, it will call HTTPClient.Do(req *http.Request) method.
func (h *HTTPClient) Get(url string, header http.Header) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return resp, errors.Wrap(err, "Couldn't create HTTP Request for Get method in HTTPClient")
	}

	if header == nil {
		header = make(http.Header)
	}
	req.Header = header
	return h.Do(req)
}

// Post receives an URL, http.Header, and io.Reader then makes HTTP POST request.
// Basically, it will call HTTPClient.Do(req *http.Request) method.
func (h *HTTPClient) Post(url string, header http.Header, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return resp, errors.Wrap(err, "Couldn't create HTTP Request for Post method in HTTPClient")
	}

	if header == nil {
		header = make(http.Header)
	}
	req.Header = header
	return h.Do(req)
}

// Put receives an URL, http.Header, and io.Reader then makes HTTP PUT request.
// Basically, it will call HTTPClient.Do(req *http.Request) method.
func (h *HTTPClient) Put(url string, header http.Header, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		return resp, errors.Wrap(err, "Couldn't create HTTP Request for Put method in HTTPClient")
	}

	if header == nil {
		header = make(http.Header)
	}
	req.Header = header
	return h.Do(req)
}

// Patch receives an URL, http.Header, and io.Reader then makes HTTP PATCH request.
// Basically, it will call HTTPClient.Do(req *http.Request) method.
func (h *HTTPClient) Patch(url string, header http.Header, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodPatch, url, body)
	if err != nil {
		return resp, errors.Wrap(err, "Couldn't create HTTP Request for Patch method in HTTPClient")
	}

	if header == nil {
		header = make(http.Header)
	}
	req.Header = header
	return h.Do(req)
}

// Delete receives an URL and http.Header then makes HTTP DELETE request.
// Basically, it will call HTTPClient.Do(req *http.Request) method.
func (h *HTTPClient) Delete(url string, header http.Header) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return resp, errors.Wrap(err, "Couldn't create HTTP Request for Delete method in HTTPClient")
	}

	if header == nil {
		header = make(http.Header)
	}
	req.Header = header
	return h.Do(req)
}

func createRequestID() string {
	uid, err := uuid.NewRandom()
	if err != nil {
		return ""
	}
	return uid.String()
}
