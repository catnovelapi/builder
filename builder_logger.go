package builder

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"log"
	"net/http"
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
func header2Map(header http.Header) map[string]string {
	h := make(map[string]string)
	for k, v := range header {
		h[k] = v[0]
	}
	return h
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
	h := header2Map(request.GetRequestHeader())
	fields := logrus.Fields{
		"Method":  request.GetMethod(),
		"Host":    request.GetHost(),
		"Path":    request.GetPath(),
		"HEADERS": h,
		"BODY":    body,
	}
	if cookies := h["Cookie"]; cookies != "" {
		fields["Cookie"] = cookies
	} else {
		fields["Cookie"] = "this request has no cookies"
	}
	return fields
}

// newFormatResponseLogText 方法用于格式化 HTTP 响应的日志信息。
func newFormatResponseLogText(response *Response) logrus.Fields {
	h := header2Map(response.GetHeader())
	fields := logrus.Fields{
		"Code":   response.GetStatusCode(),
		"Status": response.GetStatus(),
		"Header": h,
		"Result": response.String(),
	}
	if cookies := h["Cookie"]; cookies != "" {
		fields["Cookie"] = cookies
	} else {
		fields["Cookie"] = "this response has no cookies"
	}
	return fields
}
