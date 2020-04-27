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

	"github.com/Masterminds/sprig"
	prowapi "k8s.io/test-infra/prow/config"

	"github.com/clarketm/pj/api"
)

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
		AlwaysRun:    !mods.Has(string(api.Skipped)),
		Optional:     mods.Has(string(api.Optional)),
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
			SkipReport: mods.Has(string(api.Hidden)),
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
			SkipReport: mods.Has(string(api.Hidden)),
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
