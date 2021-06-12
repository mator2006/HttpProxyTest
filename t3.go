package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	TcpTimeOut  = 5 * time.Second
	HttpTimeOut = 10 * time.Second
)

const (
	ProxyServerListFile = "p.txt"
	TestURL1            = "https://www.baidu.com"
	TestURL2            = "https://www.google.com"
	TestURL3            = "https://github.com"
)

type PL struct {
	Socket         string
	SocketF        string
	URL1TestResult bool
	URL2TestResult bool
	URL3TestResult bool
}

var DW bool = false

func main() {
	var Pler PL
	PSL := Pler.INIT()
	var NewPLS []PL
	var vg sync.WaitGroup
	for _, v := range PSL {
		vg.Add(1)
		go func(v PL) {
			if v.FirstTest() {
				NewPLS = append(NewPLS, v)
			}
			vg.Done()
		}(v)
	}
	vg.Wait()

	var FPLS []PL
	for _, v := range NewPLS {
		vg.Add(1)
		go func(v PL) {

			FPLS = append(FPLS, v.MainTest())
			vg.Done()
		}(v)
	}
	vg.Wait()

	for _, v := range FPLS {
		if !v.URL1TestResult && !v.URL2TestResult && !v.URL3TestResult {
			continue
		}
		fmt.Printf("%v\t%v:%v\t%v:%v\t%v:%v\n", v.SocketF, TestURL1, v.URL1TestResult, TestURL2, v.URL2TestResult, TestURL3, v.URL3TestResult)
	}

}

func (p *PL) INIT() []PL {
	var replers []PL
	f, err := ioutil.ReadFile(ProxyServerListFile)
	if err != nil {
		panic("Can not open " + ProxyServerListFile)
	}
	fs := strings.Split(string(f), "\n")
	for _, v := range fs {
		v = strings.ReplaceAll(v, "\n", "")
		v = strings.ReplaceAll(v, "\r", "")
		v = strings.ReplaceAll(v, " ", "")
		if v == "" {
			continue
		}
		var pler PL
		pler.Socket = v
		pler.SocketF = "socks5://" + v
		replers = append(replers, pler)
	}
	return replers
}

func (p PL) FirstTest() bool {
	var reb bool
	_, err := net.DialTimeout("tcp", p.Socket, TcpTimeOut)
	if err != nil {
		DebugPrint(err)
		return reb
	}
	return true
}
func (p PL) MainTest() PL {
	var repler PL
	Proxy, err := url.Parse(p.SocketF)
	if err != nil {
		DebugPrint(err)
	}
	TestURLlist := []string{TestURL1, TestURL2, TestURL3}
	for i, v := range TestURLlist {
		method := "GET"
		client := &http.Client{
			Timeout: HttpTimeOut,
			Transport: &http.Transport{
				Proxy: http.ProxyURL(Proxy),
			},
		}

		req, err := http.NewRequest(method, v, nil)
		if err != nil {
			DebugPrint(err)
			return repler
		}

		res, err := client.Do(req)
		if err != nil {
			DebugPrint(err)
			return repler
		}
		defer res.Body.Close()
		if res.StatusCode == 200 {
			// fmt.Printf("Conut:%d\tRequest URL:%s\tTitle:%s\tStatusCode:%v\tStatus:%v\tProxy:%s\t\n", i+1, v, ts, res.StatusCode, res.Status, p.SocketF)
			switch i {
			case 0:
				repler.URL1TestResult = true
			case 1:
				repler.URL2TestResult = true
			case 2:
				repler.URL3TestResult = true
			}
			repler.Socket = p.Socket
			repler.SocketF = p.SocketF
		}
	}
	return repler
}

func DebugPrint(pc interface{}) {
	t := time.Now().Format("2006/01/02 15:04:05")
	if DW {
		fmt.Println(t, pc)
	}
}
