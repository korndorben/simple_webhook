package main

import (
	"bytes"
	"net/http"
	"io/ioutil"
	"fmt"
	"encoding/json"
	"crypto/md5"
	"encoding/hex"
)

func Md5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// 通用的http post请求方法
func PostJSON(uri string, obj interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	jsonData = bytes.Replace(jsonData, []byte("\\u003c"), []byte("<"), -1)
	jsonData = bytes.Replace(jsonData, []byte("\\u003e"), []byte(">"), -1)
	jsonData = bytes.Replace(jsonData, []byte("\\u0026"), []byte("&"), -1)

	body := bytes.NewBuffer(jsonData)
	response, err := http.Post(uri, "application/json;charset=utf-8", body)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close() // 延迟释放资源

	if response.StatusCode != http.StatusOK {
		resp, _ := ioutil.ReadAll(response.Body)
		return nil, fmt.Errorf("http get error : uri=%v , statusCode=%v, %s", uri, response.StatusCode, string(resp))
	}
	return ioutil.ReadAll(response.Body)
}
