package builder

import (
	"bytes"
	"fmt"
	"golang.org/x/net/context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type Request struct {
	URL        *url.URL
	ctx        context.Context
	Method     string // HTTP 请求的 Method 部分
	Body       interface{}
	bodyBuf    *bytes.Buffer
	bodyBytes  []byte
	client     *Client // 指向 Client 的指针
	Header     sync.Map
	QueryParam sync.Map
	FormData   sync.Map
	Cookies    []*http.Cookie
}

func (request *Request) SetBody(v interface{}) *Request {
	request.Body = v
	return request
}

// SetHeader 方法用于设置 HTTP 请求的 Header 部分。它接收两个 string 类型的参数，
func (request *Request) SetHeader(key, value string) *Request {
	//request.RequestRaw.Header.Set(key, value)
	request.Header.Store(key, value)
	return request
}

func (request *Request) SetHeaders(headers map[string]string) *Request {
	for key, value := range headers {
		request.SetHeader(key, value)
	}
	return request
}

// SetCookies 方法用于设置 HTTP 请求的 Cookies 部分。它接收一个 []*http.Cookie 类型的参数，
func (request *Request) SetCookies(cookie []*http.Cookie) *Request {
	for _, c := range cookie {
		request.SetCookie(c)
	}
	return request
}

// SetCookie 方法用于设置 HTTP 请求的 Cookie 部分。它接收一个 *http.Cookie 类型的参数，
func (request *Request) SetCookie(cookie *http.Cookie) *Request {
	request.Cookies = append(request.Cookies, cookie)
	return request
}

// SetQueryParams 方法用于设置 HTTP 请求的 Query 部分。它接收一个 map[string]interface{} 类型的参数，
func (request *Request) SetQueryParams(query map[string]string) *Request {
	for key, value := range query {
		request.SetQueryParam(key, value)
	}
	return request
}

// SetQueryParam 方法用于设置 HTTP 请求的 Query 部分。它接收两个 string 类型的参数，
func (request *Request) SetQueryParam(key string, value interface{}) *Request {
	request.QueryParam.Store(key, fmt.Sprintf("%v", value))
	return request
}
func (request *Request) SetFormData(key string, value any) *Request {
	request.FormData.Store(key, fmt.Sprintf("%v", value))
	return request
}
func (request *Request) SetFormDataMany(params map[string]string) *Request {
	for key, value := range params {
		request.SetFormData(key, value)
	}
	return request
}

// SetQueryString 方法用于设置 HTTP 请求的 Query 部分。它接收一个 string 类型的参数，
func (request *Request) SetQueryString(query string) *Request {
	if params, err := url.ParseQuery(strings.TrimSpace(query)); err == nil {
		for key, value := range params {
			request.SetQueryParam(key, value[0])
		}
	} else {
		request.client.LogError(err, query, "request.go", "SetQueryString")
	}
	return request
}

func (request *Request) SetHeaderContentType(contentType string) *Request {
	request.SetHeader("Content-Type", contentType)
	return request
}

// GetQueryParamsEncode 方法用于获取 HTTP 请求的 Query 部分的 URL 编码字符串。
func (request *Request) GetQueryParamsEncode() string {
	var parts []string
	request.QueryParam.Range(func(key, value interface{}) bool {
		k, _ := key.(string)
		v, _ := value.(string)
		part := fmt.Sprintf("%s=%s", url.QueryEscape(k), url.QueryEscape(v))
		parts = append(parts, part)
		return true
	})
	return strings.Join(parts, "&")
}

// GetFormDataEncode 方法用于获取 HTTP 请求的 Query 部分的 URL 编码字符串。
func (request *Request) GetFormDataEncode() string {
	var parts []string
	request.FormData.Range(func(key, value interface{}) bool {
		k, _ := key.(string)
		v, _ := value.(string)
		part := fmt.Sprintf("%s=%s", url.QueryEscape(k), url.QueryEscape(v))
		parts = append(parts, part)
		return true
	})
	return strings.Join(parts, "&")
}

// GetQueryParamsNopCloser 方法用于获取 HTTP 请求的 Query 部分的 ReadCloser。
func (request *Request) GetQueryParamsNopCloser() io.ReadCloser {
	// 将字符串转换为 io.ReadCloser, 并返回
	return io.NopCloser(strings.NewReader(request.GetQueryParamsEncode()))
}

// GetQueryParams 方法用于获取 HTTP 请求的 Query 部分的 url.Values。
//func (request *Request) GetQueryParams() url.Values {
//	return request.QueryParam
//}

// GetHost 方法用于获取 HTTP 请求的 Host 部分的字符串。
func (request *Request) GetHost() string {
	return request.client.baseUrl
}

// GetPath 方法用于获取 HTTP 请求的 Path 部分的字符串。
func (request *Request) GetPath() string {
	return request.URL.Path
}

// GetUrl 方法用于获取 HTTP 请求的 URL 部分的字符串。
func (request *Request) GetUrl() string {
	return request.URL.String()
}

// GetMethod 方法用于获取 HTTP 请求的 Method 部分的字符串。
func (request *Request) GetMethod() string {
	return request.Method
}

// GetRequestHeader 方法用于获取 HTTP 请求的 Header 部分的 http.Header。
func (request *Request) GetRequestHeader() http.Header {
	header := make(http.Header)
	request.Header.Range(func(key, value interface{}) bool {
		keyStr, _ := key.(string)
		valueStr, _ := value.(string)
		if keyStr != "" && valueStr != "" {
			header.Add(keyStr, valueStr)
		}
		return true
	})
	return header
}
func (request *Request) GetHeaderContentType() string {
	for key, value := range request.GetRequestHeader() {
		if key == "Content-Type" {
			return value[0]
		}
	}
	return ""
}
