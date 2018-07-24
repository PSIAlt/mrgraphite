MRGraphite
========

## Features

- Very simple and clean API
- Support tcp & udp
- Load optimized

## Usage example

```
import "github.com/PSIAlt/mrgraphite"

...
type myLog struct {}

func (l myLog) Warningf(format string, args ...interface{}) {
	fmt.Printf(format + "\n", args...)
}

...

log := myLog{}
aggr_time := time.Millisecond*50
defer mrgraphite.InitDefaultClient("udp", "graphite:2003", "prefix.myservice", aggr_time, log).Stop()

...

// Measure simple sum (will be sum'ed every 'aggr_time')
mrgraphite.SendSum("stat.bytes", bytes)

// Same, but value==1
mrgraphite.Inc("stat.requests")

// Do not aggegate values, can be used to analyze raw metrics in graphtie
mrgraphite.SendRaw("timing.request_time", time_elasped)

func measureTimer() {
	defer mrgraphite.GetTimer("timing.measureTimer").Stop()
	// blahblahblah
	// mrgraphite automatically writes number of milliseconds elasped to graphite metric
}
```
