[![Coverage Status](http://opencov.bukalapak.io/projects/5/badge.svg)](http://opencov.bukalapak.io/projects/5)

# Espeon
An enhanced HTTP client for Go which implements Bukalapak SRE standards.

## Description

Espeon provides enhanced HTTP client that implements SRE standards. Espeon implements:

  - [Retry and Overload Strategy](https://bukalapak.atlassian.net/wiki/spaces/INF/pages/101653660/RFC+Overload+and+Error+Handling)
  - [Request Context (Only X-Request-ID and X-Retry)](https://bukalapak.atlassian.net/wiki/spaces/INF/pages/119275619/RFC+Request+Context)
  - [Circuit Breaker Strategy](https://bukalapak.atlassian.net/wiki/spaces/INF/pages/106201155/RFC+Circuit+Breaker)

Espeon provides an interface for HTTP. Developers can see it in [Client interface](https://github.com/bukalapak/espeon/blob/master/http_client.go).

Espeon provides two actual implementations of HTTP client.

  - HTTPClient

    It works like and uses native [net/http](https://golang.org/pkg/net/http/) package but, in addition, has retry strategy.

  - HTTPClient with Circuit Breaker

    It works like and uses native [net/http](https://golang.org/pkg/net/http/) package but, in addition, has retry and circuit breaker strategy. Developers are strongly recommended to use this implementation over the first one.

## Owner

SRE - Library & Service

## Contact and On-Call Information

See [Contact and On-Call Information](https://bukalapak.atlassian.net/wiki/spaces/INF/pages/100923159/Contact+and+On-Call+Information)

## Installation

```sh
go get -u github.com/bukalapak/espeon
```

## Usage

### Recommended Usage

Note: Please, refer to code documentation to know more about API usage, constant values, and code implementation. 

Developers are strongly recommended to use `HystrixHTTPClient` for their HTTP client. Here is an example of how to use `HystrixHTTPClient`:

```golang
// create a new HTTP client to connect to service x
client := NewDefaultHystrixHTTPClient("http-client-for-service-x")

// suppose we want to send GET request
resp, err := client.Get("http://localhost:1616", nil)
if err != nil {
  return err
}
defer resp.Body.Close()

body, err := ioutil.ReadAll(resp.Body)
fmt.Println(string(body))
```

Developers can also add request header:

```golang
// create a new HTTP client to connect to service x
client := NewDefaultHystrixHTTPClient("http-client-for-service-x")

// make http header 
header := make(http.Header)
header.Add("Content-Type", "application/json")

// suppose we want to send GET request
resp, err := client.Get("http://localhost:1616", header)
if err != nil {
  return err
}
defer resp.Body.Close()

body, err := ioutil.ReadAll(resp.Body)
fmt.Println(string(body))
```

Developers can also send request using `http.Do`

```golang
// create a new HTTP client to connect to service x
client := NewDefaultHystrixHTTPClient("http-client-for-service-x")

// create http request
req, err := http.NewRequest("GET", "http://localhost:1616", nil)
if err != nil {
  return err
}

// send request using http.Do API
resp, err := client.Do(req)
if err != nil {
  return err
}
defer resp.Body.Close()

body, err := ioutil.ReadAll(resp.Body)
fmt.Println(string(body))
```

Espeon also provides API for `Post`, `Put`, `Patch`, and `Delete`. Please, refer to API documentation for further reference.

### Custom Usage

For developers' convenience, Espeon provides custom usage to HTTP client.

Developers can use `NewCustomHystrixHTTPClient` method to provide any configs that developers want. Any unset configuration will automatically be set using default value.

```golang
// create circuit breaker config
config := CircuitBreakerConfig{
  // the name of circuit breaker
  // please, use a client only for a specific connection
  // please, read more about what other fields mean in documentation
  Name: "circuit-breaker-for-connection-to-service-x",
  ErrorPercentThreshold: 20,
  RequestVolumeThreshold: 100,
  SleepWindow: 10000,
  MaxConcurrentRequest: 5000,
  FallbackFunc: func(err error) error {
    fmt.Println("custom fallback func")
    return err
  },
}
backoff := NewConstantBackoff()
maxRetry := 5
timeout := 2 * time.Second

client := NewCustomHystrixHTTPClient(timeout, backoff, maxRetry, config)
// omitted
```

### Another HTTP Client

If, by any chance, developers decide to ignore the circuit breaker standard but still want to follow retry and overload standard, Espeon is kind enough to provide them. Developers can use `HTTPClient`. It follows retry and overload standard but doesn't implement circuit breaker standard. Please, note that this approach is not recommended by SREs. If developers use this implementation, please keep in mind about the side effects: SREs, even the On-Call ones, will not support the services.

```golang
// create a new http client
client := NewDefaultHTTPClient()

// suppose we want to send GET request 
resp, err := client.Get("http://localhost:1616", nil)
if err != nil {
  return err
}
defer resp.Body.Close()

body, err := ioutil.ReadAll(resp.Body)
fmt.Println(string(body))
```

Developers can also use custom values:

```golang
backoff := NewConstantBackoff()
maxRetry := 5
timeout := 2 * time.Second

// create a new http client
client := NewCustomHTTPClient(timeout, backoff, maxRetry)

// suppose we want to send GET request 
resp, err := client.Get("http://localhost:1616", nil)
if err != nil {
  return err
}
defer resp.Body.Close()

body, err := ioutil.ReadAll(resp.Body)
fmt.Println(string(body))
```

### Backoff

Backoff is a mechanism in retry strategy. It defines an interval between two subsequent requests. Espeon provides three actual backoff implementations, `NoBackoff`, `ConstantBackoff` and `ExponentialBackoff`. They implement `Backoff` interface. In `Backoff` interface, method `NextInterval` receives a parameter that indicates the order of the interval (N-th interval).

SREs choose `ExponentialBackoff` as the default value for HTTP client implementation.

```golang
// create a constant backoff
backoff := NewConstantBackoff()
// get the 3rd interval
interval := backoff.NextInterval(3)

// create an exponential backoff
backoff = NewExponentialBackoff()
// get the 2nd interval
interval = backoff.NextInterval(2)

// create a no backoff
backoff = NewNoBackoff()
// get the 4th interval
interval = backoff.NextInterval(4)
```

### Metric

Espeon provides metric exporter for circuit breaker. It uses `prometheus counter`. It will record all circuit breaker usages. Developers may import this metric then show it to their Grafana dashboard. Developers are free to customize the code. Here is one of the examples of how to use metric and let Prometheus reads it.

```golang
// suppose your project is <your-project> and has package metric.
package metric

import (
  "github.com/bukalapak/espeon"
  "github.com/prometheus/client_golang/prometheus"
  "github.com/prometheus/client_golang/prometheus/promhttp"
)

var counter *prometheus.CounterVec

func init() {
  counter = espeon.CounterVec()
  prometheus.MustRegister(counter)
}

func Handler(w http.ResponseWriter, r *http.Request) {
  promhttp.Handler().ServeHTTP(w, r)
}
``` 

Then, in router definition, add metric handler to your endpoints. Remember, developers are free to customize the code. Here is only one of the examples of how to use it and may not be the implementation best practice so far.

```go
import "github.com/bukalapak/<your-project>/metric"

// you can use any http router
// unrelevant codes will be omitted
router.Get("/metrics", metric.Handler)
// omitted
```

In Grafana, developers can show the metric using this command

```
sum(rate(entity_total{job=~"espeon-circuit-.*"}[1m])) by (action,status)
```

## Documentation

Please, refer to in-code documentation for more information.

## Contributing

- Read and learn Espeon

- Create a new branch with descriptive name about the changes and checkout to the new branch

  ```sh
  git checkout -b branch-name
  ```

- Make changes and don't forget to add and/or update the unit tests.

- Please, beautify the codes. Developers can use any 3rd party application to beautify the codes, such as `gometalinter` or an extension like `lukehoban.Go` in VS Code.

- Add/update the dependencies. Espeon uses `dep` as dependency manager.

  ```sh
  dep ensure
  ```

- Commit and push the changes to upstream repository

  ```sh
  git commit -m "a meaningful commit message"
  git push origin branch-name
  ```

- Open a Pull Request in upstream repository

- Ask SREs to review the changes

- If the Pull Request is accepted by reviewers, merge Pull Request using `squash and merge` option

## FAQ
