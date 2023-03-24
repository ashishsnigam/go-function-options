// functions options pattern example based on https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

// OptReqParams contains all optional parameters which are used for valid/invalid request call like invalid token
// Making use of Function Options pattern to initialize this struct with any number of fields
type OptReqParams struct {
	httpMethod      string
	body            io.Reader
	useInvalidToken bool
	queryParam      map[string]string
	acceptHeader    string
}

// OptReqParamsOption takes pointer to OptReqParams and modifies some fields in With below
type OptReqParamsOption func(*OptReqParams)

// NewOptReqParams takes a slice of option as the rest arguments
func NewOptReqParams(options ...OptReqParamsOption) *OptReqParams {
	params := &OptReqParams{}
	params.httpMethod = http.MethodGet           // default value for http method
	params.useInvalidToken = false               // default value for invalid token
	params.acceptHeader = "application/json"     // default value for headers
	for _, o := range options {
		// Call the option giving the instantiated *OptReqParams as the argument
		o(params)
	}
	// return the modified params instance
	return params
}

// WithMethod is a higher order function which returns a function
// it is actually creating a function which is setting only the given input only
// it's called capturing the argument. you can have this function to capture more arguments also, but this is not recommended
// options pattern is all about composition and preferred way is capturing one argument per function
func WithMethod(httpMethod string) OptReqParamsOption {
	return func(s *OptReqParams) {
		s.httpMethod = httpMethod
	}
}

func WithBody(body io.Reader) OptReqParamsOption {
	// another way as above with same functionality
	f := func(s *OptReqParams) {
		s.body = body
	}
	return f
}

func WithUseInvalidToken(useInvalidToken bool) OptReqParamsOption {
	return func(s *OptReqParams) {
		s.useInvalidToken = useInvalidToken
	}
}

func WithQueryParam(queryParam map[string]string) OptReqParamsOption {
	return func(s *OptReqParams) {
		s.queryParam = queryParam
	}
}

func WithAcceptHeader(acceptHeader string) OptReqParamsOption {
	return func(s *OptReqParams) {
		s.acceptHeader = acceptHeader
	}
}

// WithTwoValues is also allowed but not a standard way of doing thing
func WithTwoValues(acceptHeader string, httpMethod string) OptReqParamsOption {
	return func(s *OptReqParams) {
		s.httpMethod = httpMethod
		s.acceptHeader = acceptHeader
	}
}

// CustomHTTPRequest makes direct call of apis with optional fields required
func CustomHTTPRequest(ctx context.Context, url, email, passwd string, p *OptReqParams) (*http.Response, error) {
	var authString string
	if p.useInvalidToken { // default set to false in constructor NewOptReqParams
		authString = fmt.Sprintf("Bearer %s", "Invalid Token")
	} else {
		// call your login api to get valid token
		resp, err := MyLoginAPI(ctx, email, passwd)
		if err != nil {
			msg := fmt.Sprintf("error in login with user provided credentials %v", err)
			return nil, errors.New(msg)
		}
		authString = fmt.Sprintf("Bearer %s", resp.Token)
	}

	// create http req
	client := http.Client{}
	req, err := http.NewRequestWithContext(ctx, p.httpMethod, url, p.body)
	if err != nil {
		return nil, err
	}

	// add required headers
	req.Header.Add("Accept", p.acceptHeader)
	req.Header.Add("Authorization", authString)
	req.Header.Add("Content-Type", "application/json")

	// build query params for request
	if p.queryParam != nil {
		q := req.URL.Query()
		for k := range p.queryParam {
			q.Add(k, p.queryParam[k])
		}
		req.URL.RawQuery = q.Encode()
	}

	// fire request
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return res, err
}

func main() {
	// call CustomHTTPRequest with no optional parameter, only default values in constructor are used
	_, _ = CustomHTTPRequest(context.Background(), "some_url", "email_addr", "email_passwd", nil)

	// call CustomHTTPRequest with few additional parameters - post method with no body and query param
	q := make(map[string]string)
	q["name"] = "xyz"
	q["age"] = "10"
	p := NewOptReqParams(WithMethod(http.MethodPost), WithBody(nil), WithQueryParam(q))
	_, _ = CustomHTTPRequest(context.Background(), "some_url", "email_addr", "email_passwd", p)
}