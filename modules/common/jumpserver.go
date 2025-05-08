package common

import (
	"bytes"
	"gopkg.in/twindagger/httpsig.v1"
	"inner/modules/kits"
	"io/ioutil"
	"net/http"
	"time"
)

type SigAuth struct {
	KeyID    string
	SecretID string
}

func (auth *SigAuth) Sign(r *http.Request) error {
	headers := []string{"(request-target)", "date"}
	signer, err := httpsig.NewRequestSigner(auth.KeyID, auth.SecretID, "hmac-sha256")
	if err != nil {
		return err
	}
	return signer.SignRequest(r, headers, nil)
}

func JumpServerApi(url, AccessKeyID, AccessKeySecret, method string, data map[string]interface{}) ([]byte, error) {
	var (
		req   *http.Request
		resp  *http.Response
		param string
		body  []byte
	)
	auth := SigAuth{KeyID: AccessKeyID, SecretID: AccessKeySecret}
	gmtFmt := "Mon, 02 Jan 2006 15:04:05 GMT"
	client := &http.Client{}
	if method == "GET" {
		if len(data) > 0 {
			for k, v := range data {
				if param == "" {
					param = k + "=" + v.(string)
				} else {
					param = k + "=" + v.(string) + "&" + param
				}
			}
		}
		if param != "" {
			param = "?" + param
		}
		req, err = http.NewRequest(method, url+param, nil)
	} else {
		req, err = http.NewRequest(method, url, bytes.NewBuffer([]byte(kits.MapToJson(data))))
	}
	if err == nil {
		if method != "GET" {
			req.Header.Add("Content-Type", "application/json")
		}
		req.Header.Add("Date", time.Now().Format(gmtFmt))
		req.Header.Add("Accept", "application/json")
		req.Header.Add("X-JMS-ORG", "00000000-0000-0000-0000-000000000002")
		err = auth.Sign(req)
		if err == nil {
			resp, err = client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				body, err = ioutil.ReadAll(resp.Body)
			}
		}
	}
	return body, err
}
