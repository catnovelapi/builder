package builder

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"log"
	"net/http"
	"sort"
	"strings"
)

// indentJson 方法用于格式化 JSON 字符串。它接收一个 string 类型的参数，该参数表示 JSON 字符串。
func indentJson(a string) string {
	// 判断是否为 JSON 字符串, 如果不是则直接返回
	if !gjson.Valid(a) {
		return a
	}
	var objmap map[string]*json.RawMessage
	err := json.Unmarshal([]byte(a), &objmap)
	if err != nil {
		log.Println("indentJson:解析 JSON 字符串失败")
		return a + "\n" + err.Error()
	}
	formatted, err := json.MarshalIndent(objmap, "", "  ")
	if err != nil {
		return a + "\n" + err.Error()
	}
	return string(formatted)
}

// copyHeaders 方法用于复制 HTTP 请求的 Header 部分。它接收一个 http.Header 类型的参数，该参数表示 HTTP 请求的 Header 部分。
func copyHeaders(hdrs http.Header) http.Header {
	nh := http.Header{}
	if hdrs != nil {
		for k, v := range hdrs {
			nh[k] = v
		}
	}
	return nh
}

// sortHeaderKeys 方法用于对 HTTP 请求的 Header 部分的 key 进行排序。它接收一个 http.Header 类型的参数，该参数表示 HTTP 请求的 Header 部分。
func sortHeaderKeys(hdrs http.Header) []string {
	keys := make([]string, 0, len(hdrs))
	for key := range hdrs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// composeHeaders 方法用于组合 HTTP 请求的 Header 部分。它接收一个 http.Header 类型的参数，该参数表示 HTTP 请求的 Header 部分。
func composeHeaders(hdrs http.Header) string {
	str := make([]string, 0, len(hdrs))
	for _, k := range sortHeaderKeys(hdrs) {
		str = append(str, "\t"+strings.TrimSpace(fmt.Sprintf("%25s: %s", k, strings.Join(hdrs[k], ", "))))
	}
	return strings.Join(str, "\n")
}

type BuilderLoggerClient struct {
	request       *Request
	response      *Response
	formatLogText string
}

// NewLogger 方法用于创建一个 builderLoggerClient 对象。它接收一个 *Response 类型的参数，该参数表示 HTTP 响应。
func NewLogger(rep *Response) *BuilderLoggerClient {
	return &BuilderLoggerClient{request: rep.RequestSource, response: rep}
}

// CreateLogInfo 方法用于创建日志信息。
func (builderLogger *BuilderLoggerClient) CreateLogInfo() string {
	return builderLogger.formatRequestLogText() + builderLogger.formatResponseLogText()
}

// formatRequestLogText 方法用于格式化 HTTP 请求的日志信息。
func (builderLogger *BuilderLoggerClient) formatRequestLogText() string {
	var reqLogText string
	var body string
	if builderLogger.request.GetQueryParams() != nil {
		body = builderLogger.request.GetQueryParamsEncode()
	}
	if body == "" {
		if builderLogger.request.RequestRaw.Body != nil {
			body = fmt.Sprintf("%v", builderLogger.request.RequestRaw.Body)
		}
	}
	reqLogText = formatLog("\n==============================================================================\n"+
		"~~~ REQUEST ~~~\n"+
		"%s %s %s\n"+
		"PATH   : %v\n"+
		"HEADERS:\n%s\n"+
		"BODY   :\n%v\n"+
		"------------------------------------------------------------------------------\n",
		builderLogger.request.GetMethod(),
		builderLogger.request.GetHost(),
		builderLogger.request.GetProto(),
		builderLogger.request.GetPath(),
		composeHeaders(copyHeaders(builderLogger.request.GetRequestHeader())),
		body,
	)

	return reqLogText
}

// formatResponseLogText 方法用于格式化 HTTP 响应的日志信息。
func (builderLogger *BuilderLoggerClient) formatResponseLogText() string {
	var repLogText string
	if cookies := builderLogger.response.GetCookies(); cookies != nil {
		repLogText += "  Cookies:\n"
		for _, cookie := range builderLogger.response.GetCookies() {
			repLogText += fmt.Sprintf("    %s=%s", cookie.Name, cookie.Value)
		}
	}
	repLogText += formatLog("\n\n"+
		"~~~ RESPONSE ~~~\n"+
		"Code   : %v\n"+
		"Status : %s\n"+
		"HEADERS:\n%s\n"+
		"BODY   :\n%v\n"+
		"------------------------------------------------------------------------------\n",
		builderLogger.response.GetStatusCode(),
		builderLogger.response.GetStatus(),
		composeHeaders(builderLogger.response.GetHeader()),
		indentJson(builderLogger.response.String()))
	return repLogText
}

// handleCookies 方法用于处理 Cookie。它接收一个 http.Header 类型的参数，该参数表示 HTTP 响应的 Header 部分。
func (builderLogger *BuilderLoggerClient) handleCookies(rh http.Header) string {
	var cookieText string
	if builderLogger.request.RequestRaw.Cookies() != nil {
		for _, cookie := range builderLogger.request.RequestRaw.Cookies() {
			cookieText += fmt.Sprintf("%s=%s\n", cookie.Name, cookie.Value)
		}
	} else {
		if rh.Get("Cookie") != "" {
			for _, cookie := range rh["Cookie"] {
				cookieText += fmt.Sprintf("%s\n", cookie)
			}
		}
	}
	return strings.TrimSpace(cookieText)
}

// formatLog 方法用于格式化日志信息。它接收一个 string 类型的参数，该参数表示日志信息的格式，以及一个 interface{} 类型的可变参数，该参数表示日志信息的参数。
func formatLog(format string, params ...interface{}) string {
	return fmt.Sprintf(format, params...)
}
