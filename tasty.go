package tasty

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
)

const (
	apiBaseURL          = "https://api.tastyworks.com"
	apiBaseHost         = "api.tastyworks.com"
	apiCertBaseURL      = "https://api.cert.tastyworks.com"
	apiCertBaseHost     = "api.cert.tastyworks.com"
	streamerBaseURL     = "wss://streamer.tastyworks.com"
	streamerCertBaseURL = "wss://streamer.cert.tastyworks.com"
)

var (
	defaultHTTPClient = &http.Client{Timeout: time.Duration(30) * time.Second}
	errorStatusCodes  = []int{400, 401, 403, 404, 415, 422, 500}
)

// Client for the tasty api wrapper.
type Client struct {
	httpClient *http.Client
	baseURL    string
	baseHost   string
	websocket  string
	Session    Session
}

// NewClient creates a new Tasty Client.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = defaultHTTPClient
	}
	c := &Client{
		httpClient: httpClient,
		baseURL:    apiBaseURL,
		baseHost:   apiBaseHost,
		websocket:  streamerBaseURL,
	}

	return c
}

// NewCertClient creates a new Tasty Cert Client.
func NewCertClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = defaultHTTPClient
	}
	c := &Client{
		httpClient: httpClient,
		baseURL:    apiCertBaseURL,
		baseHost:   apiCertBaseHost,
		websocket:  streamerCertBaseURL,
	}

	return c
}

// Getter for the tastytrade account streaming websocket url.
func (c Client) GetWebsocketURL() string {
	return c.websocket
}

// Error reasoning given by tastytrade.
type ErrorResponse struct {
	Domain string `json:"domain"`
	Reason string `json:"reason"`
}

// Error represents an error returned by the tastytrade API.
type Error struct {
	// Simple code error string
	Code string `json:"code"`
	// A short description of the error.
	Message string `json:"message"`
	// Slice of errors
	Errors []ErrorResponse `json:"errors"`
	// The HTTP status code.
	StatusCode int `json:"error,omitempty"`
}

// Error ...
func (e Error) Error() string {
	return fmt.Sprintf("\nError in request %d;\nCode: %s\nMessage: %s", e.StatusCode, e.Code, e.Message)
}

// decodeError decodes an Error from response status code based off
// the developer docs in tastytrade -> https://developer.tastytrade.com/#error-codes
func decodeError(resp *http.Response) *Error {
	e := new(Error)

	type errorRes struct {
		Error Error `json:"error"`
	}

	errRes := new(errorRes)

	err := json.NewDecoder(resp.Body).Decode(errRes)
	if err != nil {
		e.Message = fmt.Sprintf("tastytrade: unexpected HTTP %d: %s (empty error)", resp.StatusCode, err.Error())
		e.StatusCode = resp.StatusCode
		return e
	}

	errRes.Error.StatusCode = resp.StatusCode

	e = &errRes.Error

	return e
}

// customRequest handles any requests for the client with unique paths.
func (c *Client) customRequest(method, path string, params, payload, result any) (*http.Response, *Error) {
	if c.Session.SessionToken == nil {
		return nil, &Error{Code: "invalid_session", Message: "Session is invalid: Session Token cannot be nil."}
	}

	r := new(http.Request)

	r.Method = method

	r.URL = &url.URL{
		Scheme: strings.Split(c.baseURL, ":")[0],
		Host:   c.baseHost,
		Opaque: fmt.Sprintf("//%s%s", c.baseHost, path),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &Error{Message: fmt.Sprintf("Client Side Error: %v", err)}
	}

	r.Body = io.NopCloser(bytes.NewBuffer(body))

	r.Header = http.Header{}
	r.Header.Add("Authorization", *c.Session.SessionToken)
	r.Header.Add("Content-Type", "application/json")

	if params != nil {
		queryString, queryErr := query.Values(params)
		if queryErr != nil {
			return nil, &Error{Message: fmt.Sprintf("Client Side Error: %v", err)}
		}
		r.URL.RawQuery = queryString.Encode()
	}

	resp, err := c.httpClient.Do(r)
	if err != nil {
		return nil, &Error{Message: fmt.Sprintf("Client Side Error: %v", err)}
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return resp, nil
	}
	if containsInt(errorStatusCodes, resp.StatusCode) {
		return resp, decodeError(resp)
	}

	if result != nil {
		err = json.NewDecoder(resp.Body).Decode(result)
		if err != nil {
			return resp, &Error{Message: fmt.Sprintf("Client Side Error: %v", err)}
		}
	}

	return resp, nil
}

// request handles any requests for the client.
func (c *Client) request(method, path string, params, payload, result any) (*http.Response, *Error) {
	if c.Session.SessionToken == nil {
		return nil, &Error{Code: "invalid_session", Message: "Session is invalid: Session Token cannot be nil."}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &Error{Message: fmt.Sprintf("Client Side Error: %v", err)}
	}

	fullURL := c.baseURL + path

	r, err := http.NewRequest(method, fullURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, &Error{Message: fmt.Sprintf("Client Side Error: %v", err)}
	}

	r.Header = http.Header{}
	r.Header.Add("Authorization", *c.Session.SessionToken)
	r.Header.Add("Content-Type", "application/json")

	if params != nil {
		queryString, queryErr := query.Values(params)
		if queryErr != nil {
			return nil, &Error{Message: fmt.Sprintf("Client Side Error: %v", err)}
		}
		r.URL.RawQuery = queryString.Encode()
	}

	resp, err := c.httpClient.Do(r)
	if err != nil {
		return nil, &Error{Message: fmt.Sprintf("Client Side Error: %v", err)}
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return resp, nil
	}
	if containsInt(errorStatusCodes, resp.StatusCode) {
		return resp, decodeError(resp)
	}

	if result != nil {
		err = json.NewDecoder(resp.Body).Decode(result)
		if err != nil {
			return resp, &Error{Message: fmt.Sprintf("Client Side Error: %v", err)}
		}
	}

	return resp, nil
}

// noAuthRequest handles any requests for the client without authentication.
func (c *Client) noAuthRequest(method, path string, header http.Header, params, payload, result any) (*http.Response, *Error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &Error{Message: fmt.Sprintf("Client Side Error: %v", err)}
	}

	fullURL := c.baseURL + path

	r, err := http.NewRequest(method, fullURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, &Error{Message: fmt.Sprintf("Client Side Error: %v", err)}
	}

	if header == nil {
		r.Header = http.Header{}
	} else {
		r.Header = header
	}

	r.Header.Add("Content-Type", "application/json")

	if params != nil {
		queryString, queryErr := query.Values(params)
		if queryErr != nil {
			return nil, &Error{Message: fmt.Sprintf("Client Side Error: %v", err)}
		}
		r.URL.RawQuery = queryString.Encode()
	}

	resp, err := c.httpClient.Do(r)
	if err != nil {
		return nil, &Error{Message: fmt.Sprintf("Client Side Error: %v", err)}
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return resp, nil
	}
	if containsInt(errorStatusCodes, resp.StatusCode) {
		return resp, decodeError(resp)
	}

	if result != nil {
		err = json.NewDecoder(resp.Body).Decode(result)
		if err != nil {
			return resp, &Error{Message: fmt.Sprintf("Client Side Error: %v", err)}
		}
	}

	return resp, nil
}
