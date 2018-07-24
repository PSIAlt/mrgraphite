package mrgraphite

import (
	"net"
	"time"
	"sync"
)

var (
	defaultClient *Client
)
const (
	messagesBuffer = 1024
)

type Logger interface {
    Warningf(format string, args ...interface{})
}

type metric struct {
	name string
	value int64
	ts int64
}
type Client struct {
	conn net.Conn
	network, address, prefix string
	aggrtime time.Duration
	done chan struct{}
	messages chan metric
	aggrMsg map[string]int64
	mtx sync.Mutex
	writeBuf []byte
	logger Logger
}

func InitDefaultClient(network, address, prefix string, aggrtime time.Duration, log Logger) *Client {
	if defaultClient != nil {
		defaultClient.Stop()
	}
	defaultClient = NewClient(network, address, prefix, aggrtime, log)
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
	c := &Client {
		network: network,
		address: address,
		prefix: prefix,
		aggrtime: aggrtime,
		done: make(chan struct{}),
		messages: make(chan metric, messagesBuffer),
		aggrMsg: make(map[string]int64, 128),
		writeBuf: make([]byte, 0, bgWriteMax),
		logger: log,
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
		c.conn=nil
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
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if v, ok := c.aggrMsg[name]; ok {
		value += v;
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
	c *Client
	name string
	start time.Time
}
func (c *Client) GetTimer(name string) *Timer {
	return &Timer{
		c: c,
		name: name,
		start: time.Now(),
	}
}
func (t *Timer) Stop() {
	tim := time.Since(t.start).Nanoseconds() / 1000000
	t.c.SendRaw(t.name, tim)
}


//Default client methods
func SendSum(name string, value int64) {
	defaultClient.SendSum(name, value)
}
func Inc(name string) {
	defaultClient.Inc(name)
}
func SendRaw(name string, value int64) {
	defaultClient.SendRaw(name, value)
}
func GetTimer(name string) *Timer {
	return defaultClient.GetTimer(name)
}
