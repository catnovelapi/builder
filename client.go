package builder

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/EDDYCJY/fake-useragent"
	"golang.org/x/net/context"
	"golang.org/x/net/publicsuffix"
	"log"
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
	sync.RWMutex                          // 用于保证线程安全
	MaxConcurrent          chan struct{}  // 用于限制并发数
	timeout                int            // timeout 用于存储 HTTP 请求的 Timeout 部分
	baseUrl                string         // baseUrl 用于存储 HTTP 请求的 BaseUrl 部分
	debug                  bool           // debug 用于存储是否输出调试信息
	debugLoggers           *LoggerClient  // debugLoggers 用于存储调试信息的文件
	httpClientRaw          *http.Client   // httpClientRaw 用于存储 http.Client 的指针
	Header                 http.Header    // Header 用于存储 HTTP 请求的 Header 部分
	QueryParam             map[string]any // QueryParam 用于存储 HTTP 请求的 Query 部分
	setResultFunc          func(v string) (string, error)
	FormData               map[string]any
	Token                  string
	AuthScheme             string
	Cookies                []*http.Cookie
	Debug                  bool
	DisableWarn            bool
	AllowGetMethodPayload  bool
	RetryCount             int
	RetryWaitTime          time.Duration
	JSONMarshal            func(v interface{}) ([]byte, error)
	JSONUnmarshal          func(data []byte, v interface{}) error
	XMLMarshal             func(v interface{}) ([]byte, error)
	XMLUnmarshal           func(data []byte, v interface{}) error
	HeaderAuthorizationKey string
	body                   interface{} // body 用于存储 HTTP 请求的 Body 部分
}

const (
	defaultRetryCount = 3
	defaultWaitTime   = 100
)

// NewClient 方法用于创建一个新的 Client 对象, 并返回该对象的指针。
func NewClient() *Client {
	cookieJar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	client := &Client{
		MaxConcurrent:          make(chan struct{}, 500), // 用于限制并发数, 最大并发数为 500
		QueryParam:             map[string]any{},         // 初始化 QueryParam
		Header:                 make(http.Header),        // 初始化 Header
		FormData:               map[string]any{},
		Cookies:                make([]*http.Cookie, 0),
		RetryWaitTime:          defaultWaitTime,
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
	client.Header.Set("Content-Type", contentType)
	return client
}

// SetDebugFile 方法用于设置输出调试信息的文件。它接收一个 string 类型的参数，该参数表示文件名。
func (client *Client) SetDebugFile(name string) *Client {
	if fileInfo, err := os.Stat(name + ".txt"); err != nil {
		if !os.IsNotExist(err) {
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
		//client.debugFile = file
		client.debugLoggers = NewLoggerClient(file)
	}
	return client
}

// R 方法用于创建一个新的 Request 对象。它接收一个 string 类型的参数，该参数表示 HTTP 请求的 Path 部分。
func (client *Client) R() *Request {
	req := &Request{
		client: client,
		RequestRaw: &http.Request{
			//Body:       rc,
			//URL:        u,
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header:     make(http.Header),
		},
		FormData:   sync.Map{},
		QueryParam: sync.Map{},
		Cookies:    client.Cookies,
	}
	// Create a new header and copy from the original
	for key, values := range client.GetClientHeaders() {
		for _, value := range values {
			req.SetHeader(key, value)
		}
	}

	if client.FormData != nil && len(client.FormData) > 0 {
		req.SetFormDataMany(client.FormData)
		req.SetHeader("Content-Type", "application/x-www-form-urlencoded")
	}
	req.SetQueryParams(client.QueryParam)
	req.RequestRaw.WithContext(context.Background())
	return req
}

// SetCookie 方法用于设置 HTTP 请求的 Cookie 部分。它接收一个 string 类型的参数，该参数表示 Cookie 的值。
func (client *Client) SetCookie(cookie string) *Client {
	if client.Header.Get("Cookie") != "" {
		// 如果已经设置了 Cookie，那么将新的 Cookie 追加到原有的 Cookie 后面
		client.Header.Set("Cookie", client.Header.Get("Cookie")+";"+cookie)
	} else {
		client.Header.Set("Cookie", cookie)
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
	client.debug = true
	return client
}

// SetRetryCount 方法用于设置重试次数。它接收一个 int 类型的参数，该参数表示重试次数。
func (client *Client) SetRetryCount(retryCount int) *Client {
	if retryCount <= 0 {
		log.Println("retry number must be greater than 0")
	} else {
		client.RetryCount = retryCount
	}
	return client
}

// SetHeader 方法用于设置 HTTP 请求的 Header 部分。它接收两个 string 类型的参数，
func (client *Client) SetHeader(key string, value interface{}) *Client {
	// 将 value 转换为 string 类型, 并将其存储到 Header 中
	client.Header.Set(key, fmt.Sprintf("%v", value))
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
	client.QueryParam[key] = fmt.Sprintf("%v", value)
	return client
}

// SetQueryParams 方法用于设置 HTTP 请求的 Query 部分。它接收一个 map[string]interface{} 类型的参数，
func (client *Client) SetQueryParams(params map[string]any) *Client {
	for key, value := range params {
		client.SetQueryParam(key, value)
	}
	return client
}

// SetFormDataMany 方法用于设置 HTTP 请求的 Query 部分。它接收一个 url.Values 类型的参数，
func (client *Client) SetFormDataMany(params url.Values) *Client {
	for key, value := range params {
		client.SetFormData(key, value[0])
	}
	return client
}
func (client *Client) SetFormData(key string, value any) *Client {
	client.FormData[key] = value
	return client
}

// SetQueryParamString 方法用于设置 HTTP 请求的 Query 部分。它接收一个 string 类型的参数，
func (client *Client) SetQueryParamString(query string) *Client {
	// 将 query 解析为 url.Values 类型的参数
	params, err := url.ParseQuery(strings.TrimSpace(query))
	if err == nil {
		// 将 params 中的参数存储到 QueryParam 中
		client.SetFormDataMany(params)
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
