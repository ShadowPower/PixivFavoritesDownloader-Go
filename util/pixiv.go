package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Pixiv 是一个中文版 Pixiv 网站的封装库
type Pixiv struct {
	wc      *WebClient
	Illusts chan string

	illustRe *regexp.Regexp // 预编译的从收藏夹中获取作品ID的正则表达式
}

// NewPixiv 用于创建一个 Pixiv 类的对象
func NewPixiv() Pixiv {
	wc := NewWebClient(10)
	pixiv := Pixiv{wc: &wc}
	pixiv.Illusts = make(chan string, 500)

	// 编译正则表达式状态机
	pixiv.illustRe, _ = regexp.Compile("data-click-action=\"illust\"data-click-label=\"(\\d+)\"")
	return pixiv
}

// getPostKey 用于获取登录所需的 Post Key
func (p *Pixiv) getPostKey() (string, error) {
	url := "https://accounts.pixiv.net/login"
	re, _ := regexp.Compile("name=\"post_key\" value=\"([a-f0-9]{32})\"")

	p.wc.InitHeaders()
	p.wc.headers["Host"] = "accounts.pixiv.net"
	p.wc.headers["Referer"] = "https://www.pixiv.net/"
	p.wc.headers["Accept"] = "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"
	body, _, err := p.wc.Get(url, 5)
	if err != nil {
		return "", errors.New("获取 Post Key 失败")
	}
	postKey := re.FindSubmatch(body)
	return string(postKey[1]), nil
}

// Login 用于登录您的 Pixiv 账号
func (p *Pixiv) Login(username, password string) error {
	url := "https://accounts.pixiv.net/api/login?lang=zh"

	// getPostKey 会清除 Headers 设置，所以先获取
	postKey, err := p.getPostKey()
	if err != nil {
		return err
	}

	// 设置 Headers
	p.wc.InitHeaders()
	p.wc.headers["Host"] = "accounts.pixiv.net"
	p.wc.headers["Referer"] = "https://accounts.pixiv.net/login?lang=zh&source=pc&view_type=page&ref=wwwtop_accounts_index"
	p.wc.headers["Accept"] = "application/json, text/javascript, */*; q=0.01"
	p.wc.headers["X-Requested-With"] = "XMLHttpRequest"
	p.wc.headers["Content-Type"] = "application/x-www-form-urlencoded; charset=UTF-8"

	loginPara := "pixiv_id=" + username + "&password=" + password + "&captcha=&g_recaptcha_response=&post_key=" +
		postKey + "&source=pc&ref=wwwtop_accounts_index&return_to=https%3A%2F%2Fwww.pixiv.net%2F"
	body, err := p.wc.PostString(url, loginPara)
	if err != nil {
		return err
	}

	var r map[string]interface{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		return errors.New("JSON解析失败")
	}

	if r["error"] == true {
		// 如果请求出错
		fmt.Println(r["message"])
		return errors.New(r["message"].(string))
	} else if msg, ok := r["body"].(map[string]interface{})["validation_errors"]; ok {
		// 如果有登录错误信息
		errorMessage := ""
		for _, v := range msg.(map[string]interface{}) {
			errorMessage += v.(string) + "\n"
		}
		return errors.New(errorMessage)
	} else if _, ok := r["body"].(map[string]interface{})["success"]; ok {
		// 如果登录成功
		return nil
	}
	return nil
}

// IsBookmarkPageExist 用于查询书签页码是否存在
// rest: show=公开收藏夹/hide=非公开收藏夹
func (p *Pixiv) IsBookmarkPageExist(pageNumber int, rest string) bool {
	url := "https://www.pixiv.net/bookmark.php?rest=" + rest + "&p=" + strconv.Itoa(pageNumber)
	body, _, err := p.wc.Get(url, 5)
	if err != nil {
		return false
	}
	return !strings.Contains(string(body), "li class=\"_no-item\"")
}

// GetBookmarkTotalPages 使用二分查找来确认书签的最大页码
func (p *Pixiv) GetBookmarkTotalPages(rest string) int {
	min, max, temp := 0, 1, 0
	// 翻倍获取页码区间
	for p.IsBookmarkPageExist(max, rest) {
		max *= 3
	}
	min = max / 32
	// 二分查找
	for max-min > 1 {
		temp = (max + min) / 2
		if p.IsBookmarkPageExist(temp, rest) {
			min = temp // min 是存在的
		} else {
			max = temp // max 是不存在的
		}
	}
	return min
}

// ReadIllusts 读取一页收藏夹的作品列表，将作品ID写入到 Illusts 中
func (p *Pixiv) ReadIllusts(pageNumber int, rest string) {
	url := "https://www.pixiv.net/bookmark.php?rest=" + rest + "&p=" + strconv.Itoa(pageNumber)
	body, _, _ := p.wc.Get(url, 5)
	results := p.illustRe.FindAllSubmatch(body, -1)
	for _, result := range results {
		p.Illusts <- string(result[1])
	}
}

// BatchReadIllusts 启动一个异步批量读取收藏夹作品列表任务
func (p *Pixiv) BatchReadIllusts(from, to int, rest string) {
	for i := from; i <= to; i++ {
		go p.ReadIllusts(i, rest)
	}
}

// SaveCookies 保存 Cookie 到文件，以便下次使用
func (p *Pixiv) SaveCookies() {
	p.wc.Cookies.Save()
}

// IsLogged 检测登录状态，已登录返回 true，未登录或失败返回 false
func (p *Pixiv) IsLogged() bool {
	p.wc.InitHeaders()
	p.wc.headers["Host"] = "www.pixiv.net"
	p.wc.headers["Accept"] = "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"
	body, _, err := p.wc.Get("https://www.pixiv.net", 2)
	if err != nil {
		return false
	}
	return strings.Contains(string(body), "class=\"item header-logout\"")
}
