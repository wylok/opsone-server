package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/asmcos/requests"
	"strings"
	"time"
)

func SendDingDingMsg(token string, Data map[string]interface{}) bool {
	defer func() {
		if r := recover(); r != nil {
			Log.Error(errors.New(fmt.Sprint(r)))
		}
	}()
	url := "https://oapi.dingtalk.com/robot/send?access_token="
	req := requests.Requests()
	// 钉钉机器人webhook
	D, err := json.Marshal(Data)
	if err == nil {
		req.Header.Set("Content-Type", "application/json")
		for _, t := range strings.Split(token, ",") {
			resp, err := req.PostJson(url+t, string(D))
			// 返回处理
			if err != nil || resp == nil || resp.R.StatusCode != 200 {
				Log.Error(err)
				return false
			}
			time.Sleep(5 * time.Second)
		}
	} else {
		Log.Error(err)
		return false
	}
	return true
}
