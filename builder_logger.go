package builder

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net/http"
)

// indentJson 方法用于格式化 JSON 字符串成为 map[string]*json.RawMessage 类型。
func indentJson(a string) (map[string]*json.RawMessage, error) {
	var objmap map[string]*json.RawMessage
	err := json.Unmarshal([]byte(a), &objmap)
	if err != nil {
		return nil, err
	}
	return objmap, nil
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
		if request.bodyBytes != nil {
			body = string(request.bodyBytes)
		} else {
			body = "this request has no body"
		}
	}
	fields := logrus.Fields{
		"Method":  request.GetMethod(),
		"Host":    request.GetHost(),
		"Path":    request.GetPath(),
		"HEADERS": request.GetRequestHeader(),
		"BODY":    body,
	}
	if request.Cookies != nil {
		fields["Cookie"] = request.Cookies
	} else {
		fields["Cookie"] = "this request has no cookies"
	}
	return fields
}

// newFormatResponseLogText 方法用于格式化 HTTP 响应的日志信息。
func newFormatResponseLogText(response *Response) logrus.Fields {
	fields := logrus.Fields{
		"Code":   response.GetStatusCode(),
		"Status": response.GetStatus(),
		"Proto":  response.GetProto(),
	}
	if header := response.GetHeader(); header != nil {
		if cookies := header.Get("Set-Cookie"); cookies != "" {
			fields["Cookie"] = cookies
		} else {
			fields["Cookie"] = "this response has no cookies"
		}
		fields["Header"] = header
	}
	result := response.String()
	if objmap, err := indentJson(result); err != nil {
		fields["Result"] = result
	} else {
		fields["Result"] = objmap
	}
	return fields
}
