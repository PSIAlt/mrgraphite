package mrgraphite

import (
	"errors"
	"sort"
	"sync"
)

type valuesType []int64

var valPool = sync.Pool{
	New: func() interface{} {
		return make(valuesType, 0, 1024)
	},
}
var EmptyListError = errors.New("Empty list")

type Quantile struct {
	c      *Client
	name   string
	qVal   float64
	values valuesType
	mtx    sync.Mutex
}

func NewQuantile(name string, qVal float64) *Quantile {
	return NewQuantileC(defaultClient, name, qVal)
}

func NewQuantileC(c *Client, name string, qVal float64) *Quantile {
	if qVal < 0.0 || qVal > 100.0 {
		panic("Wrong qVal")
	}
	q := &Quantile{
		c:      c,
		name:   name,
		qVal:   qVal,
		values: valPool.Get().(valuesType),
	}
	if c != nil {
		c.quantileList = append(c.quantileList, q)
	} else {
		quantileListPreinit = append(quantileListPreinit, q)
	}
	return q
}

func (q *Quantile) GetTimer() *Timer {
	if q.c == nil {
		if defaultClient == nil {
			// Pass to glabal GetTimer which is empty object
			return GetTimer(q.name)
		}
		// Late default client initializaion
		q.c = defaultClient
	}
	t := q.c.GetTimer(q.name)
	t.quantile = q
	return t
}

func (q *Quantile) AddValue(v int64) {
	q.mtx.Lock()
	defer q.mtx.Unlock()
	q.values = append(q.values, v)
}

func (q *Quantile) GetName() string {
	return q.name
}

func (q *Quantile) GetQVal() float64 {
	return q.qVal
}

func (q *Quantile) GetValue() (retval int64, err error) {
	q.mtx.Lock()
	vc := q.values
	q.values = valPool.Get().(valuesType)
	q.mtx.Unlock()

	n := len(vc)
	if n == 0 {
		valPool.Put(vc[:0])
		return 0, EmptyListError
	}

	sort.Slice(vc, func(i, j int) bool { return vc[i] < vc[j] })
	nElem := float64(n) * q.qVal / 100.0
	ofs := int(nElem)

	switch {
		case ofs==0:
			retval = vc[0]
		case ofs>=n:
			retval = vc[n-1]
		default:
			retval = vc[ofs]
	}

	if cap(vc) > 100000 {
		vc = valPool.New().(valuesType)
	}
	valPool.Put(vc[:0])
	return
}
