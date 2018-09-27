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
aggr_time := 50*time.Millisecond
defer mrgraphite.InitDefaultClient("udp", "graphite:2003", "prefix.myservice", aggr_time, log).Stop()

...

// Measure simple sum (will be sum'ed every 'aggr_time')
mrgraphite.SendSum("stat.bytes", bytes)

// Same, but value==1
mrgraphite.Inc("stat.requests")

// Do not aggregate values, can be used to analyze raw metrics in graphite
mrgraphite.SendRaw("timing.request_time", time_elasped)

func measureTimer() {
	defer mrgraphite.GetTimer("timing.measureTimer").SendRaw().Stop()
	// blahblahblah
	// mrgraphite automatically writes number of milliseconds elasped to graphite metric
}

var quantileMeasureMore = mrgraphite.NewQuantile("timing.measureTimerMore", 95)
func measureTimerMore() {
	defer quantileMeasureMore.GetTimer().SendSumCnt().SetDivider(1000).Stop()
	// blahblahblah
	// mrgraphite automatically aggregated number of microseconds(divider=1000) elasped:
	// timing.measureTimerMore_sum - sum of miscroseconds
	// timing.measureTimerMore_cnt - number of metrics (you can get avg from this)
	// timing.measureTimerMore_q95 - 95-quantile
}

func measureTimerQuantiles() {
	q1 := mrgraphite.NewQuantile("timing.measureTimerQuantiles", 95)
	q2 := mrgraphite.NewQuantile("timing.measureTimerQuantiles", 99.9)
	defer q1.GetTimer().AddQuantile(q2).Stop()
	// blahblahblah
	// mrgraphite automatically aggregated number of microseconds elasped:
	// timing.measureTimerQuantiles_q95
	// timing.measureTimerQuantiles_q99_90
}
```
