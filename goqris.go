package goqris

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"slices"
	"strconv"
	"time"
)

var (
	successCodes    = []string{"00", "200"}
	successMessages = []string{"Successful", "Successfully", "success", "Payment Success"}
)

type IQris interface {
	GenerateQRCode(r IRequest) (M, error)
	CheckStatusTransaction(r IRequest) (M, error)
}

type IRequest interface {
	Payload() M
}

type M map[string]any

func (m M) GetValue(k string) string {
	v, _ := m[k].(string)
	return v
}

func (m M) SetValueIfEmpty(k string, v any) {
	if _, ok := m[k].(string); !ok {
		m[k] = v
	}
}

func (m M) SetValueIfEmptyWithFunc(k string, f func(M) M) {
	if v, ok := m[k].(M); ok {
		f(v)
	} else {
		m[k] = f(M{})
	}
}

func (m M) SuccessfulGenerate() bool {
	var (
		code    string
		message string
	)

	if code = m.GetValue("code"); code == "" {
		code = m.GetValue("responseCode")
		if len(code) >= 7 {
			code = code[:3]
		}
	}

	if message = m.GetValue("message"); message == "" {
		message = m.GetValue("responseMessage")
	}

	return slices.Contains(successCodes, code) && slices.Contains(successMessages, message)
}

func (m M) SuccessfulPaid() bool {
	var (
		code    string
		message string
	)

	if code = m.GetValue("payment_status"); code == "" {
		code = m.GetValue("latestTransactionStatus")
	}

	if message = m.GetValue("payment_description"); message == "" {
		message = m.GetValue("transactionStatusDesc")
	}

	return slices.Contains(successCodes, code) && slices.Contains(successMessages, message)
}

func (m M) RefNo() string {
	var refNo string
	if refNo = m.GetValue("bill_number"); refNo == "" {
		refNo = m.GetValue("referenceNo")
	}
	return refNo
}

func (m M) ServiceCode() string {
	sc := m.GetValue("responseCode")
	if len(sc) >= 7 {
		sc = sc[3:5]
	}
	return sc
}

// cache access token
var cache = map[string]map[string]string{}

func getAccessTokenFromCache(key string) string {
	c, ok := cache[key]
	if !ok {
		return ""
	}

	expStr, ok := c["expired_at"]
	if !ok {
		return ""
	}

	expInt, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		return ""
	}

	if time.Unix(expInt, 0).Before(time.Now().UTC()) {
		return ""
	}

	accessToken, ok := c["access_token"]
	if !ok {
		return ""
	}

	return accessToken
}

func setAccessTokenToCache(key string, data map[string]any) (string, error) {
	accessToken, ok := data["access_token"].(string)
	if !ok {
		accessToken, ok = data["accessToken"].(string)
		if !ok {
			return "", errors.New("the key of access_token or accessToken is undefined")
		}
	}

	expStr, ok := data["expires_in"].(string)
	if !ok {
		expStr, ok = data["expiresIn"].(string)
		if !ok {
			return "", errors.New("the key of expires_in or expiresIn is undefined")
		}
	}

	expInt, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		return "", err
	}

	expAtInt := time.Now().UTC().Add(time.Duration(expInt) * time.Second).Unix()
	expAtStr := strconv.FormatInt(expAtInt, 10)
	cache[key] = map[string]string{
		"access_token": accessToken,
		"expired_at":   expAtStr,
	}

	return accessToken, nil
}

type responseData struct {
	Data M
	Err  error
}

func makeRequest(uri string, bodyMap, headers M) responseData {
	ch := make(chan responseData)
	go func() {
		rd := responseData{}
		bodyBytes, err := json.Marshal(bodyMap)
		if err != nil {
			rd.Err = err
			ch <- rd
			close(ch)
			return
		}

		bodyBuf := bytes.NewBuffer(bodyBytes)
		req, err := http.NewRequest("POST", uri, bodyBuf)
		if err != nil {
			rd.Err = err
			ch <- rd
			close(ch)
			return
		}

		for k, v := range headers {
			value, ok := v.(string)
			if !ok {
				continue
			}

			req.Header.Set(k, value)
		}

		client := &http.Client{Timeout: 20 * time.Second}
		res, err := client.Do(req)
		if err != nil {
			rd.Err = err
			ch <- rd
			close(ch)
			return
		}

		defer res.Body.Close()
		if err = json.NewDecoder(res.Body).Decode(&rd.Data); err != nil && err != io.EOF {
			rd.Err = err
			ch <- rd
			close(ch)
			return
		}

		ch <- rd
		close(ch)
	}()

	return <-ch
}
