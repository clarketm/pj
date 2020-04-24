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
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/Masterminds/sprig"
	corev1 "k8s.io/api/core/v1"
	prowv1 "k8s.io/test-infra/prow/apis/prowjobs/v1"
	prowapi "k8s.io/test-infra/prow/config"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/clarketm/pjcli/api"
)

func createJobBase(job *api.Job, mods sets.String) prowapi.JobBase {
	return prowapi.JobBase{
		Name:           job.Name,
		Labels:         job.Labels,
		MaxConcurrency: job.MaxConcurrency,
		Cluster:        job.ClusterName,
		Namespace:      &job.Namespace,
		Spec: &corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Image:           job.Image,
					Command:         job.Command,
					Env:             job.Env,
					Resources:       job.JobCore.Resources,
					VolumeMounts:    job.VolumeMounts,
					SecurityContext: job.Container.SecurityContext,
				},
			},
			Volumes:      job.Volumes,
			NodeSelector: job.NodeSelector,
		},
		Annotations: job.Annotations,
		Hidden:      mods.Has(api.Private),
		//ReporterConfig:  nil, // TODO
		//RerunAuthConfig: nil, // TODO
		UtilityConfig: prowapi.UtilityConfig{
			Decorate:       true, // TODO
			PathAlias:      getOrDefault(job.Aliases, job.Org(), ""),
			CloneURI:       ResolveTemplate(job.CloneTemplate, job),
			SkipSubmodules: true, // TODO
			CloneDepth:     0,
			ExtraRefs:      createExtraRefs(job.ExtraRepos),
			//DecorationConfig: &job.DecorationConfig, // TODO
		},
	}
}

func getOrDefault(m map[string]string, key string, def string) string {
	if v, exists := m[key]; !exists {
		return def
	} else {
		return v
	}
}

func createExtraRefs(refs []string) []prowv1.Refs {
	var extraRefs []prowv1.Refs

	for _, ref := range refs {
		var branch = "master" // TODO constant

		orgrepobranch := strings.Split(ref, "@")
		if len(orgrepobranch) > 1 {
			branch = orgrepobranch[1]
		}

		orgrepo := strings.Split(orgrepobranch[0], "/")
		org := orgrepo[0]
		repo := orgrepo[1]

		extraRefs = append(extraRefs, prowv1.Refs{
			Org:     org,
			Repo:    repo,
			BaseRef: branch,
		})
	}

	return extraRefs
}

func jobModifiers(modifiers []api.Modifier) sets.String {
	mods := sets.String{}
	for _, mod := range modifiers {
		mods.Insert(string(mod))
	}
	return mods
}

func ResolveTemplate(tmplStr string, job *api.Job) string {
	if tmplStr == "" {
		return tmplStr
	}

	var b bytes.Buffer

	tmpl, err := template.New(job.Name).Funcs(sprig.FuncMap()).Parse(tmplStr)
	if err != nil {
		fmt.Println(err)
		return tmplStr
	}

	err = tmpl.Execute(&b, struct {
		Org  string
		Repo string
	}{
		Org:  job.Org(),
		Repo: job.Repo(),
	})
	if err != nil {
		fmt.Println(err)
		return tmplStr
	}

	return b.String()
}

func CreatePresubmit(job *api.Job) prowapi.Presubmit {
	mods := jobModifiers(job.Modifiers)

	return prowapi.Presubmit{
		JobBase:      createJobBase(job, mods),
		AlwaysRun:    !mods.Has(api.Skipped),
		Optional:     mods.Has(api.Optional),
		Trigger:      job.Trigger,
		RerunCommand: job.RerunCommand,
		Brancher: prowapi.Brancher{
			SkipBranches: job.SkipBranches,
			Branches:     job.Branches,
		},
		RegexpChangeMatcher: prowapi.RegexpChangeMatcher{
			RunIfChanged: job.Regex,
		},
		Reporter: prowapi.Reporter{
			SkipReport: mods.Has(api.Hidden),
		},
	}
}

func CreatePostsubmit(job *api.Job) prowapi.Postsubmit {
	mods := jobModifiers(job.Modifiers)

	return prowapi.Postsubmit{
		JobBase: createJobBase(job, mods),
		Brancher: prowapi.Brancher{
			SkipBranches: job.SkipBranches,
			Branches:     job.Branches,
		},
		RegexpChangeMatcher: prowapi.RegexpChangeMatcher{
			RunIfChanged: job.Regex,
		},
		Reporter: prowapi.Reporter{
			SkipReport: mods.Has(api.Hidden),
		},
	}
}

func CreatePeriodic(job *api.Job) prowapi.Periodic {
	mods := jobModifiers(job.Modifiers)

	return prowapi.Periodic{
		JobBase:  createJobBase(job, mods),
		Interval: job.Interval,
		Cron:     job.Cron,
	}
}

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

func (o *ProwJobConfig) AddPresubmit(orgrepo string, job *api.Job) {
	o.Presubmits[orgrepo] = append(o.Presubmits[orgrepo], CreatePresubmit(job))
}

func (o *ProwJobConfig) AddPostsubmit(orgrepo string, job *api.Job) {
	o.Postsubmits[orgrepo] = append(o.Postsubmits[orgrepo], CreatePostsubmit(job))
}

func (o *ProwJobConfig) AddPeriodic(job *api.Job) {
	o.Periodics = append(o.Periodics, CreatePeriodic(job))
}
