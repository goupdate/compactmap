package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/goupdate/compactmap/structmap"
)

type Client[V any] struct {
	baseURL string
	client  *http.Client
}

func New[V any](baseURL string) *Client[V] {
	return &Client[V]{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (c *Client[V]) post(endpoint string, requestBody interface{}) ([]byte, error) {
	var body []byte
	var err error

	if requestBody != nil {
		body, err = json.Marshal(requestBody)
		if err != nil {
			return nil, err
		}
	}

	resp, err := c.client.Post(c.baseURL+endpoint, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorMessage, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf(string(errorMessage))
	}

	return ioutil.ReadAll(resp.Body)
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
	return result.Id, err
}

func (c *Client[V]) Get(id int64) (*V, error) {
	resp, err := c.client.Get(fmt.Sprintf("%s/api/get?id=%d", c.baseURL, id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(buf))
	}
	//not found
	if len(buf) == 0 {
		return nil, err
	}

	var item V
	err = json.Unmarshal(buf, &item)
	return &item, err
}

func (c *Client[V]) Delete(id int64) error {
	resp, err := c.client.Get(fmt.Sprintf("%s/api/delete?id=%d", c.baseURL, id))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorMessage, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf(string(errorMessage))
	}

	return nil
}

func (c *Client[V]) Update(condition string, where []structmap.FindCondition, fields map[string]interface{}) (int, error) {
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
	return err
}

func (c *Client[V]) Find(condition string, where []structmap.FindCondition) ([]V, error) {
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
	return results, err
}

func (c *Client[V]) Iterate() ([]V, error) {
	resp, err := c.client.Get(fmt.Sprintf("%s/api/iterate", c.baseURL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorMessage, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf(string(errorMessage))
	}

	var results []V
	err = json.NewDecoder(resp.Body).Decode(&results)
	return results, err
}
