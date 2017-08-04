package main

import (
	"bufio"
	"fmt"
	"os"
	"./util"

	"github.com/howeyc/gopass"
)

func main() {
	p := util.NewPixiv()

	stdin := bufio.NewReader(os.Stdin)

	// 如果未登录，则提示输入登录信息
	if !p.IsLogged() {
		fmt.Print("请输入用户名：")
		var user string
		fmt.Fscan(stdin, &user)
		stdin.ReadString('\n')

		fmt.Print("请输入密码：")
		pass, _ := gopass.GetPasswd()

		err := p.Login(user, string(pass))
		if err != nil {
			fmt.Println("错误：", err)
		} else {
			fmt.Println("登录成功")
			// 保存登录状态
			p.SaveCookies()
		}
	}

	pagesOfShow := p.GetBookmarkTotalPages("show")
	pagesOfHide := p.GetBookmarkTotalPages("hide")
	fmt.Println("您的公开收藏夹一共有", pagesOfShow, "页")
	fmt.Println("您的非公开收藏夹一共有", pagesOfHide, "页")

	p.BatchReadIllusts(1, pagesOfShow, "show")

	go func() {
		for {
			fmt.Println(<-p.IllustsMeta)
		}
	}()

	for {
		id := <-p.Illusts
		fmt.Println(id)
		go p.GetIllustMetaData(id)
	}
}
