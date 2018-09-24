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
	qVal   int
	values valuesType
	mtx    sync.Mutex
}

func NewQuantile(name string, qVal int) *Quantile {
	return NewQuantileC(defaultClient, name, qVal)
}

func NewQuantileC(c *Client, name string, qVal int) *Quantile {
	if qVal < 0 || qVal > 100 {
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

func (q *Quantile) GetQVal() int {
	return q.qVal
}

func (q *Quantile) GetValue() (int64, error) {
	q.mtx.Lock()
	vc := q.values
	q.values = valPool.Get().(valuesType)
	q.mtx.Unlock()

	if len(vc) == 0 {
		valPool.Put(vc[:0])
		return 0, EmptyListError
	}

	sort.Slice(vc, func(i, j int) bool { return vc[i] < vc[j] })
	nElem := float64(len(vc)) / 100.0 * float64(q.qVal)
	retval := vc[int(nElem)]

	if cap(vc) > 100000 {
		vc = valPool.New().(valuesType)
	}
	valPool.Put(vc[:0])

	return retval, nil
}
