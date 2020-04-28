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

package api

import (
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	prowv1 "k8s.io/test-infra/prow/apis/prowjobs/v1"
)

type Empty *struct{}

type Defaults Job

type JobConfiguration struct {
	Defaults
	Jobs []Job `json:"jobs,omitempty"`
}

type JobCore struct {
	metav1.ObjectMeta
	corev1.Container
	corev1.PodSpec
}

type JobProw struct {
	*prowv1.DecorationConfig `json:"decoration_config,omitempty"`
	*prowv1.RerunAuthConfig  `json:"rerun_auth_config,omitempty"`
	*prowv1.ReporterConfig   `json:"reporter_config,omitempty"`

	Command        []string          `json:"command,omitempty"`
	Branches       []string          `json:"branches,omitempty"`
	SkipBranches   []string          `json:"skip_branches,omitempty"`
	ExtraRepos     []string          `json:"extra_repos,omitempty"`
	Require        []string          `json:"require,omitempty"`
	OrgRepo        string            `json:"repo,omitempty"`
	Name           string            `json:"name,omitempty"`
	CloneTemplate  string            `json:"clone_tmpl,omitempty"`
	OutputTemplate string            `json:"output_tmpl,omitempty"`
	Image          string            `json:"image,omitempty"`
	Regex          string            `json:"regex,omitempty"`
	Trigger        string            `json:"trigger,omitempty"`
	RerunCommand   string            `json:"rerun_command,omitempty"`
	MaxConcurrency int               `json:"max_concurrency,omitempty"`
	Aliases        map[string]string `json:"aliases,omitempty"`
	Requirements   map[string]Job    `json:"requirements,omitempty"`
	Types          []JobType         `json:"types,omitempty"`
	Modifiers      []Modifier        `json:"modifiers,omitempty"`
	Timeout        time.Duration     `json:"timeout,omitempty"`
}

type JobPeriodic struct {
	Cron     string `json:"cron,omitempty"`
	Interval string `json:"interval,omitempty"`
	//Interval *time.Duration `json:"interval,omitempty"`
}

type Job struct {
	JobCore
	JobProw
	JobPeriodic
}

func (j *Job) Org() string {
	return strings.Split(j.OrgRepo, "/")[0]
}

func (j *Job) Repo() string {
	return strings.Split(j.OrgRepo, "/")[1]
}
