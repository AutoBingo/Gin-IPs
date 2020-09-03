package test

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type API struct {
	url       string
	accessKey string
	secretKey string
	header    map[string]string
}

func NewAPI(url, ak, sk string) *API {
	header := map[string]string{
		"Content-Type": "application/json",
	}
	return &API{url, ak, sk, header}
}

// 使用 params 参数尽量使用 POST请求，Get请求可能会导致参数过长截断URL, 当前时间和服务器时间不能相差超过10s
func (ez *API) Request(reqURI string, method string, params map[string]interface{}) (string, error) {
	var ret string
	nowTS := strconv.FormatInt(time.Now().Unix(), 10) // timestamp
	method = strings.ToUpper(method)
	sign, err := ez.genSignature(reqURI, method, params, nowTS)
	if err != nil {
		return "", err
	}

	urlValue := url.Values{"accesskey": {ez.accessKey}, "expires": {nowTS}, "signature": {sign}}
	client := &http.Client{Timeout: time.Duration(60) * time.Second}
	var (
		req *http.Request
	)
	if method == "GET" || method == "DELETE" {
		if params != nil {
			for k, v := range params {
				if vStr, err := CoverInterfaceToString(v); err != nil {
					return "", errors.New(fmt.Sprintf("request parameters covered error: %s", err))
				} else {
					urlValue.Add(k, vStr)
				}
			}
		}
		fullParams := urlValue.Encode()
		fullURL := fmt.Sprintf("%s?%s", "http://"+ez.url+reqURI, fullParams)
		req, err = http.NewRequest(method, fullURL, nil)
		if err != nil {
			return ret, err
		}
	} else if method == "POST" || method == "PUT" {
		paramsJson, _ := json.Marshal(params)
		reader := bytes.NewReader(paramsJson)
		fullParams := urlValue.Encode()
		fullURL := fmt.Sprintf("%s?%s", "http://"+ez.url+reqURI, fullParams)
		req, err = http.NewRequest(method, fullURL, reader)
		if err != nil {
			return ret, err
		}
	} else {
		return "", errors.New(fmt.Sprintf("Unknown Request Method: %s", err))
	}

	for k, v := range ez.header {
		req.Header.Set(k, v)
	}
	req.Host = ez.header["Host"]
	response, err := client.Do(req)
	if err != nil {
		return ret, err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return ret, err
	}
	ret = string(body)
	return ret, nil
}

// hmacSHA1Encrypt encrypt the encryptText use encryptKey
func (ez *API) hmacSHA1Encrypt(encryptText, encryptKey string) string {
	key := []byte(encryptKey)
	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(encryptText))
	var str string = hex.EncodeToString(mac.Sum(nil))
	//fmt.Printf("[encrypt result] %v\n", str)
	return str
}

func (ez *API) genSignature(uri string, method string, params map[string]interface{}, nowTS string) (string, error) {
	//fmt.Println(reflect.ValueOf(nowTS))
	urlParams, bodyContent := "", ""
	if params != nil {
		// method is GET or DELETE , params build  the url_params
		method = strings.ToUpper(method)
		if method == "GET" || method == "DELETE" {
			// sort the params
			keys := GetMapKeysSorted(params)
			for _, k := range keys {
				if vStr, err := CoverInterfaceToString(params[k]); err != nil {
					return "", errors.New(fmt.Sprintf("signed parameters covered error: %s", err))
				} else {
					urlParams = urlParams + k + vStr
				}
			}
		} else if method == "POST" || method == "PUT" {
			// method is POST or PUT, params build the bodyContent
			jsonStr, err := json.Marshal(params)
			if err != nil {
				return "", err
			}
			md5Ctx := md5.New()
			md5Ctx.Write(jsonStr)
			cipherStr := md5Ctx.Sum(nil)
			bodyContent = hex.EncodeToString(cipherStr)
		} else {
			return "", errors.New(fmt.Sprintf("Unknown Request Method: %s", method))
		}
	}

	strSign := method + "\n" + uri + "\n" + urlParams  + "\n" + bodyContent + "\n" + nowTS + "\n" + ez.accessKey
	sign := ez.hmacSHA1Encrypt(strSign, ez.secretKey)
	return sign, nil
}

// CoverInterfaceToString 将其他类型转换成字符串
func CoverInterfaceToString(inter interface{}) (string, error) {
	var (
		ret string
		err error
	)
	switch v := inter.(type) {
	case string:
		ret = v
	case int:
		ret = strconv.FormatInt(int64(v), 10)
	case float64:
		ret = strconv.FormatFloat(v, 'G', -1, 64)
	case bool:
		ret = strconv.FormatBool(v)
	default:
		retByte, _ := json.Marshal(v)
		ret = string(retByte)
	}
	return ret, err
}

func GetMapKeysSorted(oriMap map[string]interface{}) []string {
	keys := make([]string, len(oriMap))
	i := 0
	for k, _ := range oriMap {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}
