package trace

type Config struct {
	//	CollectorURL is the URL of the collector to which the exporter will send spans.
	CollectorURL string
	ServiceName  string
}
