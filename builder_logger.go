package builder

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"log"
	"net/http"
	"os"
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

type LoggerClient struct {
	loggerArray []string
	debugFile   *os.File
}

// NewLoggerClient 方法用于创建一个 LoggerClient 对象, 它接收一个 *os.File 类型的参数，该参数表示日志文件。
func NewLoggerClient(debugFile *os.File) *LoggerClient {
	return &LoggerClient{debugFile: debugFile}
}

// formatRequestLogText 方法用于格式化 HTTP 请求的日志信息。
func (builderLogger *LoggerClient) formatRequestLogText(debug bool, request *Request) {
	var body string
	if request.GetQueryParams() != nil {
		body = request.GetQueryParamsEncode()
	}
	if body == "" {
		if request.RequestRaw.Body != nil {
			body = fmt.Sprintf("%v", request.RequestRaw.Body)
		}
	}
	formatText := formatLog("\n==============================================================================\n"+
		"~~~ REQUEST ~~~\n"+
		"%s %s %s\n"+
		"PATH   : %v\n"+
		"HEADERS:\n%s\n"+
		"Cookies:\n%v\n"+
		"BODY   :\n%v\n"+
		"------------------------------------------------------------------------------\n",
		request.GetMethod(),
		request.GetHost(),
		request.GetProto(),
		request.GetPath(),
		composeHeaders(copyHeaders(request.GetRequestHeader())),
		request.GetRequestHeader().Get("Cookies"),
		body,
	)
	if debug {
		fmt.Println(formatText)
	}
	if builderLogger.debugFile != nil {
		if _, err := builderLogger.debugFile.WriteString(formatText); err != nil {
			builderLogger.loggerArray = append(builderLogger.loggerArray, err.Error())
		}
	}
}

// formatResponseLogText 方法用于格式化 HTTP 响应的日志信息。
func (builderLogger *LoggerClient) formatResponseLogText(debug bool, response *Response) string {
	var repLogText string
	if cookies := response.GetCookies(); cookies != nil {
		repLogText += "  Cookies:\n"
		for _, cookie := range response.GetCookies() {
			repLogText += fmt.Sprintf("    %s=%s", cookie.Name, cookie.Value)
		}
	}
	formatText := formatLog("\n\n"+
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
	if debug {
		fmt.Println(formatText)
	}
	if builderLogger.debugFile != nil {
		if _, err := builderLogger.debugFile.WriteString(formatText); err != nil {
			builderLogger.loggerArray = append(builderLogger.loggerArray, err.Error())
		}
	}
	return repLogText
}
func (builderLogger *LoggerClient) ReturnLog() []string {
	return builderLogger.loggerArray
}

//func (builderLogger *LoggerClient) ResponseResultMiddleware(f func(result string) string) {
//	f(builderLogger.response.String())
//}

// formatLog 方法用于格式化日志信息。它接收一个 string 类型的参数，该参数表示日志信息的格式，以及一个 interface{} 类型的可变参数，该参数表示日志信息的参数。
func formatLog(format string, params ...interface{}) string {
	return fmt.Sprintf(format, params...)
}
