package config

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/polarismesh/polaris-go"
	"github.com/polarismesh/polaris-go/pkg/model"
)

type Watcher struct {
	configFile polaris.ConfigFile

	content chan string
	// cancelListenConfig cancelListenConfigFunc
	closed     bool
	ctx        context.Context
	cancel     context.CancelFunc
	changeChan chan model.ConfigFileChangeEvent
}

// type cancelListenConfigFunc func(params vo.ConfigParam) (err error)

func newWatcher(configFile polaris.ConfigFile) *Watcher {
	// ctx, cancel := context.WithCancel(ctx)
	changeChan := make(chan model.ConfigFileChangeEvent)
	configFile.AddChangeListenerWithChannel(changeChan)

	w := &Watcher{
		configFile: configFile,
		// cancelListenConfig: cancelListenConfig,
		content:    make(chan string, 100),
		changeChan: changeChan,

		// ctx:    ctx,
		// cancel: cancel,
	}
	return w
}

func (w *Watcher) Next() ([]*config.KeyValue, error) {
	// select {
	// case <-w.ctx.Done():
	// 	return nil, w.ctx.Err()
	// case content := <-w.content:
	// 	k := w.dataID

	select {
	case event := <-w.changeChan:
		fmt.Println(fmt.Sprintf("received change event by channel. %+v", event))
	}
	return []*config.KeyValue{
		{
			Key:    "fuck",
			Value:  []byte("fuckyou"),
			Format: strings.TrimPrefix(filepath.Ext("key.json"), "."),
		},
	}, nil
}

func (w *Watcher) Close() error {
	if w.closed == false {
		close(w.changeChan)
	}
	return nil
}

func (w *Watcher) Stop() error {
	w.Close()
	return nil
}
