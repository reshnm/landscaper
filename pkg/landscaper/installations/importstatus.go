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

package installations

import (
	"fmt"

	lsv1alpha1 "github.com/gardener/landscaper/pkg/apis/core/v1alpha1"
)

// ImportStatus is the internal representation of all import status of a installation.
type ImportStatus struct {
	Data   map[string]*lsv1alpha1.ImportStatus
	Target map[string]*lsv1alpha1.ImportStatus
}

func (s *ImportStatus) set(status lsv1alpha1.ImportStatus) {
	if status.Type == lsv1alpha1.DataImportStatusType {
		s.Data[status.Name] = &status
	}
	if status.Type == lsv1alpha1.TargetImportStatusType {
		s.Target[status.Name] = &status
	}
}

// Updates the internal import states
func (s *ImportStatus) Update(state lsv1alpha1.ImportStatus) {
	s.set(state)
}

// GetStatus returns the import states of the installation.
func (s *ImportStatus) GetStatus() []lsv1alpha1.ImportStatus {
	states := make([]lsv1alpha1.ImportStatus, 0)
	for _, state := range s.Data {
		states = append(states, *state)
	}
	for _, state := range s.Target {
		states = append(states, *state)
	}
	return states
}

// GetData returns the import data status for the given key.
func (s *ImportStatus) GetData(name string) (lsv1alpha1.ImportStatus, error) {
	state, ok := s.Data[name]
	if !ok {
		return lsv1alpha1.ImportStatus{}, fmt.Errorf("import state %s not found", name)
	}
	return *state, nil
}

// GetTarget returns the import target state for the given key.
func (s *ImportStatus) GetTarget(name string) (lsv1alpha1.ImportStatus, error) {
	state, ok := s.Target[name]
	if !ok {
		return lsv1alpha1.ImportStatus{}, fmt.Errorf("import state %s not found", name)
	}
	return *state, nil
}
