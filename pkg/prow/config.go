/*
 Copyright © 2020 Travis Clarke <travis.m.clarke@gmail.com>

 Permission is hereby granted, free of charge, to any person obtaining a copy
 of this software and associated documentation files (the "Software"), to deal
 in the Software without restriction, including without limitation the rights
 to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 copies of the Software, and to permit persons to whom the Software is
 furnished to do so, subject to the following conditions:

 The above copyright notice and this permission notice shall be included in
 all copies or substantial portions of the Software.

 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 THE SOFTWARE.
*/

package prow

import (
	"sort"

	prowapi "k8s.io/test-infra/prow/config"

	"github.com/clarketm/pj/api"
)

type ProwJobConfig struct {
	Presubmits  map[string][]prowapi.Presubmit
	Postsubmits map[string][]prowapi.Postsubmit
	Periodics   []prowapi.Periodic
}

func NewProwJobConfig() *ProwJobConfig {
	var pjc ProwJobConfig
	pjc.Presubmits = make(map[string][]prowapi.Presubmit)
	pjc.Postsubmits = make(map[string][]prowapi.Postsubmit)
	return &pjc
}

func (o *ProwJobConfig) Empty() bool {
	return len(o.Presubmits) == 0 && len(o.Postsubmits) == 0 && len(o.Periodics) == 0
}

func (o *ProwJobConfig) AddPresubmit(orgrepo string, job *api.Job) {
	o.Presubmits[orgrepo] = append(o.Presubmits[orgrepo], CreatePresubmit(job))
}

func (o *ProwJobConfig) AddPostsubmit(orgrepo string, job *api.Job) {
	o.Postsubmits[orgrepo] = append(o.Postsubmits[orgrepo], CreatePostsubmit(job))
}

func (o *ProwJobConfig) AddPeriodic(job *api.Job) {
	o.Periodics = append(o.Periodics, CreatePeriodic(job))
}

func (o *ProwJobConfig) SortPresubmit(order api.SortOrder) {
	for _, c := range o.Presubmits {
		sort.Slice(c, func(a, b int) bool {
			return comparator(order)(c[a].Name, c[b].Name)
		})
	}
}

func (o *ProwJobConfig) SortPostsubmit(order api.SortOrder) {
	for _, c := range o.Postsubmits {
		sort.Slice(c, func(a, b int) bool {
			return comparator(order)(c[a].Name, c[b].Name)
		})
	}
}

func (o *ProwJobConfig) SortPeriodic(order api.SortOrder) {
	sort.Slice(o.Periodics, func(a, b int) bool {
		return comparator(order)(o.Periodics[a].Name, o.Periodics[b].Name)
	})
}
