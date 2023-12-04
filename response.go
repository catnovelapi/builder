package builder

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/tidwall/gjson"
	"golang.org/x/net/html"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"log"
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
)

type Response struct {
	Result        string         // 响应体字符串结果
	ResponseRaw   *http.Response // 指向 http.Response 的指针
	RequestSource *Request       // 指向 Request 的指针
}

// newParseUrl 方法用于解析 URL。它接收一个 string 类型的参数，该参数表示 HTTP 请求的 Path 部分。
func (request *Request) newParseUrl(path string) (*url.URL, error) {
	request.client.Lock()
	defer request.client.Unlock()
	// 如果 baseUrl 不为空，且 path 不是以 / 开头，则在 path 前加上 /
	if request.client.GetClientBaseURL() == "" && path == "" {
		return nil, fmt.Errorf("request Error: %s", "baseUrl is empty")
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
	// 设置 URL, Host
	request.RequestRaw.URL = u
	request.RequestRaw.Host = u.Host
	return u, nil
}

// newResponse 方法用于创建一个 Response 对象。它接收两个 string 类型的参数，分别表示 HTTP 请求的方法和路径。
func (request *Request) newResponse(method, path string) (*Response, error) {
	_, err := request.newParseUrl(path)
	if err != nil {
		return nil, err
	}
	if request.RequestRaw.Method = method; request.RequestRaw.Method == MethodGet {
		// GET请求不需要设置Body,因为Body会被忽略
		request.RequestRaw.URL.RawQuery = request.GetQueryParamsEncode()
	} else {
		if len(request.queryParams) > 0 {
			request.RequestRaw.Body = request.GetQueryParamsNopCloser()
		}
	}
	if request.client.GetClientDebug() {
		request.client.debugLoggers.formatRequestLogText(request)
	}
	if request.client.GetClientRetryNumber() == 0 {
		request.client.SetRetryCount(1)
	}
	for i := 0; i < request.client.GetClientRetryNumber(); i++ {
		response, ok := request.newDoResponse()
		if ok != nil {
			log.Println(fmt.Sprintf("%s Error: %s Retry:%v", request.RequestRaw.Method, ok.Error(), i))
			continue
		}
		if request.client.GetClientDebug() {
			request.client.debugLoggers.formatResponseLogText(response)
		}
		return response, nil

	}
	return nil, fmt.Errorf("request Error: %s", err.Error())
}

// newDoResponse 方法用于执行 HTTP 请求。它接收一个 Response 对象的指针，表示 HTTP 请求的响应。
func (request *Request) newDoResponse() (*Response, error) {
	responseRaw, err := request.client.clientRaw.Do(request.RequestRaw)
	if err != nil {
		return nil, err
	}
	return &Response{RequestSource: request, ResponseRaw: responseRaw}, nil
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

// GetStatusCode 方法用于获取 HTTP 响应的状态码。
func (response *Response) GetStatusCode() int {
	return response.ResponseRaw.StatusCode
}

func (response *Response) IsStatusOk() bool {
	return response.ResponseRaw.StatusCode == 200
}

// GetStatus 方法用于获取 HTTP 响应的状态。
func (response *Response) GetStatus() string {
	return response.ResponseRaw.Status
}

// GetByte 方法用于获取 HTTP 响应的字节结果。
func (response *Response) GetByte() []byte {
	// 如果响应体为空，直接返回 nil
	if response.ResponseRaw.Body == nil {
		return nil
	}
	// 如果响应体不为空，且 Result 不为空，则直接返回 Result
	if response.Result != "" {
		return []byte(response.Result)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("GetByte Body.Close Error:", err)
		}
	}(response.ResponseRaw.Body)
	body, ok := io.ReadAll(response.ResponseRaw.Body)
	if ok != nil {
		return nil
	}
	response.Result = string(body)
	return body
}

// String 方法用于获取 HTTP 响应的字符串结果。
func (response *Response) String() string {
	return string(response.GetByte())
}

// Json 方法用于将 HTTP 响应的字符串结果解析为 JSON 对象。它接收一个 interface{} 类型的参数，该参数必须是指针类型。
func (response *Response) Json(v any) error {
	valueType := reflect.TypeOf(v)
	if valueType.Kind() != reflect.Ptr {
		return fmt.Errorf("DecodeJson:传入的对象必须是指针类型")
	}
	return json.NewDecoder(strings.NewReader(response.String())).Decode(v)
}

// StringGbk 方法用于将 HTTP 响应的字符串结果解码为 GBK 编码的字符串。
func (response *Response) StringGbk() string {
	decoder := simplifiedchinese.GBK.NewDecoder()
	utf8BodyReader := transform.NewReader(strings.NewReader(response.String()), decoder)
	utf8Body, err := io.ReadAll(utf8BodyReader)
	if err != nil {
		fmt.Println("解码失败:", err)
		return ""
	}
	return string(utf8Body)
}

// Html 方法用于将 HTTP 响应的字符串结果解析为 HTML 文档。
func (response *Response) Html() *goquery.Document {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(response.String()))
	if err != nil {
		log.Printf("读取响应体失败: %s", err)
		return nil
	}
	return doc
}

// HtmlGbk 方法用于将 HTTP 响应的字符串结果解析为 GBK 编码的 HTML 文档。
func (response *Response) HtmlGbk() *goquery.Document {
	docs, err := html.Parse(strings.NewReader(response.StringGbk()))
	if err != nil {
		fmt.Println("解析HTML失败:", err)
		return nil
	}
	doc := goquery.NewDocumentFromNode(docs)
	if err != nil {
		fmt.Println("解析HTML失败:", err)
		return nil
	}
	return doc
}

// Gjson 方法用于将 HTTP 响应的字符串结果解析为 gjson.Result 对象。
func (response *Response) Gjson() gjson.Result {
	return gjson.Parse(response.String())
}

// GetHeader 方法用于获取 HTTP 响应的 Header 部分。
func (response *Response) GetHeader() http.Header {
	return response.ResponseRaw.Header
}

// GetCookies 方法用于获取 HTTP 响应的 Cookies 部分。
func (response *Response) GetCookies() []*http.Cookie {
	return response.ResponseRaw.Cookies()
}
