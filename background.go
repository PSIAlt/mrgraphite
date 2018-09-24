package mrgraphite

import (
	"fmt"
	"time"
)

const (
	bgWriteMax   = 1400
	bgWriteFlush = 1200
)

func (c *Client) sendWorker() {
	aggr := time.NewTicker(c.aggrtime)
	defer aggr.Stop()

	writeAggrWait := make(chan struct{}, 1)
	writeAggrWait <- struct{}{} //We can run writeAggr()

SWFOR:
	for {
		eatMsg := func(m metric) {
			msg := fmt.Sprintf("%s%s %d %d\n", c.prefix, m.name, m.value, m.ts)
			if len(msg)+len(c.writeBuf) >= bgWriteMax {
				c.flushWrite()
			}
			if len(msg)+len(c.writeBuf) < bgWriteMax {
				c.writeBuf = append(c.writeBuf, msg...)
			}
		}

		// Wait something
		select {
		case m := <-c.messages:
			// Try to use writeBuf entirely to reduce syscalls
			eatMsg(m)
		EATLOOP:
			for i := 0; i < 100; i++ {
				select {
				case m := <-c.messages:
					eatMsg(m)
				default:
					break EATLOOP
				}
			}
			c.flushWrite()

		case <-aggr.C:
			// Do not call writeAggr directly - this can cause deadlock
			select {
			case <-writeAggrWait:
				go func() {
					c.writeAggr()
					writeAggrWait <- struct{}{} //We can run writeAggr() again
				}()
			default:
				c.logger.Warningf("mrgraphite: writeAggr skipped")
			}

		case <-c.done:
			break SWFOR // Closed this client
		}
	}

	// To prevent goroutines hang, we must clear c.messages
SWQUIT:
	for {
		select {
		case <-c.messages:
		default:
			break SWQUIT
		}
	}
	c.conn.Close()
	c.conn = nil
	c.done = nil
}

func (c *Client) flushWrite() {
	//c.logger.Warningf("mrgraphite: flushWrite")
	if c.conn == nil {
		if c.InitConn() == false {
			return
		}
	}
	if len(c.writeBuf) == 0 {
		return
	}

	n, err := c.conn.Write(c.writeBuf)
	if err != nil {
		c.logger.Warningf("mrgraphite: Write to %s %s failed: %v", c.network, c.address, err)
		if c.InitConn() == false {
			return
		}
	}
	c.writeBuf = c.writeBuf[n:]
}

func (c *Client) writeAggr() {
	// Aggregate aggrMsg
	c.mtx.Lock()
	aggr := c.aggrMsg
	c.aggrMsg = make(map[string]int64, 128)
	c.mtx.Unlock()
	ts := time.Now().Unix()
	for k, v := range aggr {
		c.addRaw(k, v, ts)
	}

	for _, v := range c.quantileList {
		if val, err := v.GetValue(); err == nil {
			name := fmt.Sprintf("%s_q%d", v.GetName(), v.GetQVal())
			c.addRaw(name, val, ts)
		}
	}
}

func (c *Client) addRaw(name string, value int64, ts int64) {
	c.messages <- metric{name, value, ts}
}
