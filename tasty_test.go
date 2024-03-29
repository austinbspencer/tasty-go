package tasty //nolint:testpackage // testing private field

import (
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	mux       *http.ServeMux
	server    *httptest.Server
	client    *Client
	testToken = "fake-access-token+C"
)

func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
	client = NewClient(http.DefaultClient)
	client.Session = Session{
		SessionToken: &testToken,
	}
	client.baseURL = server.URL
	// Required for customRequest method
	client.baseHost = strings.Split(server.URL, "/")[2]
}

func teardown() {
	server.Close()
}

func TestTastyCertSession(t *testing.T) {
	c := NewCertClient(nil)

	require.NotNil(t, c.httpClient)
	require.Equal(t, apiCertBaseURL, c.baseURL)
	require.Equal(t, apiCertBaseHost, c.baseHost)
	require.Equal(t, streamerCertBaseURL, c.websocket)
	require.Equal(t, streamerCertBaseURL, c.GetWebsocketURL())

	cWithHTTP := NewCertClient(&http.Client{Timeout: time.Duration(30) * time.Second})

	require.NotNil(t, cWithHTTP.httpClient)
	require.Equal(t, time.Duration(30)*time.Second, cWithHTTP.httpClient.Timeout)
}

func TestTastySession(t *testing.T) {
	c := NewClient(nil)

	require.NotNil(t, c.httpClient)
	require.Equal(t, apiBaseURL, c.baseURL)
	require.Equal(t, apiBaseHost, c.baseHost)
	require.Equal(t, streamerBaseURL, c.websocket)
	require.Equal(t, streamerBaseURL, c.GetWebsocketURL())

	cWithHTTP := NewClient(&http.Client{Timeout: time.Duration(30) * time.Second})

	require.NotNil(t, cWithHTTP.httpClient)
	require.Equal(t, time.Duration(30)*time.Second, cWithHTTP.httpClient.Timeout)
}

func TestDecodeError(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, apiBaseURL+"/customers/me/accounts", nil)
	require.NoError(t, err)

	r.Header.Add("Content-Type", "application/json")
	httpClient := http.Client{Timeout: time.Duration(30) * time.Second}
	resp, err := httpClient.Do(r)
	require.NoError(t, err)

	defer resp.Body.Close()

	require.True(t, containsInt(errorStatusCodes, resp.StatusCode))
	tastyErr := decodeError(resp)
	require.NotNil(t, tastyErr)

	require.Equal(t, 401, tastyErr.StatusCode)
	require.Equal(t, "token_invalid", tastyErr.Code)
	require.Equal(t, "This token is invalid or has expired", tastyErr.Message)
	require.Empty(t, tastyErr.Errors)

	require.Equal(
		t,
		"\nError in request 401;\nCode: token_invalid\nMessage: This token is invalid or has expired",
		tastyErr.Error(),
	)
}

func TestCustomRequest(t *testing.T) {
	c := NewCertClient(&http.Client{Timeout: time.Duration(30) * time.Second})
	c.Session.SessionToken = &testToken

	// Test invalid payload
	invalid := math.Inf(1)
	httpResp, tastyError := c.customRequest(http.MethodGet, "/test", nil, invalid, nil)
	require.NotNil(t, tastyError)
	require.Nil(t, httpResp, "payload error")

	require.Equal(t,
		"\nError in request 0;\nCode: \nMessage: Client Side Error: json: unsupported value: +Inf",
		tastyError.Error())

	// Test invalid query
	httpResp, tastyError = c.customRequest(http.MethodGet, "/test", invalid, nil, nil)
	require.NotNil(t, tastyError)
	require.Nil(t, httpResp, "invalid query")

	require.Equal(t,
		"\nError in request 0;\nCode: \nMessage: Client Side Error: <nil>",
		tastyError.Error())

	// Test invalid method
	httpResp, tastyError = c.customRequest(http.MethodGet+"/sdfl/", "/test", nil, nil, nil)
	require.NotNil(t, tastyError)
	require.Nil(t, httpResp, "invalid method")

	require.Equal(t,
		"\nError in request 0;\nCode: \nMessage: Client Side Error: Get/sdfl/ \"https://api.cert.tastyworks.com/test\": net/http: invalid method \"GET/sdfl/\"",
		tastyError.Error())
}

func TestRequest(t *testing.T) {
	c := NewCertClient(&http.Client{Timeout: time.Duration(30) * time.Second})

	httpResp, tastyError := c.request(http.MethodGet, "/no-auth", nil, nil, nil)
	require.NotNil(t, tastyError)
	require.Nil(t, httpResp)

	require.Equal(t,
		"\nError in request 0;\nCode: invalid_session\nMessage: Session is invalid: Session Token cannot be nil.",
		tastyError.Error())

	c.Session.SessionToken = &testToken

	// Test invalid payload
	invalid := math.Inf(1)
	httpResp, tastyError = c.request(http.MethodGet, "/test", nil, invalid, nil)
	require.NotNil(t, tastyError)
	require.Nil(t, httpResp)

	require.Equal(t,
		"\nError in request 0;\nCode: \nMessage: Client Side Error: json: unsupported value: +Inf",
		tastyError.Error())

	// Test invalid query
	httpResp, tastyError = c.request(http.MethodGet, "/test", invalid, nil, nil)
	require.NotNil(t, tastyError)
	require.Nil(t, httpResp)

	require.Equal(t,
		"\nError in request 0;\nCode: \nMessage: Client Side Error: <nil>",
		tastyError.Error())

	// Test invalid method
	httpResp, tastyError = c.request(http.MethodGet+"/sdfl/", "/test", nil, nil, nil)
	require.NotNil(t, tastyError)
	require.Nil(t, httpResp)

	require.Equal(t,
		"\nError in request 0;\nCode: \nMessage: Client Side Error: net/http: invalid method \"GET/sdfl/\"",
		tastyError.Error())

	// Test invalid URL
	c.baseURL = "invalid"
	httpResp, tastyError = c.request(http.MethodGet, "/test", nil, nil, nil)
	require.NotNil(t, tastyError)
	require.Nil(t, httpResp)

	require.Equal(t,
		"\nError in request 0;\nCode: \nMessage: Client Side Error: Get \"invalid/test\": unsupported protocol scheme \"\"",
		tastyError.Error())
}

func TestNoAuthRequest(t *testing.T) {
	c := NewCertClient(&http.Client{Timeout: time.Duration(30) * time.Second})

	// Test invalid payload
	invalid := math.Inf(1)
	httpResp, tastyError := c.noAuthRequest(http.MethodGet, "/test", nil, nil, invalid, nil)
	require.NotNil(t, tastyError)
	require.Nil(t, httpResp)

	require.Equal(t,
		"\nError in request 0;\nCode: \nMessage: Client Side Error: json: unsupported value: +Inf",
		tastyError.Error())

	// Test invalid query
	httpResp, tastyError = c.noAuthRequest(http.MethodGet, "/test", nil, invalid, nil, nil)
	require.NotNil(t, tastyError)
	require.Nil(t, httpResp)

	require.Equal(t,
		"\nError in request 0;\nCode: \nMessage: Client Side Error: <nil>",
		tastyError.Error())

	// Test invalid method
	httpResp, tastyError = c.noAuthRequest(http.MethodGet+"/sdfl/", "/test", nil, nil, nil, nil)
	require.NotNil(t, tastyError)
	require.Nil(t, httpResp)

	require.Equal(t,
		"\nError in request 0;\nCode: \nMessage: Client Side Error: net/http: invalid method \"GET/sdfl/\"",
		tastyError.Error())

	// Test invalid URL
	c.baseURL = "invalid"
	httpResp, tastyError = c.noAuthRequest(http.MethodGet, "/test", nil, nil, nil, nil)
	require.NotNil(t, tastyError)
	require.Nil(t, httpResp)

	require.Equal(t,
		"\nError in request 0;\nCode: \nMessage: Client Side Error: Get \"invalid/test\": unsupported protocol scheme \"\"",
		tastyError.Error())
}

func TestCustomRequestNoContent(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/no-content", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	})

	httpResp, err := client.customRequest(http.MethodGet, "/no-content", nil, nil, nil)
	require.Nil(t, err)
	require.NotNil(t, httpResp)
}

func TestRequestNoContent(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/no-content", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	})

	httpResp, err := client.request(http.MethodGet, "/no-content", nil, nil, nil)
	require.Nil(t, err)
	require.NotNil(t, httpResp)
}

func TestNoAuthRequestNoContent(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/no-content", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	})

	httpResp, err := client.noAuthRequest(http.MethodGet, "/no-content", nil, nil, nil, nil)
	require.Nil(t, err)
	require.NotNil(t, httpResp)
}

func TestCustomRequestErrorResponses(t *testing.T) {
	setup()
	defer teardown()

	for _, errCode := range errorStatusCodes {
		path := fmt.Sprintf("/error/%d", errCode)
		mux.HandleFunc(path, func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(errCode)
		})

		httpResp, err := client.customRequest(http.MethodGet, path, nil, nil, nil)
		require.NotNil(t, err)
		require.NotNil(t, httpResp)

		require.Equal(t, errCode, err.StatusCode)
	}
}

func TestRequestErrorResponses(t *testing.T) {
	setup()
	defer teardown()

	for _, errCode := range errorStatusCodes {
		path := fmt.Sprintf("/error/%d", errCode)
		mux.HandleFunc(path, func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(errCode)
		})

		httpResp, err := client.request(http.MethodGet, path, nil, nil, nil)
		require.NotNil(t, err)
		require.NotNil(t, httpResp)

		require.Equal(t, errCode, err.StatusCode)
	}
}

func TestNoAuthRequestErrorResponses(t *testing.T) {
	setup()
	defer teardown()

	for _, errCode := range errorStatusCodes {
		path := fmt.Sprintf("/error/%d", errCode)
		mux.HandleFunc(path, func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(errCode)
		})

		httpResp, err := client.noAuthRequest(http.MethodGet, path, nil, nil, nil, nil)
		require.NotNil(t, err)
		require.NotNil(t, httpResp)

		require.Equal(t, errCode, err.StatusCode)
	}
}

func TestCustomRequestInvalidResult(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/invalid", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprint(writer, map[string]string{"test-key": "value"})
	})

	httpResp, err := client.customRequest(http.MethodGet, "/invalid", nil, nil, math.Inf(1))
	require.NotNil(t, err)
	require.NotNil(t, httpResp)
}

func TestRequestInvalidResult(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/invalid", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprint(writer, map[string]string{"test-key": "value"})
	})

	httpResp, err := client.request(http.MethodGet, "/invalid", nil, nil, math.Inf(1))
	require.NotNil(t, err)
	require.NotNil(t, httpResp)
}

func TestNoAuthRequestInvalidResult(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/invalid", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprint(writer, map[string]string{"test-key": "value"})
	})

	httpResp, err := client.noAuthRequest(http.MethodGet, "/invalid", nil, nil, nil, math.Inf(1))
	require.NotNil(t, err)
	require.NotNil(t, httpResp)
}

func TestCustomRequestMissingCredentials(t *testing.T) {
	c := NewClient(&http.Client{Timeout: time.Duration(30) * time.Second})

	httpResp, tastyErr := c.customRequest(http.MethodGet, "/invalid", nil, nil, nil)
	require.NotNil(t, tastyErr)
	require.Nil(t, httpResp)

	require.Equal(t,
		"\nError in request 0;\nCode: invalid_session\nMessage: Session is invalid: Session Token cannot be nil.",
		tastyErr.Error())
}

func TestRequestMissingCredentials(t *testing.T) {
	c := NewClient(&http.Client{Timeout: time.Duration(30) * time.Second})

	httpResp, tastyErr := c.customRequest(http.MethodGet, "/invalid", nil, nil, nil)
	require.NotNil(t, tastyErr)
	require.Nil(t, httpResp)

	require.Equal(t,
		"\nError in request 0;\nCode: invalid_session\nMessage: Session is invalid: Session Token cannot be nil.",
		tastyErr.Error())
}

func TestNoAuthRequestWithParams(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/with-params", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
		require.Equal(t, "true", request.URL.Query().Get("is-etf"))
	})

	httpResp, err := client.noAuthRequest(http.MethodGet, "/with-params", nil, EquitiesQuery{IsETF: true}, nil, nil)
	require.Nil(t, err)
	require.NotNil(t, httpResp)
}

const tastyUnauthorizedError = `{
    "error": {
        "code": "unauthorized",
        "message": "Unauthorized. Unique customer support identifier: test"
    }
}`

const tastyInvalidCredentialsError = `{
    "error": {
        "code": "invalid_credentials",
        "message": "Invalid login, please check your username and password. Unique customer support identifier: test-id"
    }
}`

const tastyInvalidSessionError = `{
    "error": {
        "code": "invalid_session",
        "message": "Session user not present. Unique customer support identifier: test-id"
    }
}`

func expectedUnauthorized(t *testing.T, err error) {
	require.NotNil(t, err)

	require.Equal(t, "\nError in request 401;\nCode: unauthorized\nMessage: Unauthorized. Unique customer support identifier: test", err.Error())
}

func expectedInvalidCredentials(t *testing.T, err error) {
	require.NotNil(t, err)

	require.Equal(t,
		"\nError in request 401;\nCode: invalid_credentials\nMessage: Invalid login, please check your username and password. Unique customer support identifier: test-id",
		err.Error())
}

func expectedInvalidSession(t *testing.T, err error) {
	require.NotNil(t, err)

	require.Equal(t,
		"\nError in request 401;\nCode: invalid_session\nMessage: Session user not present. Unique customer support identifier: test-id",
		err.Error())
}
