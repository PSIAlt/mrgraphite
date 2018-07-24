package mrgraphite

import (
	"net"
	"time"
	"fmt"
	"testing"
)

func mockListen(network, address string) chan string {
    // Listen for incoming connections.
    l, err := net.Listen(network, address)
    if err != nil {
        panic("Error listening: "+err.Error())
    }

//    defer l.Close()
    chn := make(chan string, 100)
    go func() {
		conn, err := l.Accept()
		if err != nil {
		    panic("Error accepting: "+err.Error())
		}
		//fmt.Println("Accepted")

		for {
			buf := make([]byte, 1024)
			// Read the incoming connection into the buffer.
			//fmt.Println("Reading")
			conn.SetDeadline( time.Now().Add(5*time.Millisecond) )
			reqLen, err := conn.Read(buf)
			if err != nil {
				conn.Close()
				chn <- "EOF"
				return
			}
			chn <- string(buf[:reqLen])
		}//read-write
	}()
	return chn
}

type myLog struct {t *testing.T}

func (l myLog) Warningf(format string, args ...interface{}) {
	fmt.Printf(format + "\n", args...)
	//l.t.Logf(format, args...)
}

func check(t *testing.T, expr bool, msg string) {
	if !expr {
		t.Fatalf("Test expr failed: %s", msg)
	}
}
func TestInitDefaultClient(t *testing.T) {
	log := myLog{t}
	mAddr := "127.0.0.1:9993"
	chn := mockListen("tcp", mAddr)
	defClient := InitDefaultClient("tcp", mAddr, "pref.svc", time.Millisecond, log)
	check(t, defaultClient.network == "tcp", "check network")
	check(t, defaultClient.address == mAddr, "check address")
	check(t, defaultClient.prefix == "pref.svc.", "check prefix")
	check(t, defaultClient.aggrtime == time.Millisecond, "check aggrtime")

	tnow := time.Now().Unix()
	Inc("metric1")
	m1 := <-chn
	if m1 != fmt.Sprintf("pref.svc.metric1 1 %d\n", tnow) {
		t.Fatalf("Invalid metric recv'd: '%s'", m1)
	}
	SendSum("metric2", 123)
	m2 := <-chn
	if m2 != fmt.Sprintf("pref.svc.metric2 123 %d\n", tnow) {
		t.Fatalf("Invalid metric recv'd: '%s'", m2)
	}
	defClient.Stop()
	m3 := <-chn
	if m3 != "EOF" {
		t.Fatalf("Invalid metric recv'd: '%s'", m2)
	}
	defClient.Stop()
}
