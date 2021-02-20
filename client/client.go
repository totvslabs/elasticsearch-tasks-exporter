package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
)

// Client is the elasticsearch task client
type Client interface {
	Tasks() ([]Task, error)
}

type client struct {
	baseURL string
}

// New creates a new elasticsearch client
func New(url string) Client {
	return &client{
		baseURL: url,
	}
}

func (c *client) Tasks() ([]Task, error) {
	log.Debugf("querying %s...", c.baseURL)
	var result Result
	resp, err := http.Get(c.baseURL + "/_cluster/pending_tasks")
	if err != nil {
		return []Task{}, errors.Wrap(err, "failed to get metrics")
	}
	if resp.StatusCode != 200 {
		return []Task{}, fmt.Errorf("failed to get metrics: %s", resp.Status)
	}
	defer resp.Body.Close()
	bts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []Task{}, errors.Wrap(err, "failed to parse metrics")
	}
	log.Debugf("body: %s", string(bts))
	if err := json.Unmarshal(bts, &result); err != nil {
		return []Task{}, errors.Wrap(err, "failed to parse metrics")
	}

	return result.Tasks, nil
}

type Result struct {
	Tasks []Task `json:"tasks"`
}

type Task struct {
	Executing bool   `json:"executing"`
	Priority  string `json:"priority"`
	Source    string `json:"source"`
}
