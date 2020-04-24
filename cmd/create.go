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

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/imdario/mergo"
	"github.com/spf13/cobra"
	prowapi "k8s.io/test-infra/prow/config"
	"sigs.k8s.io/yaml"

	"github.com/clarketm/pjcli/api"
	osutil "github.com/clarketm/pjcli/pkg/os"
	"github.com/clarketm/pjcli/pkg/prow"
)

var createShort = "A brief description of your command"

var createLong = `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: createShort,
	Long:  createLong,
	Run:   create,
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringSliceP("global", "g", []string{}, "Global configuration files.")
	createCmd.Flags().StringSliceP("input", "i", []string{"/dev/stdin"}, "Input files and/or directories.")
	createCmd.Flags().StringP("output", "o", "/dev/stdout", "Output directory.")
}

func create(cmd *cobra.Command, args []string) {
	var globalConfig api.Job
	var err error
	var prowjobs = make(map[string]*prow.ProwJobConfig)

	clean, _ := cmd.Flags().GetBool("clean")
	if clean {
		fmt.Println("\ncleaning...\n")
	}

	global, _ := cmd.Flags().GetStringSlice("global")
	input, _ := cmd.Flags().GetStringSlice("input")
	output, _ := cmd.Flags().GetString("output")

	if output, err = filepath.Abs(output); err != nil {
		fmt.Println(err)
		return
	}

	for i, g := range global {
		if global[i], err = filepath.Abs(g); err != nil {
			continue
		}

		if !osutil.Exists(global[i]) {
			continue
		}

		f, err := ioutil.ReadFile(global[i])
		if err != nil {
			continue
		}

		var gc api.JobConfiguration
		if err := yaml.Unmarshal(f, &gc); err != nil {
			continue
		}

		if err := mergo.Merge(&globalConfig, gc.GlobalDefaults); err != nil {
			continue
		}
	}

	for i, j := range input {
		if input[i], err = filepath.Abs(j); err != nil {
			continue
		}

		if !osutil.Exists(input[i]) {
			continue
		}

		f, err := ioutil.ReadFile(input[i])
		if err != nil {
			continue
		}

		var jc api.JobConfiguration
		if err := yaml.Unmarshal(f, &jc); err != nil {
			continue
		}

		for i := range jc.Jobs {

			for _, m := range []interface{}{jc.LocalDefaults, jc.GlobalDefaults, globalConfig} {
				if err := mergo.Merge(&jc.Jobs[i], m); err != nil {
					fmt.Println(err)
					continue
				}
			}

			job := &jc.Jobs[i]

			for _, req := range job.Require {
				if err := mergo.Merge(job, job.Requirements[req]); err != nil {
					fmt.Println(err)
					continue
				}
			}

			var path = output
			err = os.MkdirAll(output, os.ModePerm)

			if osutil.IsDirectory(output) {
				if job.OutputTemplate != "" {
					tmpl := prow.ResolveTemplate(job.OutputTemplate, job)
					path = filepath.Join(output, tmpl)
					if !osutil.HasExtension(path, "ya?ml") { // TODO constant
						path += ".yaml"
					}
				} else {
					path = filepath.Join(output, "prowjobs.yaml")
				}
			}

			if _, exists := prowjobs[path]; !exists {
				prowjobs[path] = prow.NewProwJobConfig()
			}

			// Default to presubmit
			if len(job.Types) == 0 {
				job.Types = []api.JobType{api.Presubmit}
			}

			for _, jobType := range job.Types {
				switch jobType {
				case api.Postsubmit:
					prowjobs[path].AddPresubmit(job.OrgRepo, job)
				case api.Periodic:
					prowjobs[path].AddPeriodic(job)
				case api.Presubmit:
					prowjobs[path].AddPresubmit(job.OrgRepo, job)
				}
			}
		}
	}

	for path, jobs := range prowjobs {
		jobConfig := prowapi.JobConfig{}

		dir := filepath.Dir(path)
		err = os.MkdirAll(dir, os.ModePerm)

		if err = jobConfig.SetPresubmits(jobs.Presubmits); err != nil {
			fmt.Println(err)
		}

		if err = jobConfig.SetPostsubmits(jobs.Postsubmits); err != nil {
			fmt.Println(err)
		}

		jobConfig.Periodics = jobs.Periodics

		jobConfigYaml, err := yaml.Marshal(jobConfig)
		if err != nil {
			fmt.Println(err)
		}

		if err = ioutil.WriteFile(path, jobConfigYaml, 0644); err != nil {
			fmt.Println(err)
		}
	}
}
