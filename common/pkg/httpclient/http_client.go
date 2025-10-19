package httpclient

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"pulse/common/pkg/logger"
	"time"
)

// Get 用于发起一个GET请求
func Get(url string, timeout int64) (result string, err error) {
	var client = &http.Client{}
	// 创建一个HTTP请求对象
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	if timeout > 0 {
		client.Timeout = time.Duration(timeout) * time.Second
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = fmt.Errorf("response status code is not 200")
		return
	}

	// 读取响应体
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("http get api url:%s send err: %s", url, err.Error()))
		return
	}

	result = string(data)
	return
}

// PostParams 用于发起一个表单参数的POST请求
func PostParams(url string, params string, timeout int64) (result string, err error) {
	var client = &http.Client{}
	// 将 `params`包装成一个`bytes.Buffer`，实现了`io.Reader`接口, 可以作为请求体
	buf := bytes.NewBufferString(params)
	// 创建一个HTTP请求对象
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return
	}
	if timeout > 0 {
		client.Timeout = time.Duration(timeout) * time.Second
	}

	// 发送请求
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = fmt.Errorf("response status code is not 200")
		return
	}

	// 读取响应体
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("http post api url:%s send send: %s", url, err.Error()))
		return
	}

	result = string(data)
	return
}

// PostJson 用于发起一个JSON参数的POST请求
func PostJson(url string, body string, timeout int64) (result string, err error) {
	var client = &http.Client{}
	// 将 `body`包装成一个`bytes.Buffer`，实现了`io.Reader`接口, 可以作为请求体
	buf := bytes.NewBufferString(body)
	// 创建一个HTTP请求对象
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return
	}
	if timeout > 0 {
		client.Timeout = time.Duration(timeout) * time.Second
	}

	// 发送请求
	req.Header.Set("Content-type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = fmt.Errorf("response status code is not 200")
		return
	}

	// 读取响应体
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("http post api url: %s send err: %s", url, err.Error()))
		return
	}

	result = string(data)
	return
}
