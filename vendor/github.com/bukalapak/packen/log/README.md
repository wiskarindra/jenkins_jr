## Log

Log package is used to write log to Kibana.

To use packen log, developers must import **github.com/bukalapak/packen/log**. There are two methods developers can use, `RequestInfo` and `RequestError`. All methods receive two parameters. The first is string and the second is list of log.Field. The second parameter must contain at least three mandatory fields: `request_id`, `tags`, and `duration`. **Otherwise, log won't be written**.

All methods return an error. It will be nil if all mandatory fields are sufficed. Otherwise, it will be an error.

Example:

```golang
import "github.com/bukalapak/packen/log"

func main() {
  err := log.RequestInfo("I promise I am going to give at least 20 highfives to the creator of this awesome repository",
		log.NewField("request_id", "KMZ-WA-8-AWAA"),
		log.NewField("duration", "124"),
		log.NewField("tags", `["service_name","method_called"]`),
  )
  
  err = log.RequestError("If I don't give at least 20 highfives, I am a loser and will be embarassed for the rest of my life",
		log.NewField("request_id", "KMZ-WA-8-AWAA"),
		log.NewField("duration", "619"),
		log.NewField("tags", `["service_name","method_called"]`),
	)

  // omitted
}
```