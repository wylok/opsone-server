package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kirinlabs/HttpRequest"
)

func RequestApiPost(ApiUrl, Token string, Data map[string]interface{}) (interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			Log.Error(errors.New(fmt.Sprint(r)))
		}
	}()
	var (
		JsonData = map[string]interface{}{}
	)
	req := HttpRequest.NewRequest()
	defer func() {
		if err != nil {
			err = errors.New("请求接口" + ApiUrl + "发生错误:" + err.Error())
		}
	}()
	req.SetHeaders(map[string]string{"Content-Type": "application/json", "token": Token})
	resp, err := req.JSON().Post(ApiUrl, Data)
	if err == nil && resp != nil {
		body, err := resp.Body()
		if err == nil {
			err = json.Unmarshal(body, &JsonData)
		}
		_ = resp.Close()
	}
	return JsonData, err
}
func RequestApiPut(ApiUrl, Token string, Data map[string]interface{}) (interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			Log.Error(errors.New(fmt.Sprint(r)))
		}
	}()
	var (
		JsonData = map[string]interface{}{}
	)
	req := HttpRequest.NewRequest()
	defer func() {
		if err != nil {
			err = errors.New("请求接口" + ApiUrl + "发生错误:" + err.Error())
		}
	}()
	req.SetHeaders(map[string]string{"Content-Type": "application/json", "token": Token})
	resp, err := req.JSON().Put(ApiUrl, Data)
	if err == nil && resp != nil {
		body, err := resp.Body()
		if err == nil {
			err = json.Unmarshal(body, &JsonData)
		}
		_ = resp.Close()
	}
	return JsonData, err
}
func RequestApiDelete(ApiUrl, Token string, Data map[string]interface{}) (interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			Log.Error(errors.New(fmt.Sprint(r)))
		}
	}()
	var (
		JsonData = map[string]interface{}{}
	)
	req := HttpRequest.NewRequest()
	defer func() {
		if err != nil {
			err = errors.New("请求接口" + ApiUrl + "发生错误:" + err.Error())
		}
	}()
	req.SetHeaders(map[string]string{"Content-Type": "application/json", "token": Token})
	resp, err := req.JSON().Delete(ApiUrl, Data)
	if err == nil && resp != nil {
		body, err := resp.Body()
		if err == nil {
			err = json.Unmarshal(body, &JsonData)
		}
		_ = resp.Close()
	}
	return JsonData, err
}
func RequestApiGet(ApiUrl, Token string, Params map[string]interface{}) (interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			Log.Error(errors.New(fmt.Sprint(r)))
		}
	}()
	var (
		Data     interface{}
		JsonData = map[string]interface{}{}
	)
	defer func() {
		if err != nil {
			err = errors.New("请求接口:" + ApiUrl + "发生错误:" + err.Error())
		}
	}()
	req := HttpRequest.NewRequest()
	req.SetHeaders(map[string]string{"Content-Type": "application/json", "token": Token})
	resp, err := req.JSON().Get(ApiUrl, Params)
	if err == nil && resp != nil {
		body, err := resp.Body()
		if err == nil {
			err = json.Unmarshal(body, &JsonData)
			if err == nil && JsonData["success"] != nil && JsonData["success"].(bool) {
				Data = JsonData["data"]
			}
		}
		_ = resp.Close()
	}
	return Data, err
}
func RequestGet(Url string, Params map[string]interface{}) (map[string]interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			Log.Error(errors.New(fmt.Sprint(r)))
		}
	}()
	var (
		JsonData = map[string]interface{}{}
	)
	defer func() {
		if err != nil {
			err = errors.New("请求接口:" + Url + "发生错误:" + err.Error())
		}
	}()
	req := HttpRequest.NewRequest()
	req.SetHeaders(map[string]string{"Content-Type": "application/json"})
	resp, err := req.JSON().Get(Url, Params)
	if err == nil && resp != nil {
		body, err := resp.Body()
		if err == nil {
			err = json.Unmarshal(body, &JsonData)
		}
		_ = resp.Close()
	}
	return JsonData, err
}
