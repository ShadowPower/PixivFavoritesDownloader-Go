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
// 可以自动记录Cookie
type WebClient struct {
	Client        *http.Client
	Cookies       *cookiejar.Jar
	commonHeaders map[string]string // don't write
}

// NewWebClient 创建一个 WebClient 对象
func NewWebClient() WebClient {
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
	return wc
}

// InitHeaders 初始化公共 Header
func (wc *WebClient) InitHeaders() {
	wc.commonHeaders = make(map[string]string)
	wc.commonHeaders["User-Agent"] = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:54.0) Gecko/20100101 Firefox/54.0"
	wc.commonHeaders["Accept-Language"] = "zh-CN,zh;q=0.5"
}

// Get 向指定 URL 发送GET请求，返回响应的主体和状态码
// url: 指定的 URL， retry: 重试次数
func (wc *WebClient) Get(url string, headers map[string]string, retry int) ([]byte, int, error) {
	req, _ := http.NewRequest("GET", url, nil)
	for k, v := range wc.commonHeaders {
		req.Header.Set(k, v)
	}
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
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

func (wc *WebClient) PostString(url string, headers map[string]string, body string) ([]byte, error) {
	data := bytes.NewBufferString(body)
	req, _ := http.NewRequest("POST", url, data)
	for k, v := range wc.commonHeaders {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Length", strconv.Itoa(data.Len()))
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
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
