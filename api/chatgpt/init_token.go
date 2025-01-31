package chatgpt

import (
	"bufio"
	"github.com/linweiyuan/go-chatgpt-api/api"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type accessToken struct {
	tokens []AuthResult
	lock   sync.Mutex
}

var TokenManager accessToken

func InitToken() {
	//read accounts.json
	//file, err := os.Open("accounts.json")
	//if err != nil {
	//	panic(err)
	//}
	//defer file.Close()
	//decoder := json.NewDecoder(file)
	//var accountList []api.LoginInfo
	//err = decoder.Decode(&accountList)
	file, err := os.Open("accounts.txt")
	if err != nil {
		log.Fatalf("未找到账号文件: %s", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println(err)
			panic(err)
		}
	}(file)

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var accountList []api.LoginInfo
	for scanner.Scan() {
		line := scanner.Text()
		arr := strings.Split(line, "----")
		accountList = append(accountList, api.LoginInfo{
			Username: arr[0],
			Password: arr[1],
		})
	}

	if err != nil {
		log.Println(err)
		panic(err)
	}
	var tokens []AuthResult
	failCount := 0
	cc, _ := strconv.Atoi(os.Getenv("LOGIN_FAILED_RETRY_COUNT"))
	for _, v := range accountList {
		authResult, err := Login(&v)
		if err == nil {
			tokens = append(tokens, *authResult)
			time.Sleep(30 * time.Second)
			log.Printf("账号:%s 登录失败：, 错误成功 \n\n", v.Username)
			continue
		}
		log.Printf("账号:%s 登录失败：, 错误信息：%v \n\n", v.Username, err)
		addFlag := false
		for retryCount := 1; retryCount <= cc; retryCount++ {
			authResult, err = Login(&v)
			if err == nil {
				tokens = append(tokens, *authResult)
				log.Printf("账号:%s 第%d次重试登录成功 \n\n", v.Username, retryCount)
				break
			}
			if !addFlag {
				failCount++
				addFlag = true
			}
			log.Printf("账号:%s 经过%d次重试登录依然失败：, 错误信息：%v \n\n", v.Username, retryCount, err)
		}

	}
	log.Printf("所有账号登录完成,成功数量:%d   失败数量:%d  \n\n", len(accountList)-failCount, failCount)
	if err != nil {
		log.Println(err)
		panic(err)
	}
	TokenManager = accessToken{
		tokens: tokens,
	}
}

func (a *accessToken) GetToken() AuthResult {
	a.lock.Lock()
	defer a.lock.Unlock()

	if len(a.tokens) == 0 {
		return AuthResult{}
	}

	token := a.tokens[0]
	a.tokens = append(a.tokens[1:], token)
	return token
}
