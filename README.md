# Peekalo - Web Page Analyzer

This repository contains both the client and server components of a web page analyzer. It analyzes a given web page URL and returns structural and semantic details.

![image](https://github.com/user-attachments/assets/0df1fcdd-979d-41cf-9bfd-1cb45bf5f6a1)


The client is a barebones implementation using HTML, CSS, and JavaScript. The primary focus was on the server, written in Go. See the Server Documentation section for more details on endpoints and functionality.

## Running the Solution Locally

### Prerequisites
- Docker installed
- Go installed (if building the server from source)

### Steps

Navigate to the root directory where the docker-compose.yml is present.

Build and start the application using Docker Compose:

```bash
docker-compose up --build
```

- The client will be available at: http://localhost:5000

- The server will be running at: http://localhost:8080


## Server Documentation

### Features
- Analyze the HTML version of a web page
- Extract the page title
- Count headings
- Classify and count links:
  - Internal
  - External
  - Inaccessible
- Detect presence of login forms

### Running the Server
You can run the server either via Docker or directly using Go:

#### Using Docker:

```bash
cd web_analyzer_server
docker build -t peekalo-server .
docker run --rm -p 8080:8080 peekalo-server
```
#### Using Go:
```bash
cd web_analyzer_server
go mod tidy
go run .
```

### API Endpoints

**`POST /analyze`**
Analyzes the given web page URL.

```json
{
  "url": "https://example.com"
}
```

Response:
***200 OK***

```json
{
    "success": true,
    "data": {
        "html_version": "HTML 5",
        "title": "Sri Lanka - Wikipedia",
        "headings": {
            "h1": 1,
            "h2": 16,
            "h3": 28,
            "h4": 0,
            "h5": 0,
            "h6": 0
        },
        "link_stats": {
            "internal": 2240,
            "external": 1014,
            "inaccessible": 1002
        },
        "has_login": false
    }
}
```

***400 Bad Request***
For client side errors where validations fail, such as invalid payload.

Response:

```json
{
    "success": false,
    "error": "Validation failed: Key: 'UrlAnalyzeRequest.URL' Error:Field validation for 'URL' failed on the 'url' tag"
}
```

***500 Internal Server Error***

Errors when processing a validated request.

Response:

```json
{
    "success": false,
    "error": "Failed to analyze URL: failed to fetch URL: Get \"https://en.wikipedddias.org/wiki/Ssssris_Landdkaaa\": dial tcp: lookup en.wikipedddias.org on 127.0.0.11:53: no such host"
}

```
 **`GET /healthz`**

Simple health check to verify if the server is running.

Response Code: 200 OK

Response Body: Application is healthy!

**`GET /metrics`**

The GET /metrics endpoint exposes the following Prometheus counters, which instrument request handling and analysis behavior in the application:
### Prometheus Counters

| Metric Name                        | Description                                  |
|-----------------------------------|----------------------------------------------|
| `request_invalid_count`           | Number of invalid requests received          |
| `request_received_success_count`  | Number of successfully received requests     |
| `request_analyzer_success_count`  | Number of requests successfully analyzed     |
| `request_analyzer_failure_count`  | Number of requests that failed to be analyzed |


### Config
Currently set via the `config.go`. External configurations have not been specified yet.

### Third-party Packages
- `github.com/go-chi`: Middleware, enabling CORS and routing
- `github.com/go-playground/validator/v10`: Struct validation
- `github.com/prometheus/client_golang`: Prometheus counters
- `github.com/rs/zerolog`: Structured logging
- `golang.org/x/net`: HTML parsing
- `github.com/stretchr/testify`: Support for unit testing - assertions, mocking.


Future Improvements
- Implement response caching with configurable Time-To-Live (TTL) values.
- Enhance overall configurability for greater flexibility. (context deadlines for http client etc.)
- Offload heavy processing to a separate worker via Kafka to ensure the API can handle larger request volumes efficiently.
