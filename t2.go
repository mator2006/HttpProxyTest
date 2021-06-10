package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
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

const (
	debugmode = false
)

func main() {

	TestURLlist := []string{TestURL1, TestURL2, TestURL3}

	CanUseProxyList := FirstProcessServerlist(ProxyServerListFile)

	var ProxyList []string

	for _, p := range CanUseProxyList {

		var testok bool = true
		for i, u := range TestURLlist {
			b, s := Dotest(i, u, p)
			if !b {
				testok = false
			}
			if debugmode {
				fmt.Println(s)
			}
		}
		if testok {
			ProxyList = append(ProxyList, p)
		}
	}
	fmt.Println(ProxyList)
}

func Dotest(i int, url1 string, Proxystr string) (bool, string) {
	var reb bool
	var Outstr string = fmt.Sprintf("Test fail.\tTarget:[%s]\tProxyServer:[%s]", url1, Proxystr)
	Proxy, err := url.Parse(Proxystr)

	if err != nil {
		log.Fatal("parse url error: ", err)
	}
	if debugmode {
		fmt.Printf("\nTest %s Start.\n", Proxystr)
	}
	method := "GET"
	client := &http.Client{
		Timeout: HttpTimeOut,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(Proxy),
		},
	}
	req, err := http.NewRequest(method, url1, nil)

	if err != nil {
		//fmt.Println(err)
		return reb, Outstr
	}
	res, err := client.Do(req)
	if err != nil {
		//fmt.Println(err)
		return reb, Outstr
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		var ts string
		doc.Find("title").Each(func(i int, selection *goquery.Selection) {
			ts = fmt.Sprintf("%v", selection.Text())
		})

		Outstr = fmt.Sprintf("Conut:%d\tRequest URL:%s\tTitle:%s\tStatusCode:%v\tStatus:%v\tProxy:%s\t", i+1, url1, ts, res.StatusCode, res.Status, Proxystr)
		reb = true
	}
	return reb, Outstr
}

func TestTcp(netstr string) bool {
	var reb bool
	_, err := net.DialTimeout("tcp", netstr, TcpTimeOut)
	if err != nil {
		// fmt.Println(err)
		return reb
	}
	return true
	// fmt.Fprintf(conn, "\n\r\n\r")
	// status, err := bufio.NewReader(conn).ReadString('\n')
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(status)
}

func FirstProcessServerlist(ProxyServerListFile string) []string {
	f, err := ioutil.ReadFile(ProxyServerListFile)
	if err != nil {
		panic("Can not open " + ProxyServerListFile)
	}
	var Resl []string
	ProxyList := strings.Split(string(f), "\n")
	for _, v := range ProxyList {
		v = strings.ReplaceAll(v, " ", "")
		v = strings.ReplaceAll(v, "\n", "")
		v = strings.ReplaceAll(v, "\r", "")
		if v == "" {
			continue
		}
		t := time.Now().Format("2006/01/02 15:04:05")
		if !TestTcp(v) {
			if debugmode {
				fmt.Printf("%v\t[%s]\t\t\t live NO.\n", t, v)
			}
			continue
		}
		if debugmode {
			fmt.Printf("%v\t[%s]\t\t\t live YES.\n", t, v)
		}
		Resl = append(Resl, "socks5://"+v)
	}
	if debugmode {
		fmt.Printf("\nFirst test result:[%d/%d]live.\n", len(Resl), len(ProxyList))
	}
	return Resl
}
