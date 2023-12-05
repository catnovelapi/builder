package builder

import (
	"net/http"
	"net/url"
)

// GetClientQueryParams 方法用于获取 HTTP 请求的 Query 部分。它返回一个 url.Values 类型的参数。
func (client *Client) GetClientQueryParams() url.Values {
	return client.QueryParam
}

// GetClientBody 方法用于获取 HTTP 请求的 Body 部分。它返回一个 interface{} 类型的参数。
func (client *Client) GetClientBody() interface{} {
	return client.body
}

// GetClientQueryParamsEncode 方法用于获取 HTTP 请求的 Query 部分。它返回一个 Encode 后的 string 类型的参数。
func (client *Client) GetClientQueryParamsEncode() string {
	return client.QueryParam.Encode()
}

// GetClientHeaders 方法用于获取 HTTP 请求的 Header 部分。它返回一个 http.Header 类型的参数。
func (client *Client) GetClientHeaders() http.Header {
	return client.Header
}

// GetClientBaseURL 方法用于获取 HTTP 请求的 BaseUrl 部分。它返回一个 string 类型的参数。
func (client *Client) GetClientBaseURL() string {
	return client.baseUrl
}

// GetClientDebug 方法用于获取 HTTP 请求的 Debug 部分。它返回一个 bool 类型的参数。
func (client *Client) GetClientDebug() bool {
	return client.debug
}

// GetClientRetryNumber 方法用于获取 HTTP 请求的 RetryNumber 部分。它返回一个 int 类型的参数。
func (client *Client) GetClientRetryNumber() int {
	return client.RetryCount
}

// GetClientTimeout 方法用于获取 HTTP 请求的 Timeout 部分。它返回一个 int 类型的参数。
func (client *Client) GetClientTimeout() int {
	return client.timeout
}

// GetClientCookie 方法用于获取 HTTP 请求的 Cookie 部分。它返回一个 string 类型的参数。
func (client *Client) GetClientCookie() string {
	return client.Header.Get("Cookie")
}
