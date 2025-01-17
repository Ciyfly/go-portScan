package tcp

import (
	"context"
	"errors"
	"fmt"
	"github.com/XinRoom/go-portScan/core/port"
	"github.com/XinRoom/go-portScan/core/port/fingerprint"
	limiter "golang.org/x/time/rate"
	"net"
	"time"
)

var DefaultTcpOption = port.Option{
	Rate:    1000,
	Timeout: 800,
}

type TcpScanner struct {
	ports   []uint16             // 指定端口
	retChan chan port.OpenIpPort // 返回值队列
	limiter *limiter.Limiter
	ctx     context.Context
	timeout time.Duration
	isDone  bool
	option  port.Option
}

// NewTcpScanner Tcp扫描器
func NewTcpScanner(retChan chan port.OpenIpPort, option port.Option) (ts *TcpScanner, err error) {
	// option verify
	if option.Rate <= 0 {
		err = errors.New("rate can not set to 0")
		return
	}
	if option.Timeout <= 0 {
		err = errors.New("timeout can not set to 0")
		return
	}

	ts = &TcpScanner{
		retChan: retChan,
		limiter: limiter.NewLimiter(limiter.Every(time.Second/time.Duration(option.Rate)), 10),
		ctx:     context.Background(),
		timeout: time.Duration(option.Timeout) * time.Millisecond,
		option:  option,
	}

	return
}

// Scan 对指定IP和dis port进行扫描
func (ts *TcpScanner) Scan(ip net.IP, dst uint16) error {
	if ts.isDone {
		return errors.New("scanner is closed")
	}
	openIpPort := port.OpenIpPort{
		Ip:   ip,
		Port: dst,
	}
	var isTimeout bool
	if ts.option.FingerPrint {
		openIpPort.Service, isTimeout = fingerprint.PortIdentify("tcp", ip, dst, ts.timeout)
		if isTimeout {
			return nil
		}
	}
	if ts.option.Httpx && (openIpPort.Service == "" || openIpPort.Service == "http" || openIpPort.Service == "https") {
		openIpPort.HttpInfo, isTimeout = fingerprint.ProbeHttpInfo(ip, dst, ts.timeout)
		if isTimeout {
			return nil
		}
	}
	if !ts.option.FingerPrint && !ts.option.Httpx {
		conn, _ := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, dst), ts.timeout)
		if conn != nil {
			conn.Close()
		} else {
			return nil
		}
	}
	ts.retChan <- openIpPort
	return nil
}

func (ts *TcpScanner) Wait() {
}

// Close chan
func (ts *TcpScanner) Close() {
	ts.isDone = true
	ts.retChan <- port.OpenIpPort{}
}

// WaitLimiter Waiting for the speed limit
func (ts *TcpScanner) WaitLimiter() error {
	return ts.limiter.Wait(ts.ctx)
}
