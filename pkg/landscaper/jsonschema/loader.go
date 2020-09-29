// Copyright 2020 Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package jsonschema

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/xeipuuv/gojsonreference"
	"github.com/xeipuuv/gojsonschema"
	"sigs.k8s.io/yaml"

	lsv1alpha1 "github.com/gardener/landscaper/pkg/apis/core/v1alpha1"
)

type LoaderWrapper struct {
	LoaderConfig
	gojsonschema.JSONLoader
}

func NewWrappedLoader(config LoaderConfig, loader gojsonschema.JSONLoader) gojsonschema.JSONLoader {
	if config.DefaultLoader == nil {
		config.DefaultLoader = loader
	}
	return &LoaderWrapper{
		LoaderConfig: config,
		JSONLoader:   loader,
	}
}

func (l LoaderWrapper) LoaderFactory() gojsonschema.JSONLoaderFactory {
	return &LoaderFactory{
		LoaderConfig: l.LoaderConfig,
	}
}

// Loader is the landscaper specific jsonscheme loader.
// It resolves referecens of type: local, blueprint and cd.
type Loader struct {
	LoaderConfig
	source string
}

// LoaderConfig is the landscaper specific laoder configuration
// to resolve landscaper specific schema refs.
type LoaderConfig struct {
	// LocalTypes is a map of blueprint locally defined types.
	// It is a map of schema name to schema definition
	LocalTypes map[string]lsv1alpha1.JSONSchemaDefinition
	// BlueprintFs is the virtual filesystem that is used to resolve "blueprint" refs
	BlueprintFs vfs.FileSystem
	// DefaultLoader is the fallback loader that is used of the protocol is unknown.
	DefaultLoader gojsonschema.JSONLoader
}

// LoaderFactory is the factory that creates a new landscaper specific loader.
type LoaderFactory struct {
	LoaderConfig
}

func (l LoaderFactory) New(source string) gojsonschema.JSONLoader {
	return &Loader{
		LoaderConfig: l.LoaderConfig,
		source:       source,
	}
}

var _ gojsonschema.JSONLoader = &Loader{}

func (l Loader) JsonSource() interface{} {
	return l.source
}

func (l *Loader) LoadJSON() (interface{}, error) {
	var err error

	reference, err := l.JsonReference()
	if err != nil {
		return nil, err
	}

	refURL := reference.GetUrl()
	var schemaJSONBytes []byte
	switch refURL.Scheme {
	case "local":
		schemaJSONBytes, err = l.loadLocalReference(refURL)
	case "blueprint":
		schemaJSONBytes, err = l.loadBlueprintReference(refURL)
	default:
		if l.DefaultLoader == nil {
			return nil, fmt.Errorf("unsupported ref %s", refURL.String())
		}
		return l.DefaultLoader.LoaderFactory().New(l.source).LoadJSON()
	}
	if err != nil {
		return nil, err
	}

	if err := ValidateSchema(schemaJSONBytes); err != nil {
		return nil, err
	}

	var schemaJSON interface{}
	if err := yaml.Unmarshal(schemaJSONBytes, &schemaJSON); err != nil {
		return nil, err
	}
	return schemaJSON, nil
}

func (l Loader) JsonReference() (gojsonreference.JsonReference, error) {
	return gojsonreference.NewJsonReference(l.JsonSource().(string))
}

func (l Loader) LoaderFactory() gojsonschema.JSONLoaderFactory {
	return &LoaderFactory{
		LoaderConfig: l.LoaderConfig,
	}
}

func (l *Loader) loadLocalReference(refURL *url.URL) ([]byte, error) {
	if len(refURL.Path) != 0 {
		return nil, errors.New("a path is not supported for local resources")
	}
	schemaBytes, ok := l.LocalTypes[refURL.Host]
	if !ok {
		return nil, fmt.Errorf("type %s is not defined in local types", refURL.Host)
	}
	return schemaBytes, nil
}

func (l *Loader) loadBlueprintReference(refURL *url.URL) ([]byte, error) {
	if l.BlueprintFs == nil {
		return nil, errors.New("no filesystem defined to read a local schema")
	}
	filePath := filepath.Join(refURL.Host, refURL.Path)
	schemaBytes, err := vfs.ReadFile(l.BlueprintFs, filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to read local schema from %s: %w", filePath, err)
	}
	return schemaBytes, nil
}