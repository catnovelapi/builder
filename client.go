package builder

import (
	"encoding/base64"
	"fmt"
	"github.com/EDDYCJY/fake-useragent"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// Client 类型用于存储 HTTP 请求的相关信息。
type Client struct {
	sync.RWMutex                // 用于保证线程安全
	MaxConcurrent chan struct{} // 用于限制并发数
	timeout       int           // timeout 用于存储 HTTP 请求的 Timeout 部分
	retryNum      int           // retryNum 用于存储重试次数
	baseUrl       string        // baseUrl 用于存储 HTTP 请求的 BaseUrl 部分
	debug         bool          // debug 用于存储是否输出调试信息
	debugFile     *os.File      // debugFile 用于存储调试信息的文件
	clientRaw     *http.Client  // clientRaw 用于存储 http.Client 的指针
	headers       http.Header   // headers 用于存储 HTTP 请求的 Header 部分
	queryParams   url.Values    // queryParams 用于存储 HTTP 请求的 Query 部分
	body          interface{}   // body 用于存储 HTTP 请求的 Body 部分
}

// NewClient 方法用于创建一个新的 Client 对象, 并返回该对象的指针。
func NewClient() *Client {
	client := &Client{
		MaxConcurrent: make(chan struct{}, 500), // 用于限制并发数, 最大并发数为 500
		queryParams:   make(url.Values),         // 初始化 queryParams
		headers:       make(http.Header),        // 初始化 headers
		clientRaw: &http.Client{
			Transport: &http.Transport{}, // 初始化 Transport
		},
	}
	client.SetTimeout(30)                 // 默认超时时间为 30 秒
	client.SetRetryCount(3)               // 默认重试次数为 3 次
	client.SetUserAgent(browser.Random()) // 默认 User-Agent 为随机生成的浏览器 User-Agent
	return client
}

// SetBaseURL 方法用于设置HTTP请求的 BaseUrl 部分。它接收一个 string 类型的参数，该参数表示 BaseUrl 的值。
func (client *Client) SetBaseURL(baseUrl string) *Client {
	client.baseUrl = strings.TrimRight(baseUrl, "/")
	return client
}

// SetDebugFile 方法用于设置输出调试信息的文件。它接收一个 string 类型的参数，该参数表示文件名。
func (client *Client) SetDebugFile(name string) *Client {
	if fileInfo, err := os.Stat(name + ".txt"); err != nil {
		if !os.IsNotExist(err) {
			// Other error occurred
			log.Println(err)
		}
	} else {
		if fileInfo.Size() > 1024*1024 {
			newName := name + fileInfo.ModTime().Format("20060102") + ".txt"
			if err = os.Rename(name+".txt", newName); err != nil {
				log.Println(err)
			}
		}
	}
	file, err := os.OpenFile(name+".txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Println("SetDebugFile error: ", err)
	} else {
		client.debugFile = file
	}
	return client
}

// SetCookie 方法用于设置 HTTP 请求的 Cookie 部分。它接收一个 string 类型的参数，该参数表示 Cookie 的值。
func (client *Client) SetCookie(cookie string) *Client {
	if client.headers.Get("Cookie") != "" {
		// 如果已经设置了 Cookie，那么将新的 Cookie 追加到原有的 Cookie 后面
		client.headers.Set("Cookie", client.headers.Get("Cookie")+";"+cookie)
	} else {
		client.headers.Set("Cookie", cookie)
	}
	return client
}

// SetCookieJar 方法用于设置 HTTP 请求的 CookieJar 部分。它接收一个 http.CookieJar 类型的参数，该参数表示 CookieJar 的值。
func (client *Client) SetCookieJar(cookieJar http.CookieJar) *Client {
	client.clientRaw.Jar = cookieJar
	return client
}

// SetDebug 方法用于设置是否输出调试信息,如果调用该方法，那么将输出调试信息。
func (client *Client) SetDebug() *Client {
	client.debug = true
	return client
}

// SetRetryCount 方法用于设置重试次数。它接收一个 int 类型的参数，该参数表示重试次数。
func (client *Client) SetRetryCount(num int) *Client {
	if num <= 0 {
		log.Println("retry number must be greater than 0")
	} else {
		client.retryNum = num
	}
	return client
}

// SetHeader 方法用于设置 HTTP 请求的 Header 部分。它接收两个 string 类型的参数，
func (client *Client) SetHeader(key string, value interface{}) *Client {
	// 将 value 转换为 string 类型, 并将其存储到 headers 中
	client.headers.Set(key, fmt.Sprintf("%v", value))
	return client
}

// SetHeaders 方法用于设置 HTTP 请求的 Header 部分。它接收一个 map[string]interface{} 类型的参数，
func (client *Client) SetHeaders(headers map[string]interface{}) *Client {
	if headers != nil {
		for key, value := range headers {
			client.SetHeader(key, value)
		}
	}
	return client
}

// SetUserAgent 方法用于设置 HTTP 请求的 User-Agent 部分。它接收一个 string 类型的参数，该参数表示 User-Agent 的值。
func (client *Client) SetUserAgent(userAgent string) *Client {
	client.SetHeader("User-Agent", userAgent)
	return client
}

// SetQueryParam 方法用于设置 HTTP 请求的 Query 部分。它接收两个 string 类型的参数，
func (client *Client) SetQueryParam(key string, value any) *Client {
	client.queryParams.Add(key, fmt.Sprintf("%v", value))
	return client
}

// SetQueryParams 方法用于设置 HTTP 请求的 Query 部分。它接收一个 map[string]interface{} 类型的参数，
func (client *Client) SetQueryParams(params map[string]any) *Client {
	for key, value := range params {
		client.SetQueryParam(key, value)
	}
	return client
}

// SetFormDataQueryParams 方法用于设置 HTTP 请求的 Query 部分。它接收一个 url.Values 类型的参数，
func (client *Client) SetFormDataQueryParams(params url.Values) *Client {
	for key, value := range params {
		client.SetQueryParam(key, value)
	}
	return client
}

// SetQueryParamString 方法用于设置 HTTP 请求的 Query 部分。它接收一个 string 类型的参数，
func (client *Client) SetQueryParamString(query string) *Client {
	// 将 query 解析为 url.Values 类型的参数
	params, err := url.ParseQuery(strings.TrimSpace(query))
	if err == nil {
		// 将 params 中的参数存储到 queryParams 中
		client.SetFormDataQueryParams(params)
	} else {
		log.Println("SetQueryString url.ParseQuery error:", err)
	}
	return client
}

// SetProxy 方法用于设置 HTTP 请求的 Proxy 部分。它接收一个 string 类型的参数，该参数表示 Proxy 的值。
func (client *Client) SetProxy(proxy string) *Client {
	u, err := url.Parse(proxy)
	if err != nil {
		log.Println("SetProxy url.Parse error:", err)
		return client
	}
	// 设置 Transport 的 Proxy 字段
	client.clientRaw.Transport = &http.Transport{Proxy: http.ProxyURL(u)}
	return client
}

// SetTimeout 方法用于设置 HTTP 请求的 Timeout 部分, timeout 单位为秒。它接收一个 int 类型的参数，该参数表示 Timeout 的值。
func (client *Client) SetTimeout(timeout int) *Client {
	// 设置 clientRaw 的 Timeout 字段, timeout 单位为秒
	client.clientRaw.Timeout = time.Duration(timeout * int(time.Second))
	return client
}

// SetBasicAuth 方法用于设置 HTTP 请求的 BasicAuth 部分。它接收两个 string 类型的参数，分别表示用户名和密码。
func (client *Client) SetBasicAuth(username, password string) *Client {
	auth := username + ":" + password
	client.SetHeader("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
	return client
}

// GetClientQueryParams 方法用于获取 HTTP 请求的 Query 部分。它返回一个 url.Values 类型的参数。
func (client *Client) GetClientQueryParams() url.Values {
	return client.queryParams
}

// GetClientBody 方法用于获取 HTTP 请求的 Body 部分。它返回一个 interface{} 类型的参数。
func (client *Client) GetClientBody() interface{} {
	return client.body
}

// GetClientQueryParamsEncode 方法用于获取 HTTP 请求的 Query 部分。它返回一个 Encode 后的 string 类型的参数。
func (client *Client) GetClientQueryParamsEncode() string {
	return client.queryParams.Encode()
}

// GetClientHeaders 方法用于获取 HTTP 请求的 Header 部分。它返回一个 http.Header 类型的参数。
func (client *Client) GetClientHeaders() http.Header {
	return client.headers
}

// GetClientBaseURL 方法用于获取 HTTP 请求的 BaseUrl 部分。它返回一个 string 类型的参数。
func (client *Client) GetClientBaseURL() string {
	return client.baseUrl
}

// GetClientDebug 方法用于获取 HTTP 请求的 Debug 部分。它返回一个 bool 类型的参数。
func (client *Client) GetClientDebug() bool {
	return client.debug
}

// GetClientDebugFile 方法用于获取 HTTP 请求的 DebugFile 部分。它返回一个 *os.File 类型的参数。
func (client *Client) GetClientDebugFile() *os.File {
	return client.debugFile
}

// GetClientRetryNumber 方法用于获取 HTTP 请求的 RetryNumber 部分。它返回一个 int 类型的参数。
func (client *Client) GetClientRetryNumber() int {
	return client.retryNum
}

// GetClientTimeout 方法用于获取 HTTP 请求的 Timeout 部分。它返回一个 int 类型的参数。
func (client *Client) GetClientTimeout() int {
	return client.timeout
}

// GetClientCookie 方法用于获取 HTTP 请求的 Cookie 部分。它返回一个 string 类型的参数。
func (client *Client) GetClientCookie() string {
	return client.headers.Get("Cookie")
}
