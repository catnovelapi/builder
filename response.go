package builder

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
)

const (
	// MethodGet HTTP method
	MethodGet = "GET"

	// MethodPost HTTP method
	MethodPost = "POST"

	// MethodPut HTTP method
	MethodPut = "PUT"

	// MethodDelete HTTP method
	MethodDelete = "DELETE"

	// MethodPatch HTTP method
	MethodPatch = "PATCH"

	// MethodHead HTTP method
	MethodHead = "HEAD"

	// MethodOptions HTTP method
	MethodOptions = "OPTIONS"
)

type Response struct {
	Request       *http.Request
	Result        string         // 响应体字符串结果
	ResponseRaw   *http.Response // 指向 http.Response 的指针
	RequestSource *Request       // 指向 Request 的指针
}

// newParseUrl 方法用于解析 URL。它接收一个 string 类型的参数，该参数表示 HTTP 请求的 Path 部分。
func (request *Request) newParseUrl(path string) (*url.URL, error) {
	// 如果 baseUrl 不为空，且 path 不是以 / 开头，则在 path 前加上 /
	if request.client.GetClientBaseURL() == "" && path == "" {
		err := fmt.Errorf("request Error: %s", "baseUrl is empty")
		request.client.LogError(err, path, "response.go", "newParseUrl")
		return nil, err
	}
	if request.client.GetClientBaseURL() != "" && path != "" {
		if path[0] != '/' {
			path = "/" + path
		}
	}
	// 解析 URL, 如果失败则返回错误
	u, err := url.Parse(request.client.GetClientBaseURL() + path)
	if err != nil {
		return nil, err
	}
	request.URL = u
	urlRawQuery := request.GetQueryParamsEncode()
	if request.URL.RawQuery == "" {
		request.URL.RawQuery = urlRawQuery
	} else {
		request.URL.RawQuery = request.URL.RawQuery + "&" + urlRawQuery
	}
	return u, nil
}

// newRequestWithContext 方法用于创建一个 HTTP 请求。它接收一个 string 类型的参数，该参数表示 HTTP 请求的 Path 部分。
func (request *Request) newRequestWithContext() (*http.Request, error) {
	defer func() {
		request.client.log.WithFields(newFormatRequestLogText(request)).Debug("request debug")
		_, _ = request.client.log.Out.Write([]byte("------------------------------------------------------------------------------\n"))
	}()
	req, err := http.NewRequestWithContext(request.ctx, request.Method, request.URL.String(), request.bodyBuf)
	if err != nil {
		request.client.LogError(err, request.Method, "response.go", "http.NewRequestWithContext")
		return nil, err
	}
	// 设置请求头
	req.Header = request.GetRequestHeader()
	for _, v := range request.Cookies {
		req.AddCookie(v)
	}
	return req, nil
}

func (request *Request) newResponse(method, path string) (*Response, error) {
	var err error
	var response *Response
	defer func() {
		if request.client.GetClientDebug() {
			request.client.log.WithFields(newFormatResponseLogText(response)).Debug("response debug")
			_, _ = request.client.log.Out.Write([]byte("------------------------------------------------------------------------------\n"))
		}
	}()
	request.Method = method
	if _, err = request.newParseUrl(path); err != nil {
		return nil, err
	}
	// 设置请求方法, 如果请求方法为 GET, 则不设置请求体
	if err = parseRequestBody(request); err != nil {
		request.client.LogError(err, path, "util.go", "parseRequestBody")
		return nil, err
	}
	if request.bodyBuf == nil {
		request.bodyBuf = &bytes.Buffer{}
	}
	request.NewRequest, err = request.newRequestWithContext()
	if err != nil {
		return nil, err
	}
	if request.client.GetClientRetryNumber() == 0 {
		// 如果重试次数为 0，则设置重试次数为 1
		request.client.SetRetryCount(1)
	}
	response, err = request.newDoRequest()
	if err != nil {
		request.client.LogError(err, path, "response.go", "newDoRequest")
		return nil, err
	}
	if request.client.setResultFunc != nil {
		response.Result, err = request.client.setResultFunc(response.String())
		if err != nil || response.Result == "" {
			request.client.LogError(err, path, "response.go", "setResultFunc")
			response.Result = response.String()
			return nil, err
		}
	} else {
		response.Result = response.String()
	}
	return response, nil
}

// newDoResponse 方法用于执行 HTTP 请求。它接收一个 Response 对象的指针，表示 HTTP 请求的响应。
func (request *Request) newDoRequest() (*Response, error) {
	var err error
	var raw *http.Response
	for i := 0; i < request.client.GetClientRetryNumber(); i++ {
		raw, err = request.client.httpClientRaw.Do(request.NewRequest)
		if err != nil {
			request.client.LogError(err, fmt.Sprintf("retry:%v", i), "response.go", "httpClientRaw.Do")
			continue
		}
		return &Response{RequestSource: request, ResponseRaw: raw, Request: request.NewRequest}, nil
	}
	return nil, fmt.Errorf("request Error: %s", err.Error())
}

// Get 方法用于创建一个 GET 请求。它接收一个 string 类型的参数，表示 HTTP 请求的路径。
func (request *Request) Get(url string) (*Response, error) {
	return request.newResponse(MethodGet, url)
}

// Post 方法用于创建一个 POST 请求。它接收一个 string 类型的参数，表示 HTTP 请求的路径。
func (request *Request) Post(url string) (*Response, error) {
	return request.newResponse(MethodPost, url)
}

// Put 方法用于创建一个 PUT 请求。它接收一个 string 类型的参数，表示 HTTP 请求的路径。
func (request *Request) Put(url string) (*Response, error) {
	return request.newResponse(MethodPut, url)
}

// Delete 方法用于创建一个 DELETE 请求。它接收一个 string 类型的参数，表示 HTTP 请求的路径。
func (request *Request) Delete(url string) (*Response, error) {
	return request.newResponse(MethodDelete, url)
}

// Patch 方法用于创建一个 PATCH 请求。它接收一个 string 类型的参数，表示 HTTP 请求的路径。
func (request *Request) Patch(url string) (*Response, error) {
	return request.newResponse(MethodPatch, url)
}

// Head 方法用于创建一个 HEAD 请求。它接收一个 string 类型的参数，表示 HTTP 请求的路径。
func (request *Request) Head(url string) (*Response, error) {
	return request.newResponse(MethodHead, url)
}

// Options 方法用于创建一个 OPTIONS 请求。它接收一个 string 类型的参数，表示 HTTP 请求的路径。
func (request *Request) Options(url string) (*Response, error) {
	return request.newResponse(MethodOptions, url)
}
