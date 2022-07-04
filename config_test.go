package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/polarismesh/polaris-go"
)

func makeJsonRequest(uri string, data string, method string) ([]byte, error) {
	client := http.Client{}
	req, err := http.NewRequest(method, uri, strings.NewReader(data))
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}

func TestJsonPost(t *testing.T) {
	data, err := json.Marshal(map[string]string{
		"name":      "test.yaml",
		"namespace": "default",
		"group":     "test",
		"content":   "useLocalCache: true",
		"modifyBy":  "polaris",
		"format":    "yaml",
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := makeJsonRequest("http://localhost:8093/config/v1/configfiles", string(data), http.MethodPost)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%s", res)
}

func TestConfig(t *testing.T) {
	configApi, err := polaris.NewConfigAPI()
	if err != nil {
		t.Fatal(err)
	}
	config, err := New(&configApi, WithNamespace("default"), WithFileGroup("aaa"), WithFileName("a.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	kv, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range kv {
		fmt.Printf("%d-%s", k, v)
	}
	w, _ := config.Watch()
	defer func() {
		w.Stop()
	}()

	v, err := w.Next()
	if err != nil {
		t.Fatal(err)
	}
	// newConfig, err := configApi.GetConfigFile("default", "aaa", "a.yaml")
	if err != nil {
		t.Fatal(err)
	}

	// makeJsonPost("http://localhost:8093/config/v1/configfiles")
	// http.NewRequest(http.MethodPost, "localhost:8083/")

	fmt.Println(v)
}
