package mrgraphite

import (
	"net"
	"sync"
	"time"
)

var (
	defaultClient       *Client
	quantileListPreinit []*Quantile
)

const (
	messagesBuffer = 1024
)

type Logger interface {
	Warningf(format string, args ...interface{})
}

type metric struct {
	name  string
	value int64
	ts    int64
}
type Client struct {
	conn                     net.Conn
	network, address, prefix string
	aggrtime                 time.Duration
	done                     chan struct{}
	messages                 chan metric
	aggrMsg                  map[string]int64
	mtx                      sync.Mutex
	writeBuf                 []byte
	quantileList             []*Quantile
	logger                   Logger
}

func InitDefaultClient(network, address, prefix string, aggrtime time.Duration, log Logger) *Client {
	if defaultClient != nil {
		defaultClient.Stop()
	}
	defaultClient = NewClient(network, address, prefix, aggrtime, log)
	defaultClient.quantileList = quantileListPreinit
	return defaultClient
}
func GetDefaultClient() *Client {
	if defaultClient == nil {
		panic("GetDefaultClient without InitDefaultClient()")
	}
	return defaultClient
}

func NewClient(network, address, prefix string, aggrtime time.Duration, log Logger) *Client {
	if prefix != "" {
		if prefix[len(prefix)-1] != '.' {
			prefix += "."
		}
	}
	c := &Client{
		network:      network,
		address:      address,
		prefix:       prefix,
		aggrtime:     aggrtime,
		done:         make(chan struct{}),
		messages:     make(chan metric, messagesBuffer),
		aggrMsg:      make(map[string]int64, 128),
		writeBuf:     make([]byte, 0, bgWriteMax),
		quantileList: make([]*Quantile, 0, 4),
		logger:       log,
	}
	c.InitConn()
	go c.sendWorker()
	return c
}

func (c *Client) InitConn() bool {
	var err error
	c.conn, err = net.Dial(c.network, c.address)
	if err != nil {
		c.logger.Warningf("mrgraphite: Connecting to %s %s failed: %v", c.network, c.address, err)
		c.conn = nil
		return false
	}
	return true
}

// Client methods
func (c *Client) Stop() {
	if c.done != nil {
		c.done <- struct{}{}
	}
}

// SendSum. Can be sum'ed and will be aggregated with aggrtime
func (c *Client) SendSum(name string, value int64) {
	if value == 0 {
		return
	}
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if v, ok := c.aggrMsg[name]; ok {
		value += v
	}
	c.aggrMsg[name] = value
}
func (c *Client) Inc(name string) {
	c.SendSum(name, 1)
}

// SendRaw can not be sum'ed or avg'ed
func (c *Client) SendRaw(name string, value int64) {
	c.addRaw(name, value, time.Now().Unix())
}

type Timer struct {
	c          *Client
	name       string
	start      time.Time
	sendZero   bool
	sendRaw    bool
	sendSumCnt bool
	divider uint
	quantile   *Quantile
}

func (c *Client) GetTimer(name string) *Timer {
	return &Timer{
		c:        c,
		name:     name,
		start:    time.Now(),
		quantile: nil,
		divider: 1000000, //Milliseconds
	}
}
func (t *Timer) SetDivider(d uint) *Timer {
	if d == 0 {
		panic("SetDivider: cannot accept 0")
	}
	t.divider = d
	return t
}
func (t *Timer) SendZero() *Timer {
	t.sendZero = true
	return t
}
func (t *Timer) SendRaw() *Timer {
	t.sendRaw = true
	return t
}
func (t *Timer) SendSumCnt() *Timer {
	t.sendSumCnt = true
	return t
}

func (t *Timer) Stop() {
	if t.c == nil {
		return
	}
	tim := time.Since(t.start).Nanoseconds() / int64(t.divider)
	if t.sendRaw {
		if tim > 0 || t.sendZero {
			t.c.SendRaw(t.name, tim)
		}
	}
	if t.sendSumCnt {
		t.c.SendSum(t.name+"_sum", tim)
		t.c.Inc(t.name + "_cnt")
	}
	if t.quantile != nil {
		t.quantile.AddValue(tim)
	}
}

//Default client methods
func SendSum(name string, value int64) {
	if defaultClient != nil && defaultClient.done != nil {
		defaultClient.SendSum(name, value)
	}
}
func Inc(name string) {
	if defaultClient != nil && defaultClient.done != nil {
		defaultClient.Inc(name)
	}
}
func SendRaw(name string, value int64) {
	if defaultClient != nil && defaultClient.done != nil {
		defaultClient.SendRaw(name, value)
	}
}
func GetTimer(name string) *Timer {
	if defaultClient != nil && defaultClient.done != nil {
		return defaultClient.GetTimer(name)
	} else {
		return &Timer{}
	}
}
