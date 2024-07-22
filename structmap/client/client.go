package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MasterDimmy/zipologger"
	"github.com/goupdate/compactmap/structmap"
	"github.com/valyala/fasthttp"
)

var (
	Timeout = 15 * time.Second
)

type Client[V any] struct {
	baseURL   string
	client    *fasthttp.Client
	error_log *zipologger.Logger
}

func New[V any](baseURL string) *Client[V] {
	return &Client[V]{
		baseURL: baseURL,
		client: &fasthttp.Client{
			ReadTimeout:  Timeout,
			WriteTimeout: Timeout,
		},
	}
}

func (c *Client[V]) SetErrorLog(log *zipologger.Logger) *Client[V] {
	c.error_log = log
	return c
}

// panics on unknown condition
func checkCondition(cond string) {
	switch cond {
	case "AND", "OR", "":
	default:
		panic("unknown condition: \"" + cond + "\"")
	}
}

func (c *Client[V]) post(endpoint string, requestBody interface{}) ([]byte, error) {
	var body []byte
	var err error

	if requestBody != nil {
		body, err = json.Marshal(requestBody)
		if err != nil {
			if c.error_log != nil {
				c.error_log.Printf("post error: [%s] : [%s]", string(body), err.Error())
			}
			return nil, err
		}
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.SetRequestURI(c.baseURL + endpoint)
	req.Header.SetMethod("POST")
	req.Header.SetContentType("application/json")
	req.SetBody(body)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err = c.client.DoTimeout(req, resp, Timeout)
	if err != nil {
		if c.error_log != nil {
			c.error_log.Printf("timeout: [%s]", err.Error())
		}
		return nil, err
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		if c.error_log != nil {
			c.error_log.Printf("incorrect status [%d] : [%s]", resp.StatusCode(), string(resp.Body()))
		}
		return nil, fmt.Errorf(string(resp.Body()))
	}

	return bytes.Clone(resp.Body()), nil
}

func (c *Client[V]) get(endpoint string) ([]byte, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.SetRequestURI(c.baseURL + endpoint)
	req.Header.SetMethod("GET")

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err := c.client.DoTimeout(req, resp, Timeout)
	if err != nil {
		if c.error_log != nil {
			c.error_log.Printf("get error: [%s]", err.Error())
		}
		return nil, err
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		if c.error_log != nil {
			c.error_log.Printf("incorrect status [%d] : [%s]", resp.StatusCode(), string(resp.Body()))
		}
		return nil, fmt.Errorf(string(resp.Body()))
	}

	return bytes.Clone(resp.Body()), nil
}

func (c *Client[V]) Clear() error {
	_, err := c.post("/api/clear", nil)
	return err
}

func (c *Client[V]) Add(item *V) (int64, error) {
	response, err := c.post("/api/add", item)
	if err != nil {
		return 0, err
	}

	var result struct {
		Id int64 `json:"id"`
	}

	err = json.Unmarshal(response, &result)
	if err != nil {
		if c.error_log != nil {
			c.error_log.Printf("add error: [%s] : [%s]", string(response), err.Error())
		}
	}
	return result.Id, err
}

func (c *Client[V]) Get(id int64) (*V, error) {
	response, err := c.get(fmt.Sprintf("/api/get?id=%d", id))
	if err != nil {
		return nil, err
	}

	if len(response) == 0 {
		return nil, nil
	}

	var item V
	err = json.Unmarshal(response, &item)
	if err != nil {
		if c.error_log != nil {
			c.error_log.Printf("get error: [%s] : [%s]", string(response), err.Error())
		}
	}
	return &item, err
}

func (c *Client[V]) Delete(id int64) error {
	_, err := c.get(fmt.Sprintf("/api/delete?id=%d", id))
	return err
}

func (c *Client[V]) Update(condition string, where []structmap.FindCondition, fields map[string]interface{}) (int, error) {
	checkCondition(condition)

	req := struct {
		Condition string                    `json:"condition"`
		Where     []structmap.FindCondition `json:"where"`
		Fields    map[string]interface{}    `json:"fields"`
	}{
		Condition: condition,
		Where:     where,
		Fields:    fields,
	}

	response, err := c.post("/api/update", req)
	if err != nil {
		return 0, err
	}

	var result struct {
		Updated int `json:"updated"`
	}
	err = json.Unmarshal(response, &result)
	if err != nil {
		if c.error_log != nil {
			c.error_log.Printf("update error: [%s] : [%s]", string(response), err.Error())
		}
	}
	return result.Updated, err
}

// count - limit of elements to update
func (c *Client[V]) UpdateCount(condition string, where []structmap.FindCondition, fields map[string]interface{}, elCount int) ([]int64, error) {
	checkCondition(condition)

	req := struct {
		Random    bool                      `json:"random"` //update random count values?
		Count     int                       `json:"count"`
		Condition string                    `json:"condition"`
		Where     []structmap.FindCondition `json:"where"`
		Fields    map[string]interface{}    `json:"fields"`
	}{
		Count:     elCount,
		Condition: condition,
		Where:     where,
		Fields:    fields,
	}

	response, err := c.post("/api/updatecount", req)
	if err != nil {
		return nil, err
	}

	var result struct {
		Updated []int64 `json:"updated"`
	}
	err = json.Unmarshal(response, &result)
	if err != nil {
		if c.error_log != nil {
			c.error_log.Printf("update count error: [%s] : [%s]", string(response), err.Error())
		}
	}
	return result.Updated, err
}

// update random count elements
func (c *Client[V]) UpdateCountRandom(condition string, where []structmap.FindCondition, fields map[string]interface{}, elCount int) ([]int64, error) {
	checkCondition(condition)

	req := struct {
		Random    bool                      `json:"random"` //update random count values?
		Count     int                       `json:"count"`
		Condition string                    `json:"condition"`
		Where     []structmap.FindCondition `json:"where"`
		Fields    map[string]interface{}    `json:"fields"`
	}{
		Random:    true,
		Count:     elCount,
		Condition: condition,
		Where:     where,
		Fields:    fields,
	}

	response, err := c.post("/api/updatecount", req)
	if err != nil {
		return nil, err
	}

	var result struct {
		Updated []int64 `json:"updated"`
	}
	err = json.Unmarshal(response, &result)
	if err != nil {
		if c.error_log != nil {
			c.error_log.Printf("update count random error: [%s] : [%s]", string(response), err.Error())
		}
	}
	return result.Updated, err
}

func (c *Client[V]) SetField(id int64, field string, value interface{}) error {
	req := struct {
		Id    int64       `json:"id"`
		Field string      `json:"field"`
		Value interface{} `json:"value"`
	}{
		Id:    id,
		Field: field,
		Value: value,
	}

	_, err := c.post("/api/setfield", req)
	if err != nil {
		if c.error_log != nil {
			c.error_log.Printf("setfield error: [%s]", err.Error())
		}
	}
	return err
}

func (c *Client[V]) SetFields(id int64, fields map[string]interface{}) error {
	req := struct {
		Id     int64                  `json:"id"`
		Fields map[string]interface{} `json:"fields"`
	}{
		Id:     id,
		Fields: fields,
	}

	_, err := c.post("/api/setfields", req)
	if err != nil {
		if c.error_log != nil {
			c.error_log.Printf("setfields error: [%s]", err.Error())
		}
	}
	return err
}

func (c *Client[V]) Find(condition string, where []structmap.FindCondition) ([]V, error) {
	checkCondition(condition)

	req := struct {
		Condition string                    `json:"condition"`
		Where     []structmap.FindCondition `json:"where"`
	}{
		Condition: condition,
		Where:     where,
	}

	response, err := c.post("/api/find", req)
	if err != nil {
		return nil, err
	}

	var results []V
	err = json.Unmarshal(response, &results)
	if err != nil {
		if c.error_log != nil {
			c.error_log.Printf("find error: [%s] : [%s]", string(response), err.Error())
		}
	}
	return results, err
}

func (c *Client[V]) All() ([]V, error) {
	response, err := c.get("/api/all")
	if err != nil {
		return nil, err
	}

	var results []V
	err = json.Unmarshal(response, &results)
	if err != nil {
		if c.error_log != nil {
			c.error_log.Printf("all error: [%s] : [%s]", string(response), err.Error())
		}
	}
	return results, err
}
