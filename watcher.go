package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/polarismesh/polaris-go"
	"github.com/polarismesh/polaris-go/pkg/model"
)

type Watcher struct {
	configFile polaris.ConfigFile
	closed     bool
	changeChan chan model.ConfigFileChangeEvent
}

func newWatcher(configFile polaris.ConfigFile) *Watcher {
	changeChan := make(chan model.ConfigFileChangeEvent)
	configFile.AddChangeListenerWithChannel(changeChan)

	w := &Watcher{
		configFile: configFile,
		changeChan: changeChan,
	}
	return w
}

func (w *Watcher) Next() ([]*config.KeyValue, error) {

	select {
	case event := <-w.changeChan:
		fmt.Println(event)
		return []*config.KeyValue{
			{
				Key:    w.configFile.GetFileName(),
				Value:  []byte(event.NewValue),
				Format: strings.TrimPrefix(filepath.Ext(w.configFile.GetFileName()), "."),
			},
		}, nil
	}
}

func (w *Watcher) Close() error {
	if !w.closed {
		close(w.changeChan)
	}
	return nil
}

func (w *Watcher) Stop() error {
	w.Close()
	return nil
}
