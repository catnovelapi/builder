package builder

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/EDDYCJY/fake-useragent"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"golang.org/x/net/publicsuffix"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

func createTransport(localAddr net.Addr) *http.Transport {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}
	if localAddr != nil {
		dialer.LocalAddr = localAddr
	}
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
	}
}

// Client 类型用于存储 HTTP 请求的相关信息。
type Client struct {
	sync.RWMutex                         // 用于保证线程安全
	MaxConcurrent          chan struct{} // 用于限制并发数
	timeout                int           // timeout 用于存储 HTTP 请求的 Timeout 部分
	baseUrl                string        // baseUrl 用于存储 HTTP 请求的 BaseUrl 部分
	log                    *logrus.Logger
	httpClientRaw          *http.Client      // httpClientRaw 用于存储 http.Client 的指针
	Header                 map[string]string // Header 用于存储 HTTP 请求的 Header 部分
	QueryParam             map[string]any    // QueryParam 用于存储 HTTP 请求的 Query 部分
	setResultFunc          func(v string) (string, error)
	Token                  string
	AuthScheme             string
	Cookies                []*http.Cookie
	Debug                  bool
	AllowGetMethodPayload  bool
	RetryCount             int
	JSONMarshal            func(v interface{}) ([]byte, error)
	JSONUnmarshal          func(data []byte, v interface{}) error
	XMLMarshal             func(v interface{}) ([]byte, error)
	XMLUnmarshal           func(data []byte, v interface{}) error
	HeaderAuthorizationKey string
	body                   interface{} // body 用于存储 HTTP 请求的 Body 部分
}

const defaultRetryCount = 3

// NewClient 方法用于创建一个新的 Client 对象, 并返回该对象的指针。
func NewClient() *Client {
	cookieJar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	client := &Client{
		MaxConcurrent:          make(chan struct{}, 500), // 用于限制并发数, 最大并发数为 500
		QueryParam:             map[string]any{},         // 初始化 QueryParam
		Header:                 map[string]string{},      // 初始化 Header
		Cookies:                make([]*http.Cookie, 0),
		log:                    logrus.New(),
		JSONMarshal:            json.Marshal,
		JSONUnmarshal:          json.Unmarshal,
		XMLMarshal:             xml.Marshal,
		XMLUnmarshal:           xml.Unmarshal,
		HeaderAuthorizationKey: http.CanonicalHeaderKey("Authorization"),
		AuthScheme:             "Bearer",
		httpClientRaw:          &http.Client{Jar: cookieJar},
	}

	if client.httpClientRaw.Transport == nil {
		client.httpClientRaw.Transport = createTransport(nil)
	}

	// 设置日志格式为json格式
	client.log.SetFormatter(&logrus.JSONFormatter{PrettyPrint: true})
	client.log.SetOutput(os.Stdout)

	// 设置日志级别为DebugLevel
	client.log.SetLevel(logrus.DebugLevel)

	// 默认超时时间为 30 秒
	client.SetTimeout(30)
	// 默认重试次数为 3 次
	client.SetRetryCount(defaultRetryCount)
	// 默认 User-Agent 为随机生成的浏览器 User-Agent
	client.SetUserAgent(browser.Random())
	return client
}

// SetBaseURL 方法用于设置HTTP请求的 BaseUrl 部分。它接收一个 string 类型的参数，该参数表示 BaseUrl 的值。
func (client *Client) SetBaseURL(baseUrl string) *Client {
	client.baseUrl = strings.TrimRight(baseUrl, "/")
	return client
}

// SetContentType 方法用于设置 HTTP 请求的 ContentType 部分。它接收一个 string 类型的参数，该参数表示 ContentType 的值。
func (client *Client) SetContentType(contentType string) *Client {
	client.Header["Content-Type"] = contentType
	return client
}

// SetDebugFile 方法用于设置输出调试信息的文件。它接收一个 string 类型的参数，该参数表示文件名。
func (client *Client) SetDebugFile(name string) *Client {
	client.Debug = true
	if file, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err != nil {
		client.LogError(err, name, "client.go", "SetDebugFile")
	} else {
		client.log.SetOutput(file)
	}
	return client
}

// R 方法用于创建一个新的 Request 对象。它接收一个 string 类型的参数，该参数表示 HTTP 请求的 Path 部分。
func (client *Client) R() *Request {
	req := &Request{
		client:     client,
		URL:        &url.URL{},
		ctx:        context.Background(),
		Header:     sync.Map{},
		QueryParam: sync.Map{},
	}
	cookies := make([]*http.Cookie, 0)
	for i, cookie := range client.Cookies {
		// 创建一个新的cookie实例
		newCookie := new(http.Cookie)
		// 使用一个结构体赋值，复制cookie的值到新的实例
		*newCookie = *cookie
		// 将新的cookie指针放入新的切片中
		cookies[i] = newCookie
	}
	// 现在 cookies 切片包含了原始cookies的深拷贝
	req.Cookies = cookies // 将深拷贝的cookies设置到请求中

	// 设置 Header
	req.SetHeaders(client.Header)

	req.SetQueryParams(client.QueryParam)
	return req
}
func (client *Client) LogError(err any, query any, fileName, funcName string) {
	client.log.WithFields(logrus.Fields{
		"query": query,
		"func":  funcName,
		"file":  fileName,
	}).Error(err)

}

func (client *Client) LogInfo(err any, query any, funcName string) {
	client.log.WithFields(logrus.Fields{
		"query": query,
		"func":  funcName,
	}).Info(err)
}
func (client *Client) LogDebug(info string) {
	client.log.WithFields(logrus.Fields{}).Debug(info)
}

func (client *Client) LogFatal(err error, query any, fileName string, funcName string) {
	client.log.WithFields(logrus.Fields{
		"query": query,
		"func":  funcName,
		"file":  fileName,
	}).Fatal(err.Error())
}

// SetCookieString 方法用于设置 HTTP 请求的 Cookie 部分。它接收一个 string 类型的参数，该参数表示 Cookie 的值。
func (client *Client) SetCookieString(cookieStr string) *Client {
	// 按照分号拆分 Cookie 字符串
	parts := strings.Split(cookieStr, ";")
	for _, part := range parts {
		// 分割每个键值对
		keyValue := strings.SplitN(part, "=", 2)
		if len(keyValue) == 2 {
			// 修剪可能的空白字符并设置 Cookie
			name := strings.TrimSpace(keyValue[0])
			value := strings.TrimSpace(keyValue[1])
			client.SetCookie(&http.Cookie{Name: name, Value: value})
		}
	}
	return client
}
func (client *Client) SetCookie(cookie *http.Cookie) *Client {
	client.Cookies = append(client.Cookies, cookie)
	return client
}
func (client *Client) SetCookies(cookie []*http.Cookie) *Client {
	for _, c := range cookie {
		client.SetCookie(c)
	}
	return client
}

// SetCookieJar 方法用于设置 HTTP 请求的 CookieJar 部分。它接收一个 http.CookieJar 类型的参数，该参数表示 CookieJar 的值。
func (client *Client) SetCookieJar(cookieJar http.CookieJar) *Client {
	client.httpClientRaw.Jar = cookieJar
	return client
}

func (client *Client) SetResultFunc(f func(v string) (string, error)) *Client {
	client.setResultFunc = f
	return client
}

// SetDebug 方法用于设置是否输出调试信息,如果调用该方法，那么将输出调试信息。
func (client *Client) SetDebug() *Client {
	client.Debug = true
	return client
}

// SetRetryCount 方法用于设置重试次数。它接收一个 int 类型的参数，该参数表示重试次数。
func (client *Client) SetRetryCount(count int) *Client {
	if count <= 0 {
		client.LogInfo("retry number must be greater than 0", count, "SetRetryCount")
	} else {
		client.RetryCount = count
	}
	return client
}

// SetHeader 方法用于设置 HTTP 请求的 Header 部分。它接收两个 string 类型的参数，
func (client *Client) SetHeader(key string, value interface{}) *Client {
	client.Header[key] = fmt.Sprintf("%v", value)
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
	client.QueryParam[key] = value
	return client
}

// SetQueryParams 方法用于设置 HTTP 请求的 Query 部分。它接收一个 map[string]interface{} 类型的参数，
func (client *Client) SetQueryParams(params map[string]any) *Client {
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
		// 将 params 中的参数存储到 QueryParam 中
		for key, value := range params {
			client.SetQueryParam(key, value[0])
		}
	} else {
		client.LogError(err, query, "client.go", "SetQueryParamString")
	}
	return client
}

// SetProxy 方法用于设置 HTTP 请求的 Proxy 部分。它接收一个 string 类型的参数，该参数表示 Proxy 的值。
func (client *Client) SetProxy(proxy string) *Client {
	u, err := url.Parse(proxy)
	if err != nil {
		client.LogError(err, proxy, "client.go", "SetProxy")
		return client
	}
	// 设置 Transport 的 Proxy 字段
	client.httpClientRaw.Transport = &http.Transport{Proxy: http.ProxyURL(u)}
	return client
}

// SetTimeout 方法用于设置 HTTP 请求的 Timeout 部分, timeout 单位为秒。它接收一个 int 类型的参数，该参数表示 Timeout 的值。
func (client *Client) SetTimeout(timeout int) *Client {
	// 设置 httpClientRaw 的 Timeout 字段, timeout 单位为秒
	client.httpClientRaw.Timeout = time.Duration(timeout * int(time.Second))
	return client
}

// SetBasicAuth 方法用于设置 HTTP 请求的 BasicAuth 部分。它接收两个 string 类型的参数，分别表示用户名和密码。
func (client *Client) SetBasicAuth(username, password string) *Client {
	client.SetAuthorizationKey(client.AuthScheme + base64.StdEncoding.EncodeToString([]byte(username+":"+password)))
	return client
}

// SetAuthorizationKey 方法用于设置 HTTP 请求的 Authorization 部分。它接收一个 string 类型的参数，该参数表示 Authorization 的值。
func (client *Client) SetAuthorizationKey(authToken string) *Client {
	client.SetHeader(client.HeaderAuthorizationKey, authToken)
	return client
}
