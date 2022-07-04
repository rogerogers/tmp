package config

import (
	"fmt"
	"testing"

	"github.com/polarismesh/polaris-go"
)

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
	w, err := config.Watch()
	defer func() {
		w.Stop()
	}()

	v, err := w.Next()
	if err != nil {
		t.Fatal(err)
	}
	newConfig, err := configApi.GetConfigFile("default", "aaa", "a.yaml")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(v)
}
