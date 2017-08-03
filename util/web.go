package util

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/juju/persistent-cookiejar"
)

// WebClient 是一个用来模拟浏览器的类
// 可以自动记录Cookie，可以限制最大并发连接数
type WebClient struct {
	Client   *http.Client
	Cookies  *cookiejar.Jar
	headers  map[string]string
	taskFlag chan int // 记录任务个数
}

// NewWebClient 创建一个 WebClient 对象
// maxConnections: 最大并发连接数
func NewWebClient(maxConnections int) WebClient {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    10 * time.Second,
		DisableCompression: true,
	}
	wc := WebClient{Client: &http.Client{Transport: tr}}
	options := cookiejar.Options{PublicSuffixList: nil, Filename: "cookies"}
	wc.Cookies, _ = cookiejar.New(&options)
	wc.Client.Jar = wc.Cookies
	wc.InitHeaders()
	wc.taskFlag = make(chan int, maxConnections)
	return wc
}

// InitHeaders 清除 HTTP 客户端的 Header，并设置默认 UA 和语言
func (wc *WebClient) InitHeaders() {
	wc.headers = make(map[string]string)
	wc.headers["User-Agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:54.0) Gecko/20100101 Firefox/54.0"
	wc.headers["Accept-Language"] = "zh-CN,zh;q=0.5"
}

// Get 向指定 URL 发送GET请求，返回响应的主体和状态码
// url: 指定的 URL， retry: 重试次数
func (wc *WebClient) Get(url string, retry int) ([]byte, int, error) {
	// 限制连接数
	wc.taskFlag <- 0
	defer func() { <-wc.taskFlag }()

	req, _ := http.NewRequest("GET", url, nil)
	for k, v := range wc.headers {
		req.Header.Set(k, v)
	}
GET:
	resp, err := wc.Client.Do(req)
	if err != nil {
		if retry > 0 {
			retry--
			fmt.Println("请求 GET", url, "失败，正在重试……")
			goto GET
		}
		return nil, 0, errors.New("GET请求失败")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, errors.New("无法读取数据")
	}
	return body, resp.StatusCode, nil
}

func (wc *WebClient) PostString(url, body string) ([]byte, error) {
	// 限制连接数
	wc.taskFlag <- 0
	defer func() { <-wc.taskFlag }()

	data := bytes.NewBufferString(body)
	wc.headers["Content-Length"] = strconv.Itoa(data.Len())
	req, _ := http.NewRequest("POST", url, data)
	for k, v := range wc.headers {
		req.Header.Set(k, v)
	}
	resp, err := wc.Client.Do(req)
	if err != nil {
		return nil, errors.New("POST请求失败")
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("无法读取数据")
	}
	return respBody, nil
}
