Here is a generated GitHub README.md based on the provided Go code:

# Go HTTP Client Builder

Go HTTP Client Builder is a Go package for building HTTP clients with fluent interface. It provides an easy way to configure HTTP requests and handle responses.

## Installation

```
go get github.com/catnovel/builder
```

## Usage

### Create a new client

```go
client := builder.NewClient()
```

### Configure the client

```go 
client.SetBaseUrl("https://api.example.com")
       .SetTimeout(30)
       .SetUserAgent("my-app/1.0")
```

### Make a request

```go
req := client.R()
resp, err := req.Get("/users")
```

### Handle response

```go
if err != nil {
  // handle error
}

fmt.Println(resp.GetStatusCode())
fmt.Println(resp.String())
```

### Post form data

```go
req.SetFormData(map[string]string{
  "name": "John",
  "email": "john@email.com", 
})

resp, err := req.Post("/users")
```

## Request builder

The `R()` method creates a new `Request` object that can be used to configure an HTTP request.

### Set request body

```go
req.SetBody(map[string]interface{}{
  "name": "John",
  "age": 30,
})
```

Supports body types: `string`, `[]byte`, `map[string]interface{}`, etc.

### Set request headers

```go
req.SetHeader("Content-Type", "application/json")
   .SetHeader("Authorization", "Bearer token")
```

### Set query parameters

```go 
req.SetQueryParam("sort", "desc")
   .SetQueryParam("limit", 100)
```

### Set cookies

```go
req.SetCookies([]*http.Cookie{
  {Name: "session", Value: "1234"},
})
```

## Response builder

The `Response` object contains the HTTP response and useful methods to handle it.

### Get response as string

```go
resp.String()
``` 

### Get response as JSON

```go
var data struct{}
resp.Json(&data) 
```

### Get response as HTML document

```go 
doc := resp.Html()
```

### Get response headers

```go
headers := resp.GetHeader()
```

## Full Example

```go
client := builder.NewClient()
req := client.R()

req.SetQueryParam("page", 1)
   .SetHeader("Authorization", "token")

resp, err := req.Get("/users")

if err != nil {
  // handle error 
}

var result []User
resp.Json(&result)
```

## Contributing

Pull requests are welcome. Feel free to open an issue for any bugs or feature requests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.