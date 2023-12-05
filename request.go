package builder

import (
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"
)

type Request struct {
	client     *Client       // 指向 Client 的指针
	RequestRaw *http.Request // 指向 http.Request 的指针
	QueryParam sync.Map
	FormData   sync.Map
	Cookies    []*http.Cookie
}

func (request *Request) SetJsonBody(v interface{}) {
	var jsonString string
	switch value := v.(type) {
	case map[string]interface{}, map[string]string, *map[string]interface{}:
		jsonMarshal, err := request.client.JSONMarshal(value)
		if err != nil {
			log.Println("SetBody error:", err)
			return
		}
		jsonString = string(jsonMarshal)
	case string:
		if gjson.Valid(value) {
			jsonString = value
		}
	case []byte:
		if gjson.Valid(string(value)) {
			jsonString = string(value)
		}
	default:
		if reflect.TypeOf(value).Kind() == reflect.Struct || reflect.TypeOf(value).Kind() == reflect.Ptr {
			jsonMarshal, err := request.client.JSONMarshal(value)
			if err != nil {
				log.Println(err)
				return
			}
			jsonString = string(jsonMarshal)
		}
	}
	if jsonString != "" {
		request.SetHeader("Content-Type", "application/json;charset=UTF-8")
		request.RequestRaw.ContentLength = int64(len(jsonString))
		request.RequestRaw.Body = io.NopCloser(strings.NewReader(jsonString))
	}
}

// SetHeader 方法用于设置 HTTP 请求的 Header 部分。它接收两个 string 类型的参数，
func (request *Request) SetHeader(key, value string) *Request {
	//request.client.Lock()
	//defer request.client.Unlock()
	request.RequestRaw.Header.Set(key, value)
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
	request.client.Lock()
	defer request.client.Unlock()
	request.RequestRaw.AddCookie(cookie)
	return request
}

// SetQueryParams 方法用于设置 HTTP 请求的 Query 部分。它接收一个 map[string]interface{} 类型的参数，
func (request *Request) SetQueryParams(query map[string]any) *Request {
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
func (request *Request) SetFormDataMany(params map[string]any) *Request {
	for key, value := range params {
		request.SetFormData(key, value)
	}
	return request
}

// SetQueryString 方法用于设置 HTTP 请求的 Query 部分。它接收一个 string 类型的参数，
func (request *Request) SetQueryString(query string) *Request {
	params, err := url.ParseQuery(strings.TrimSpace(query))
	if err == nil {
		for key, value := range params {
			request.SetQueryParam(key, value[0])
		}
	} else {
		log.Println("SetQueryString url.ParseQuery error:", err)
	}
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
	return request.RequestRaw.URL.Host
}

// GetPath 方法用于获取 HTTP 请求的 Path 部分的字符串。
func (request *Request) GetPath() string {
	return request.RequestRaw.URL.Path
}

// GetUrl 方法用于获取 HTTP 请求的 URL 部分的字符串。
func (request *Request) GetUrl() string {
	return request.RequestRaw.URL.String()
}

// GetProto 方法用于获取 HTTP 请求的 Proto 部分的字符串。
func (request *Request) GetProto() string {
	return request.RequestRaw.Proto
}

// GetMethod 方法用于获取 HTTP 请求的 Method 部分的字符串。
func (request *Request) GetMethod() string {
	return request.RequestRaw.Method
}

// GetRequestHeader 方法用于获取 HTTP 请求的 Header 部分的 http.Header。
func (request *Request) GetRequestHeader() http.Header {
	return request.RequestRaw.Header
}
