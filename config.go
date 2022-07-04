package config

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/polarismesh/polaris-go"
)

// Option is etcd config option.
type Option func(o *options)

type options struct {
	ctx        context.Context
	namespace  string
	fileGroup  string
	fileName   string
	configFile polaris.ConfigFile
}

// WithContext with registry context.
func WithContext(ctx context.Context) Option {
	return func(o *options) {
		o.ctx = ctx
	}
}

// WithNamespace is config namespace
func WithNamespace(namespace string) Option {
	return func(o *options) {
		o.namespace = namespace
	}
}

// WithFileGroup is config fileGroup
func WithFileGroup(fileGroup string) Option {
	return func(o *options) {
		o.fileGroup = fileGroup
	}
}

// WithFileName is config fileName
func WithFileName(fileName string) Option {
	return func(o *options) {
		o.fileName = fileName
	}
}

type source struct {
	client  *polaris.ConfigAPI
	options *options
}

func New(client *polaris.ConfigAPI, opts ...Option) (config.Source, error) {
	options := &options{
		ctx:       context.Background(),
		namespace: "default",
		fileGroup: "",
		fileName:  "",
	}

	for _, opt := range opts {
		opt(options)
	}

	if options.fileGroup == "" {
		return nil, errors.New("fileGroup invalid")
	}

	if options.fileName == "" {
		return nil, errors.New("fileName invalid")
	}

	return &source{
		client:  client,
		options: options,
	}, nil
}

// Load return the config values
func (s *source) Load() ([]*config.KeyValue, error) {
	configAPI, err := polaris.NewConfigAPI()

	if err != nil {
		return nil, err
	}

	configFile, err := configAPI.GetConfigFile(s.options.namespace, s.options.fileGroup, s.options.fileName)
	if err != nil {
		fmt.Println("fail to get config.", err)
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	content := configFile.GetContent()
	// kvs := make([]*config.KeyValue, 0, len(content))
	k := s.options.fileName

	s.options.configFile = configFile

	return []*config.KeyValue{
		{
			Key:    k,
			Value:  []byte(content),
			Format: strings.TrimPrefix(filepath.Ext(k), "."),
		},
	}, nil
}

// Watch return the watcher
func (s *source) Watch() (config.Watcher, error) {
	return newWatcher(s.options.configFile), nil
}
