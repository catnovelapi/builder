package builder

// GetClientQueryParams 方法用于获取 HTTP 请求的 Query 部分。它返回一个
func (client *Client) GetClientQueryParams() map[string]any {
	return client.QueryParam
}

// GetClientBody 方法用于获取 HTTP 请求的 Body 部分。它返回一个 interface{} 类型的参数。
func (client *Client) GetClientBody() interface{} {
	return client.body
}

// GetClientBaseURL 方法用于获取 HTTP 请求的 BaseUrl 部分。它返回一个 string 类型的参数。
func (client *Client) GetClientBaseURL() string {
	return client.baseUrl
}

// GetClientDebug 方法用于获取 HTTP 请求的 Debug 部分。它返回一个 bool 类型的参数。
func (client *Client) GetClientDebug() bool {
	return client.Debug
}

// GetClientRetryNumber 方法用于获取 HTTP 请求的 RetryNumber 部分。它返回一个 int 类型的参数。
func (client *Client) GetClientRetryNumber() int {
	return client.RetryCount
}

// GetClientTimeout 方法用于获取 HTTP 请求的 Timeout 部分。它返回一个 int 类型的参数。
func (client *Client) GetClientTimeout() int {
	return client.timeout
}

// GetClientCookie 方法用于获取 HTTP 请求的 Cookie 部分。它返回一个 string 类型的参数。
func (client *Client) GetClientCookie() string {
	return client.Header["Cookie"]
}
