package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/polarismesh/polaris-go"
)

var (
	namespace     = "default"
	fileGroup     = "test"
	fileName      = "default.yaml"
	originContent = `server:
		port: 8080`
	updatedContent = `server:
		port: 8090`
	configCenterUrl = "http://127.0.0.1:8090"
)

func makeJsonRequest(uri string, data string, method string, headers map[string]string) ([]byte, error) {
	client := http.Client{}
	req, err := http.NewRequest(method, uri, strings.NewReader(data))
	req.Header.Add("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Add(k, v)
	}
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

type commonRes struct {
	Code int32 `json:"code"`
}

type LoginRes struct {
	Code          int32 `json:"code"`
	LoginResponse struct {
		Token string `json:"token"`
	} `json:"loginResponse"`
}

type configClient struct {
	token string
}

func newConfigClient() (*configClient, error) {
	token, err := getToken()
	if err != nil {
		return nil, err

	}
	return &configClient{
		token: token,
	}, nil
}

func getToken() (string, error) {
	data, err := json.Marshal(map[string]string{
		"name":     "polaris",
		"password": "polaris",
	})
	if err != nil {
		return "", err
	}
	// login use default user
	res, err := makeJsonRequest(fmt.Sprintf("%s/core/v1/user/login", configCenterUrl), string(data), http.MethodPost, map[string]string{})
	if err != nil {
		return "", nil
	}
	var loginRes LoginRes
	if err = json.Unmarshal(res, &loginRes); err != nil {
		return "", err
	}
	return loginRes.LoginResponse.Token, nil
}

func (client *configClient) createConfigFile() error {
	data, err := json.Marshal(map[string]string{
		"name":      fileName,
		"namespace": namespace,
		"group":     fileGroup,
		"content":   originContent,
		"modifyBy":  "polaris",
		"format":    "yaml",
	})
	if err != nil {
		return err
	}
	res, err := makeJsonRequest(fmt.Sprintf("%s/config/v1/configfiles", configCenterUrl), string(data), http.MethodPost, map[string]string{
		"X-Polaris-Token": client.token,
	})
	if err != nil {
		return err
	}

	var resJson commonRes
	err = json.Unmarshal(res, &resJson)
	if err != nil {
		return err
	}
	if resJson.Code != 200000 {
		return errors.New("create error")
	}
	return nil
}

func (client *configClient) updateConfigFile() error {
	data, err := json.Marshal(map[string]string{
		"name":      fileName,
		"namespace": namespace,
		"group":     fileGroup,
		"content":   updatedContent,
		"modifyBy":  "polaris",
		"format":    "yaml",
	})
	if err != nil {
		return err
	}
	res, err := makeJsonRequest(fmt.Sprintf("%s/config/v1/configfiles", configCenterUrl), string(data), http.MethodPut, map[string]string{
		"X-Polaris-Token": client.token,
	})
	if err != nil {
		return err
	}
	var resJson commonRes
	err = json.Unmarshal(res, &resJson)
	if err != nil {
		return err
	}
	if resJson.Code != 200000 {
		return errors.New("update error")
	}
	return nil
}

func (client *configClient) deleteConfigFile() error {
	data, err := json.Marshal(map[string]string{})
	if err != nil {
		return err
	}
	res, err := makeJsonRequest(fmt.Sprintf("%s/config/v1/configfiles?namespace=%s&group=%s&name=%s", configCenterUrl, namespace, fileGroup, fileName), string(data), http.MethodDelete, map[string]string{
		"X-Polaris-Token": client.token,
	})
	if err != nil {
		return err
	}
	var resJson commonRes
	err = json.Unmarshal(res, &resJson)
	if err != nil {
		return err
	}
	if resJson.Code != 200000 {
		return errors.New("delete error")
	}
	return nil
}

func (client *configClient) publishConfigFile() error {
	data, err := json.Marshal(map[string]string{
		"namespace": namespace,
		"group":     fileGroup,
		"fileName":  fileName,
		"comment":   "config update",
	})
	if err != nil {
		return err
	}
	res, err := makeJsonRequest(fmt.Sprintf("%s/config/v1/configfiles/release", configCenterUrl), string(data), http.MethodPost, map[string]string{
		"X-Polaris-Token": client.token,
	})
	if err != nil {
		return err
	}
	var resJson commonRes
	err = json.Unmarshal(res, &resJson)
	if err != nil {
		return err
	}
	if resJson.Code != 200000 {
		return errors.New("publish error")
	}
	return nil

}

func TestConfig(t *testing.T) {
	client, err := newConfigClient()
	if err != nil {
		t.Fatal(err)
	}
	if err = client.createConfigFile(); err != nil {
		t.Fatal(err)
	}
	if err = client.publishConfigFile(); err != nil {
		t.Fatal(err)
	}

	// Always remember clear test resource
	defer func() {
		client.deleteConfigFile()
	}()
	configApi, err := polaris.NewConfigAPI()
	if err != nil {
		t.Fatal(err)
	}
	config, err := New(&configApi, WithNamespace(namespace), WithFileGroup(fileGroup), WithFileName(fileName))
	if err != nil {
		t.Fatal(err)
	}
	kv, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}

	if len(kv) != 1 || kv[0].Key != fileName || string(kv[0].Value) != originContent {
		t.Fatal("config error")
	}

	w, err := config.Watch()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		w.Stop()
	}()

	if err = client.updateConfigFile(); err != nil {
		t.Fatal(err)
	}

	if err = client.publishConfigFile(); err != nil {
		t.Fatal(err)
	}

	if kv, err = w.Next(); err != nil {
		t.Fatal(err)
	}

	if len(kv) != 1 || kv[0].Key != fileName || string(kv[0].Value) != updatedContent {
		t.Fatal("config error")
	}
}

func TestExtToFormat(t *testing.T) {
	client, err := newConfigClient()
	if err != nil {
		t.Fatal(err)
	}
	if err = client.createConfigFile(); err != nil {
		t.Fatal(err)
	}
	if err = client.publishConfigFile(); err != nil {
		t.Fatal(err)
	}

	// Always remember clear test resource
	defer func() {
		client.deleteConfigFile()
	}()

	configApi, err := polaris.NewConfigAPI()
	if err != nil {
		t.Fatal(err)
	}

	config, err := New(&configApi, WithNamespace(namespace), WithFileGroup(fileGroup), WithFileName(fileName))
	if err != nil {
		t.Fatal(err)
	}
	kv, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(len(kv), 1) {
		t.Errorf("len(kvs) = %d", len(kv))
	}
	if !reflect.DeepEqual(fileName, kv[0].Key) {
		t.Errorf("kvs[0].Key is %s", kv[0].Key)
	}
	if !reflect.DeepEqual(originContent, string(kv[0].Value)) {
		t.Errorf("kvs[0].Value is %s", kv[0].Value)
	}
	if !reflect.DeepEqual("yaml", kv[0].Format) {
		t.Errorf("kvs[0].Format is %s", kv[0].Format)
	}
}
