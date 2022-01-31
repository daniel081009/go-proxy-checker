package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type proxy struct {
	Ip     string
	Port   string
	ssl    bool
	status bool
}

var timeout = time.Duration(10 * time.Second)

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}
func check(e error) int {
	if e != nil {
		return 1
	}
	return 0
}

func readProxy(filename string) []proxy { // 파일에서 아이피랑 port 를 추출해서 struct 배열에 집어넣음
	var proxyList []proxy
	data, err := ioutil.ReadFile(filename)
	if check(err) != 0 { //예외처리
		return proxyList
	}
	split := strings.Split(string(data), "\n") //개행에 따라 나누기
	for _, data := range split {
		if len(strings.Split(data, ":")) != 2 { //잘못된 문자면 넘김
			continue
		}
		da := strings.Split(data, ":") // :를 기준으로 ip port 를 나눔
		proxyList = append(proxyList, proxy{da[0], da[1], false, false})
	}
	return proxyList
}

func checkProxy(proxy proxy, dan chan proxy) { //이름대로 프록시를 체크함
	u, err := url.Parse("https://" + proxy.Ip + ":" + proxy.Port)
	if err != nil {
		panic(err)
	}
	tr := &http.Transport{
		Proxy:        http.ProxyURL(u),
		Dial:         dialTimeout,
		TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
	}
	client := &http.Client{Transport: tr}
	_, err = client.Get("https://google.com")
	if err != nil { //첫번쨰로 https 요청을 함
		u, err := url.Parse("http://" + proxy.Ip + ":" + proxy.Port)
		if err != nil {
			panic(err)
		}
		tr := &http.Transport{
			Proxy:        http.ProxyURL(u),
			Dial:         dialTimeout,
			TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
		}
		client := &http.Client{Transport: tr}
		_, err = client.Get("https://google.com")
		if err == nil {
			proxy.status = true
		}
	} else { //https 가능
		proxy.ssl = true
		proxy.status = true
	}
	dan <- proxy
}

func saveProxy(list []proxy) {
	string := ""
	for _, data := range list {
		string += data.Ip + ":" + data.Port + "\n"
	}
	err := ioutil.WriteFile("./proxy.txt", []byte(string), 0644)
	check(err)
}
func main() {
	list := readProxy("./test")
	dan := make(chan proxy, 1)
	fmt.Println("start")
	for _, dataProxy := range list {
		go checkProxy(dataProxy, dan)
	}
	for i := 0; i <= len(list); i++ {
		dans := <-dan
		if dans.status != true {
			list = list[i:]
			continue
		}
	}
	saveProxy(list)
}
