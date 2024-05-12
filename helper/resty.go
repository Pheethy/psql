package helper

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cast"
)

type Client struct {
	restyClient *resty.Client
	hostURL     string
	debug       bool
}

func initRestyClient(host string, debug bool) *resty.Client {
	rc := resty.New()
	rc.SetHostURL(host)
	rc.SetContentLength(true)
	rc.SetDebug(debug)
	if debug {
		rc.EnableTrace()
	}
	return rc
}

func New(host string, debug bool) *Client {
	c := &Client{
		restyClient: initRestyClient(host, debug),
		hostURL:     host,
		debug:       debug,
	}
	return c
}

func Get(url string, req *resty.Request) (*resty.Response, error) {
	resp, err := execute(resty.MethodGet, url, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func Post(url string, req *resty.Request) (*resty.Response, error) {
	resp, err := execute(resty.MethodPost, url, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func SetRequestBody(data map[string]interface{}, req *resty.Request) {
	req.SetFormData(cast.ToStringMapString(data))
}

func execute(method string, url string, req *resty.Request) (*resty.Response, error) {
	return req.Execute(method, url)
}

func GetBodyJSON(resp *resty.Response) (map[string]interface{}, error) {
	if resp == nil {
		return nil, errors.New("response not found")
	}
	var body = make(map[string]interface{})

	bodyBytes := resp.Body()

	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		return nil, err
	}

	if len(body) == 0 {
		return nil, errors.New("body not found")
	}
	return body, nil
}

func (c *Client) GetRestyClient() *resty.Client {
	return c.restyClient
}

func (c *Client) GetHost() string {
	return c.hostURL
}

func (c *Client) GetDebug() bool {
	return c.debug
}

func (c *Client) SetTimeout(second int) {
	newRc := c.restyClient.SetTimeout(time.Duration(second * int(time.Second)))
	c.restyClient = newRc
}

/*
-----------------------------------
Request
-----------------------------------
*/
func (c *Client) initRequest(header map[string]string) *resty.Request {
	req := c.GetRestyClient().R()
	c.setHeader(req, header)
	return req
}
func (c *Client) setHeader(req *resty.Request, header map[string]string) {
	if header == nil {
		header = map[string]string{}
	}

	// set default header here
	req.SetHeader("Content-Type", "application/json")
	req.SetHeader("Accept", "application/json")

	// set another header here
	if len(header) > 0 {
		for key, val := range header {
			if key == "Authorization" {
				req.SetAuthToken(val)
			} else {
				req.SetHeader(key, val)
			}
		}
	}
}

func (c *Client) NewRequest(header map[string]string) *resty.Request {
	return c.initRequest(header)
}

func (c *Client) Get(url string, header map[string]string) (*resty.Response, error) {
	req := c.initRequest(header)
	resp, err := execute(resty.MethodGet, url, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) Post(url string, header map[string]string, data map[string]string) (*resty.Response, error) {
	req := c.initRequest(header)
	if data != nil {
		req.SetFormData(cast.ToStringMapString(data))
	}
	resp, err := execute(resty.MethodPost, url, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) PostWithRawData(url string, header map[string]string, data interface{}) (*resty.Response, error) {
	req := c.initRequest(header)
	if data != nil {
		req.SetBody(data)
	}
	return execute(resty.MethodPost, url, req)
}

func (c *Client) Delete(url string, header map[string]string) (*resty.Response, error) {
	req := c.initRequest(header)
	return execute(resty.MethodDelete, url, req)
}

func (c *Client) DeleteWithRawData(url string, header map[string]string, data interface{}) (*resty.Response, error) {
	req := c.initRequest(header)
	if data != nil {
		req.SetBody(data)
	}
	return execute(resty.MethodDelete, url, req)
}

func (c *Client) Put(url string, header map[string]string, data map[string]string) (*resty.Response, error) {
	req := c.initRequest(header)
	if data != nil {
		req.SetFormData(cast.ToStringMapString(data))
	}
	resp, err := execute(resty.MethodPut, url, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) PatchWithRawData(url string, header map[string]string, data interface{}) (*resty.Response, error) {
	req := c.initRequest(header)
	if data != nil {
		req.SetBody(data)
	}
	return execute(resty.MethodPatch, url, req)
}

func (c *Client) Head(url string, header map[string]string) (*resty.Response, error) {
	req := c.initRequest(header)
	return execute(resty.MethodHead, url, req)
}