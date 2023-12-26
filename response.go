package builder

import (
	"bytes"
	"fmt"
	"github.com/tidwall/gjson"
	"net/http"
	"net/url"
	"reflect"
	"strings"
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

	plainTextType = "text/plain; charset=utf-8"

	jsonContentType = "application/json"

	formContentType = "application/x-www-form-urlencoded"
)

type Response struct {
	Request       *http.Request
	Result        string         // 响应体字符串结果
	ResponseRaw   *http.Response // 指向 http.Response 的指针
	RequestSource *Request       // 指向 Request 的指针
}

// newParseUrl 方法用于解析 URL。它接收一个 string 类型的参数，该参数表示 HTTP 请求的 Path 部分。
func (request *Request) newParseUrl(path string) (*url.URL, error) {
	var err error
	baseURL := request.client.GetClientBaseURL()

	// Return an error if both the base URL and the path are empty
	if baseURL == "" && path == "" {
		err = fmt.Errorf("request Error: baseUrl and path are empty")
		request.client.LogError(err, path, "response.go", "newParseUrl")
		return nil, err
	}

	// Ensure path is properly prefixed with a "/"
	if path != "" && !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	fullURL := baseURL + path
	request.URL, err = url.Parse(fullURL)
	if err != nil {
		request.client.LogError(err, fullURL, "response.go", "newParseUrl")
		return nil, err
	}
	// Set URL and append query parameters
	return request.URL, nil
}

// newRequestWithContext 方法用于创建一个 HTTP 请求。它接收一个 string 类型的参数，该参数表示 HTTP 请求的 Path 部分。
func (request *Request) newRequestWithContext() (*http.Request, error) {
	defer func() {
		if request.client.GetClientDebug() {
			request.client.log.WithFields(newFormatRequestLogText(request)).Debug("request debug")
		}
	}()
	newParamsEncode := request.GetQueryParamsEncode()
	if newParamsEncode != "" {
		if request.Method == MethodGet {
			if request.URL.RawQuery != "" {
				request.URL.RawQuery += "&"
			}
			request.URL.RawQuery += newParamsEncode
		} else {
			request.bodyBuf.WriteString(newParamsEncode)
		}
	}

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
		}
	}()
	request.Method = method
	if _, err = request.newParseUrl(path); err != nil {
		return nil, err
	}
	if request.bodyBuf == nil {
		request.bodyBuf = &bytes.Buffer{}
	}
	if request.Body != nil {
		request.setBody()
	}
	request.client.httpClientRaw.Jar.SetCookies(request.URL, request.Cookies)
	request.NewRequest, err = request.newRequestWithContext()
	if err != nil {
		return nil, err
	}
	if request.client.GetClientRetryNumber() == 0 {
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

func (request *Request) setBody() {
	contentType := request.GetHeaderContentType()
	switch body := request.Body.(type) {
	case string:
		if contentType == formContentType {
			if gjson.Valid(body) {
				request.SetQueryParams(request.jsonToMap(body))
			}
		} else {
			request.bodyBuf = bytes.NewBufferString(body)
		}
	case map[string]string, map[string]interface{}:
		b := request.mapToJson(body)
		if contentType == formContentType {
			request.SetQueryParams(request.jsonToMap(b))
		} else {
			request.bodyBuf = bytes.NewBufferString(b)
		}
	default:
		kind := reflect.TypeOf(body).Kind()
		if kind == reflect.Struct || kind == reflect.Ptr {
			b := request.structToJson(body)
			if contentType == formContentType {
				request.SetQueryParams(request.jsonToMap(b))
			} else {
				request.bodyBuf = bytes.NewBufferString(b)
			}
		}
	}
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
