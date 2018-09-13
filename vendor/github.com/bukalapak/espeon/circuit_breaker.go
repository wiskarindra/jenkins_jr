package espeon

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/pkg/errors"
)

const (
	// DefaultErrorPercentThreshold defines default value for error percent rate threshold
	DefaultErrorPercentThreshold = 10
	// DefaultRequestVolumeThreshold defines default value for minimum number of request
	DefaultRequestVolumeThreshold = 1000
	// DefaultSleepWindow defines default value for circuit before it tries for recovery
	DefaultSleepWindow = 3000
	// DefaultMaxConcurrentRequest defines maximum number of concurrent processes that are allowed to happen at the same time
	DefaultMaxConcurrentRequest = 1000
	circuitOpen                 = "open"
	circuitClose                = "close"
)

// CircuitBreakerConfig contains any configs needed to run a circuit breaker.
// Based on SRE standard, we use Hystrix as our circuit breaker. Please, read more about Hystrix in the following link https://github.com/Netflix/Hystrix/wiki.
//
// In Espeon, we choose to use https://github.com/afex/hystrix-go as our dependency for hystrix-like circuit breaker.
// Therefore, this struct is only a bridge to config the dependency.
type CircuitBreakerConfig struct {
	// Name defines the name of circuit breaker.
	Name string
	// ErrorPercentThreshold defines how many errors should be tolerated before the circuit goes to open state.
	// In other words, once the threshold is surpassed, the circuit will be in open state.
	ErrorPercentThreshold int
	// RequestVolumeThreshold defines minimum number of request before circuit can be tripped.
	RequestVolumeThreshold int
	// SleepWindow defines how long, in millisecond, a circuit will ignore the request when in open state. I
	SleepWindow int
	// MaxConcurrentRequest defines how many concurrent processes allowed to happen at the same time.
	MaxConcurrentRequest int
	// FallbackFunc defines fallback function that will be executed when the request fails.
	FallbackFunc func(err error) error
}

// NewDefaultHystrixConfig return an instance of CircuitBreakerConfig with default values.
//
// The default values are:
//
// - ErrorPercentThreshold: DefaultErrorPercentThreshold
//
// - RequestVolumeThreshold: DefaultRequestVolumeThreshold
//
// - SleepWindow: DefaultSleepWindow
//
// - MaxConcurrentRequest: DefaultMaxConcurrentRequest
//
// - Fallback: nil
func NewDefaultHystrixConfig(name string) CircuitBreakerConfig {
	return CircuitBreakerConfig{
		Name: name,
		ErrorPercentThreshold:  DefaultErrorPercentThreshold,
		RequestVolumeThreshold: DefaultRequestVolumeThreshold,
		SleepWindow:            DefaultSleepWindow,
		MaxConcurrentRequest:   DefaultMaxConcurrentRequest,
		FallbackFunc:           nil,
	}
}

// HystrixHTTPClient implements Client behaviors and circuit breaker.
type HystrixHTTPClient struct {
	client *http.Client
	config CircuitBreakerConfig
	// Backoff is used to define the interval to retry request.
	Backoff Backoff
	// MaxRetry indicates how many times should a request be retried when fail.
	MaxRetry int
}

// NewDefaultHystrixHTTPClient creates an instance of HystrixHTTPClient with default values.
// Its timeout is DefaultTimeout.
// Its circuit breaker config is NewDefaultHystrixConfig(name string)
// Default value for MaxRetry is DefaultMaxRetry.
// Default value for Backoff is ExponentialBackoff.
func NewDefaultHystrixHTTPClient(name string) *HystrixHTTPClient {
	client := &http.Client{
		Timeout: DefaultTimeout,
	}

	config := NewDefaultHystrixConfig(name)
	configureHystrix(config, int(DefaultTimeout/time.Millisecond))

	return &HystrixHTTPClient{
		client:   client,
		config:   config,
		Backoff:  NewExponentialBackoff(),
		MaxRetry: DefaultMaxRetry,
	}
}

// NewCustomHystrixHTTPClient creates an instance of HystrixHTTPClient with all value provided by caller.
func NewCustomHystrixHTTPClient(timeout time.Duration, backoff Backoff, maxRetry int, config CircuitBreakerConfig) *HystrixHTTPClient {
	client := &http.Client{
		Timeout: timeout,
	}

	configureHystrix(config, int(timeout/time.Millisecond))

	return &HystrixHTTPClient{
		client:   client,
		config:   config,
		Backoff:  backoff,
		MaxRetry: maxRetry,
	}
}

// Do receives http.Request and sends it over HTTP then returns the response.
// It will retry sending request if HTTP response status code is 5xx.
func (h *HystrixHTTPClient) Do(req *http.Request) (resp *http.Response, err error) {
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

		// when retry == 0 (initial request), backoffTime will always be 0 seconds
		// so, it does make sense to put time.Sleep() code here
		backoffTime := h.Backoff.NextInterval(retry)
		time.Sleep(backoffTime)

		circuit, _, _ := hystrix.GetCircuit(h.config.Name)
		if circuit.IsOpen() {
			counter.WithLabelValues("espeon-circuit-"+h.config.Name, circuitOpen).Inc()
		} else {
			counter.WithLabelValues("espeon-circuit-"+h.config.Name, circuitClose).Inc()
		}

		// implement circuit breaker strategy
		err = hystrix.Do(h.config.Name, func() error {
			resp, err = h.client.Do(req)
			if err != nil {
				return err
			}
			if resp.StatusCode >= http.StatusInternalServerError {
				msg := fmt.Sprintf("Error occured on server. Status code: %d", resp.StatusCode)
				return errors.New(msg)
			}
			return nil
		}, h.config.FallbackFunc)

		if err != nil {
			continue
		}
		break
	}

	return
}

// Get receives an URL and http.Header then makes HTTP GET request.
// Basically, it will call HystrixHTTPClient.Do(req *http.Request) method.
func (h *HystrixHTTPClient) Get(url string, header http.Header) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return resp, errors.Wrap(err, "Couldn't create HTTP Request for Get method in HystrixHTTPClient")
	}

	if header == nil {
		header = make(http.Header)
	}
	req.Header = header
	return h.Do(req)
}

// Post receives an URL, http.Header, and io.Reader then makes HTTP POST request.
// Basically, it will call HystrixHTTPClient.Do(req *http.Request) method.
func (h *HystrixHTTPClient) Post(url string, header http.Header, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return resp, errors.Wrap(err, "Couldn't create HTTP Request for Post method in HystrixHTTPClient")
	}

	if header == nil {
		header = make(http.Header)
	}
	req.Header = header
	return h.Do(req)
}

// Put receives an URL, http.Header, and io.Reader then makes HTTP PUT request.
// Basically, it will call HystrixHTTPClient.Do(req *http.Request) method.
func (h *HystrixHTTPClient) Put(url string, header http.Header, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		return resp, errors.Wrap(err, "Couldn't create HTTP Request for Put method in HystrixHTTPClient")
	}

	if header == nil {
		header = make(http.Header)
	}
	req.Header = header
	return h.Do(req)
}

// Patch receives an URL, http.Header, and io.Reader then makes HTTP PATCH request.
// Basically, it will call HystrixHTTPClient.Do(req *http.Request) method.
func (h *HystrixHTTPClient) Patch(url string, header http.Header, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodPatch, url, body)
	if err != nil {
		return resp, errors.Wrap(err, "Couldn't create HTTP Request for Patch method in HystrixHTTPClient")
	}

	if header == nil {
		header = make(http.Header)
	}
	req.Header = header
	return h.Do(req)
}

// Delete receives an URL and http.Header then makes HTTP DELETE request.
// Basically, it will call HystrixHTTPClient.Do(req *http.Request) method.
func (h *HystrixHTTPClient) Delete(url string, header http.Header) (resp *http.Response, err error) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return resp, errors.Wrap(err, "Couldn't create HTTP Request for Delete method in HystrixHTTPClient")
	}

	if header == nil {
		header = make(http.Header)
	}
	req.Header = header
	return h.Do(req)
}

func configureHystrix(config CircuitBreakerConfig, timeout int) {
	if config.ErrorPercentThreshold <= 0 {
		config.ErrorPercentThreshold = DefaultErrorPercentThreshold
	}
	if config.RequestVolumeThreshold <= 0 {
		config.RequestVolumeThreshold = DefaultRequestVolumeThreshold
	}
	if config.SleepWindow <= 0 {
		config.SleepWindow = DefaultSleepWindow
	}
	if config.MaxConcurrentRequest <= 0 {
		config.MaxConcurrentRequest = DefaultMaxConcurrentRequest
	}

	hystrix.ConfigureCommand(config.Name, hystrix.CommandConfig{
		Timeout:                timeout,
		ErrorPercentThreshold:  config.ErrorPercentThreshold,
		RequestVolumeThreshold: config.RequestVolumeThreshold,
		SleepWindow:            config.SleepWindow,
		MaxConcurrentRequests:  config.MaxConcurrentRequest,
	})
}
