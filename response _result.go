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
	"net/http"
	"reflect"
	"strings"
)

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
func (response *Response) GetProto() string {
	return response.ResponseRaw.Proto
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
			response.RequestSource.client.LogError(err, "", "response.go", "GetByte")
		}
	}(response.ResponseRaw.Body)
	body, ok := io.ReadAll(response.ResponseRaw.Body)
	if ok != nil {
		return nil
	}
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
		response.RequestSource.client.LogError(err, "", "response.go", "StringGbk")
		return ""
	}
	return string(utf8Body)
}

// Html 方法用于将 HTTP 响应的字符串结果解析为 HTML 文档。
func (response *Response) Html() *goquery.Document {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(response.String()))
	if err != nil {
		response.RequestSource.client.LogError(err, "", "response.go", "Html")
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
		response.RequestSource.client.LogError(err, "", "response.go", "HtmlGbk")
		//fmt.Println("解析HTML失败:", err)
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
func (response *Response) GetCookieString() string {
	return response.ResponseRaw.Header.Get("Set-Cookie")
}
