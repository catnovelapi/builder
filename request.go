package builder

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Request struct {
	client      *Client       // 指向 Client 的指针
	RequestRaw  *http.Request // 指向 http.Request 的指针
	queryParams url.Values    // 用于存储 Query 参数的 url.Values
}

// R 方法用于创建一个新的 Request 对象。它接收一个 string 类型的参数，该参数表示 HTTP 请求的 Path 部分。
func (client *Client) R() *Request {
	req := &http.Request{
		//Body:       rc,
		//URL:        u,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     client.GetClientHeaders(),
	}
	return &Request{
		client:      client,
		queryParams: client.GetClientQueryParams(),
		RequestRaw:  req.WithContext(context.Background()),
	}
}

// SetBody 方法用于设置 HTTP 请求的 Body 部分。它接收一个 interface{} 类型的参数，
// 该参数可以是以下几种类型：string, []byte, map[string]interface{}, map[string]string,
// map[string]int, map[string]int64, map[string]float64, map[string]float32, map[string]bool。
// 对于不支持的类型，方法会设置 ContentLength 为 -1，并将 GetBody 方法设置为返回 nil。
// 如果成功设置了 body，方法会返回 Request 指针本身，以便进行链式调用。
func (request *Request) SetBody(body interface{}) *Request {
	if body != nil {
		// 加锁以确保线程安全
		request.client.Lock()
		defer request.client.Unlock()

		// 使用 type switch 来检查 body 的实际类型
		switch v := body.(type) {
		case string: // 如果 body 是 string 类型
			// 设置 ContentLength 为字符串的长度，并将字符串转换为 ReadCloser
			request.RequestRaw.ContentLength = int64(len(v))
			request.RequestRaw.Body = io.NopCloser(strings.NewReader(v))
		case []byte: // 如果 body 是 []byte 类型
			// 设置 ContentLength 为字节数组转换为字符串后的长度，并将字节数组转换为字符串后转换为 ReadCloser
			request.RequestRaw.ContentLength = int64(len(string(v)))
			request.RequestRaw.Body = io.NopCloser(strings.NewReader(string(v)))
		case map[string]interface{}, map[string]string, map[string]int, map[string]int64, map[string]float64, map[string]float32, map[string]bool: // 如果 body 是 map 类型
			// 尝试将 map 转换为 JSON 字符串
			if m, err := json.Marshal(v); err != nil {
				// 如果转换失败，打印错误信息
				log.Println(err)
			} else {
				// 如果转换成功，设置 ContentLength 为 JSON 字符串的长度，并将 JSON 字符串转换为 ReadCloser
				request.RequestRaw.ContentLength = int64(len(string(m)))
				request.RequestRaw.Body = io.NopCloser(strings.NewReader(string(m)))
			}
		default: // 对于其他类型
			// 设置 ContentLength 为 -1，并将 GetBody 方法设置为返回 nil
			request.RequestRaw.ContentLength = -1
			request.RequestRaw.GetBody = func() (io.ReadCloser, error) {
				return nil, nil
			}
		}
	}

	// 返回 Request 指针本身，以便进行链式调用
	return request
}

// SetHeader 方法用于设置 HTTP 请求的 Header 部分。它接收两个 string 类型的参数，
func (request *Request) SetHeader(key, value string) *Request {
	request.client.Lock()
	defer request.client.Unlock()
	request.RequestRaw.Header.Set(key, value)
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
func (request *Request) SetQueryParams(query map[string]interface{}) *Request {
	for key, value := range query {
		request.SetQueryParam(key, value)
	}
	return request
}

// SetQueryParam 方法用于设置 HTTP 请求的 Query 部分。它接收两个 string 类型的参数，
func (request *Request) SetQueryParam(key string, value interface{}) *Request {
	request.client.Lock()
	defer request.client.Unlock()
	request.queryParams.Set(key, fmt.Sprintf("%v", value))
	return request
}

// SetQueryString 方法用于设置 HTTP 请求的 Query 部分。它接收一个 string 类型的参数，
func (request *Request) SetQueryString(query string) *Request {
	params, err := url.ParseQuery(strings.TrimSpace(query))
	if err == nil {
		for p, v := range params {
			for _, pv := range v {
				request.SetQueryParam(p, pv)
			}
		}
	} else {
		log.Println(err)
	}
	return request
}

// SetContentType 方法用于设置 HTTP 请求的 Content-Type 部分。它接收一个 string 类型的参数，
func (request *Request) SetContentType(contentType string) *Request {
	request.SetHeader("Content-Type", contentType)
	return request
}

// GetQueryParamsEncode 方法用于获取 HTTP 请求的 Query 部分的 URL 编码字符串。
func (request *Request) GetQueryParamsEncode() string {
	return request.GetQueryParams().Encode()

}

// GetQueryParamsNopCloser 方法用于获取 HTTP 请求的 Query 部分的 ReadCloser。
func (request *Request) GetQueryParamsNopCloser() io.ReadCloser {
	return io.NopCloser(strings.NewReader(request.GetQueryParamsEncode()))
}

// GetQueryParams 方法用于获取 HTTP 请求的 Query 部分的 url.Values。
func (request *Request) GetQueryParams() url.Values {
	return request.queryParams
}

// GetContentType 方法用于获取 HTTP 请求的 Content-Type 部分的字符串。
func (request *Request) GetContentType() string {
	return request.RequestRaw.Header.Get("Content-Type")
}

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
