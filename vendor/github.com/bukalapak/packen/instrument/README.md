## Instrument

Instrument package is used to write metric to Prometheus. There are three mechanisms in this package.

### 1. Provider

  This mechanism provides an HTTP handler that can be used by Prometheus to scrap metrics. Developers **must** provide a special endpoint to be scrapped by Prometheus. Based on SRE standards, the special endpoint must be `/metrics`. The example below shows us how to provide `/metrics` endpoint.

  ```golang
  import (
    "github.com/bukalapak/packen/instrument"
    "github.com/julienschmidt/httprouter"
  )

  func main() {
    router := httprouter.New()
    router.HandlerFunc("GET", "/metrics", instrument.Handler)
  }
  ```

  Please, note that this mechanism only provide an endpoint that can be used by Prometheus to scrap metrics. Developers still must write their own metrics. The second mechanism will help developers.


### 2. Writer

  There are at least three types of metric: histogram, counter, and gauge.

  #### Histogram

  Developer can write histogram-typed metric using instrument. Package instrument provide an exported function called `ObserveLatency`. The metric's name is `service_latency_seconds`. It has four parameters.

  | Number | Name | Description |
  |---|---|---|
  | 1 | `method` | Indicates which method the metric belongs to, such as HTTP GET, HTTP POST, or just an arbitrary custom method |
  | 2 | `action` | Indicates which action the metric belongs to. It can be a function name or an HTTP Path |
  | 3 | `status` | Indicates the result of the process. It should only have one of the following: "ok" or "fail" |
  | 4 | `latency` | Indicates how long the process was run |

  
  Look at this divisor function.

  ```golang
  func div(a, b float64) (float64, error) {
    if b == 0 {
      return 0, errors.New("can't divide by zero")
    }
    return a / b, nil
  }
  ```

  Suppose developers want to know how long the function run, developers can use histogram-typed metric.

  ```golang
  import "github.com/bukalapak/packen/instrument"

  func main() {
    startTime := time.Now()
    res, err := div(2147483647, 3)
    elapsedTime := time.Since(startTime).Seconds()

    if err != nil {
      instrument.ObserveLatency("COMMAND", "divisor", "fail", elapsedTime)
    } else {
      instrument.ObserveLatency("COMMAND", "divisor", "ok", elapsedTime)
    }
  }
  ```

  #### Counter

  As the name implies, counter-typed metric usually be used to count something, such as number of request. Package instrument provides an exported function called `IncrementByOne`. The metric's name is `service_entity_counter`. It has three parameters.

  | Number | Name | Description |
  |---|---|---|
  | 1 | `service` | Indicates the service name |
  | 2 | `entity` | Indicates entity of the service |
  | 3 | `status` | Indicates the result of the process. It should only have one of the following: "ok" or "fail" |

  Let's take a look at the example below. Please, remember the divisor function above.

  ```golang
  import "github.com/bukalapak/packen/instrument"

  func main() {
    res, err := div(12345, 6)
    if err != nil {
      instrument.IncrementByOne("calculator", "divisor", "fail")
    } else {
      instrument.IncrementByOne("calculator", "divisor", "ok")
    }
  }
  ```

  #### Gauge

  Gauge is a metric that represents a single numerical value that arbitrarily go up and down. The only difference between Counter and Gauge is that Counter always goes up. Package instrument provides an exported function called `Gauge`. It has four parameters.

  | Number | Name | Description |
  |---|---|---|
  | 1 | `service` | Indicates the service name |
  | 2 | `entity` | Indicates entity of the service |
  | 3 | `status` | Indicates the result of the process. It should only have one of the following: "ok" or "fail" |
  | 4 | `value` | Indicates the current value of the process |


  Let's take a look at the example below.

  ```golang
  import (
    "math/rand" 

    "github.com/bukalapak/packen/instrument"
  )

  func randomNumber() int {
    return rand.Intn(100)
  }

  func main() {
    number := randomNumber()
    instrument.Gauge("cost-adder", "random", "ok", number)
  }
  ```

### 3. Runner

Sometimes, a function needs to be observed. This mechanism provides that functionality.

To use this mechanism, developers must create a task (type of instrument.Task). Developers must define the task's name and task's job. The task's job itself is a function that receives nothing and return an error.

Let's see the example below.

```golang
package main

import (
	"errors"
	"fmt"

	"github.com/bukalapak/packen/instrument"
)

func main() {
	var result float64
	var err error

	// create a task to be observed
	task := instrument.Task{
		// give the task a name
		Name: "divisor function",
		// wrap divisor function so it
		// satifies a function that receives nothing and return an error
		Job: func() error {
			result, err = divisor(123, 3)
			return err
		},
	}

	// run task inside circuit breaker
	err = instrument.ObserveTask(task)
	fmt.Println(result)
	fmt.Println(err)
}

func divisor(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("can't divide by zero")
	}
	return a / b, nil
}

```

The instrument will write the metrics to Prometheus using `ObserveLatency` method. Since `ObserveTask` uses `ObserveLatency` to write the metrics, the contract to see the metric is the same as what has been describe above in [histogram section](https://github.com/bukalapak/packen/tree/master/instrument#histogram). For clarity, the method is **packen-instrument-observe-task** and the action is **task's name**.