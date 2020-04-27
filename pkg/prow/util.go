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
	"strings"

	corev1 "k8s.io/api/core/v1"
	prowv1 "k8s.io/test-infra/prow/apis/prowjobs/v1"
	prowapi "k8s.io/test-infra/prow/config"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/clarketm/pj/api"
	"github.com/clarketm/pj/pkg/maps"
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
		Annotations:     job.Annotations,
		Hidden:          mods.Has(string(api.Private)),
		ReporterConfig:  job.ReporterConfig,
		RerunAuthConfig: job.RerunAuthConfig,
		UtilityConfig: prowapi.UtilityConfig{
			Decorate:         true, // TODO
			PathAlias:        maps.GetOrDefault(job.Aliases, job.Org(), ""),
			CloneURI:         ResolveTemplate(job.CloneTemplate, job),
			SkipSubmodules:   true, // TODO
			CloneDepth:       0,    // TODO
			ExtraRefs:        createExtraRefs(job.ExtraRepos),
			DecorationConfig: job.DecorationConfig,
		},
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

func comparator(order api.SortOrder) func(a, b string) bool {
	switch order {
	case api.Descending:
		return func(a, b string) bool {
			return a < b
		}
	case api.Ascending:
		fallthrough
	default:
		return func(a, b string) bool {
			return a < b
		}
	}
}
