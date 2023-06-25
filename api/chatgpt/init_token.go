package chatgpt

import (
	"encoding/json"
	"github.com/linweiyuan/go-chatgpt-api/api"
	"log"
	"os"
	"sync"
)

type accessToken struct {
	tokens []AuthResult
	lock   sync.Mutex
}

var TokenManager accessToken

func InitToken() {
	//read accounts.json
	file, err := os.Open("accounts.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	var accountList []api.LoginInfo
	err = decoder.Decode(&accountList)
	if err != nil {
		log.Println(err)
		panic(err)
	}
	var tokens []AuthResult
	for _, v := range accountList {
		authResult, err := Login(&v)
		if err != nil {
			log.Printf("账号:%s 登录失败：\n", v.Username, err)
			continue
		}
		tokens = append(tokens, *authResult)
	}
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
