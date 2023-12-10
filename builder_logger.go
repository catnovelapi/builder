package builder

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
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

// formatRequestLogText 方法用于格式化 HTTP 请求的日志信息。
func formatRequestLogText(request *Request) string {
	var body string
	if body = request.GetQueryParamsEncode(); body == "" {
		if body = request.GetFormDataEncode(); body == "" {
			if request.bodyBytes != nil {
				body = string(request.bodyBytes)
			} else {
				body = "this request has no body"
			}
		}
	}
	return formatLog("\n==============================================================================\n"+
		"~~~ REQUEST ~~~\n"+
		"%s %s %s\n"+
		"PATH   : %v\n"+
		"HEADERS:\n%s\n"+
		"Cookies:\n%v\n"+
		"BODY   :\n%v\n"+
		"------------------------------------------------------------------------------\n",
		request.GetMethod(),
		request.GetHost(),
		"request.GetProto()",
		request.GetPath(),
		composeHeaders(copyHeaders(request.GetRequestHeader())),
		request.GetRequestHeader().Get("Cookies"),
		body,
	)
}

// newFormatRequestLogText 方法用于格式化 HTTP 请求的日志信息。
func newFormatRequestLogText(request *Request) logrus.Fields {
	var body string
	if body = request.GetQueryParamsEncode(); body == "" {
		if body = request.GetFormDataEncode(); body == "" {
			if request.bodyBytes != nil {
				body = string(request.bodyBytes)
			} else {
				body = "this request has no body"
			}
		}
	}
	h := request.GetRequestHeader()
	fields := logrus.Fields{
		"Method":  request.GetMethod(),
		"Host":    request.GetHost(),
		"Path":    request.GetPath(),
		"HEADERS": h,
		"Cookies": h.Get("Cookies"),
		"BODY":    body,
	}
	return fields
}

// formatResponseLogText 方法用于格式化 HTTP 响应的日志信息。
func formatResponseLogText(response *Response) string {
	var repLogText string
	if cookies := response.GetCookies(); cookies != nil {
		repLogText += "  Cookies:\n"
		for _, cookie := range response.GetCookies() {
			repLogText += fmt.Sprintf("    %s=%s", cookie.Name, cookie.Value)
		}
	}
	return formatLog("\n\n"+
		"~~~ RESPONSE ~~~\n"+
		"Code   : %v\n"+
		"Status : %s\n"+
		"HEADERS:\n%s\n"+
		"BODY   :\n%v\n"+
		"------------------------------------------------------------------------------\n",
		response.GetStatusCode(),
		response.GetStatus(),
		composeHeaders(response.GetHeader()),
		indentJson(response.String()))
}

// newFormatResponseLogText 方法用于格式化 HTTP 响应的日志信息。
func newFormatResponseLogText(response *Response) logrus.Fields {
	h := response.GetHeader()
	fields := logrus.Fields{
		"Code":   response.GetStatusCode(),
		"Status": response.GetStatus(),
		"Header": h,
		"Result": response.String(),
	}
	if cookies := h.Get("Cookies"); cookies != "" {
		fields["Cookies"] = cookies
	} else {
		fields["Cookies"] = "this response has no cookies"
	}
	return fields
}

// formatLog 方法用于格式化日志信息。它接收一个 string 类型的参数，该参数表示日志信息的格式，以及一个 interface{} 类型的可变参数，该参数表示日志信息的参数。
func formatLog(format string, params ...interface{}) string {
	return fmt.Sprintf(format, params...)
}
