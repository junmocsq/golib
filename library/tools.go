package library

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"
)

// map参数转换为url参数形式
func MapToUrlParams(baseUrl string, m map[string]string) (string, error) {
	params := url.Values{}
	Url, err := url.Parse(baseUrl)
	if err != nil {
		return "", err
	}
	for k, v := range m {
		params.Set(k, v)
	}
	//如果参数中有中文参数,这个方法会进行URLEncode
	Url.RawQuery = params.Encode()
	urlPath := Url.String()
	return urlPath, nil
}

/**
获取提交参数数组
*/
func CpOpenToken(params map[string]string, token string) error {
	params["timestamp"] = strconv.Itoa(int(time.Now().Unix()))
	sortKeyArr := make([]string, 0)
	for k, _ := range params {
		sortKeyArr = append(sortKeyArr, k)
	}
	sort.Strings(sortKeyArr)
	var s string
	for _, v := range sortKeyArr {
		s += v + params[v]
	}
	s += token
	t, err := Md5(s)
	if err != nil {
		return err
	}
	params["token"] = t
	return nil
}

func Md5(ss string) (res string, err error) {

	md5Ctx := md5.New()
	_, err = md5Ctx.Write([]byte(ss))
	if err != nil {
		return
	}
	res = hex.EncodeToString(md5Ctx.Sum(nil))
	return
}

func HttpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		// handle error
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
		return nil, err
	}
	return body, nil
}

func HttpPost(url string, params map[string]string) ([]byte, error) {
	b, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
		return nil, err
	}
	return body, nil
}
