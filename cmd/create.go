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
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	prowapi "k8s.io/test-infra/prow/config"
	"sigs.k8s.io/yaml"

	"github.com/hashicorp/go-multierror"

	"github.com/clarketm/pj/api"
	osutil "github.com/clarketm/pj/pkg/os"
	"github.com/clarketm/pj/pkg/prow"
)

var createShort = "Create ProwJob yaml configuration"

var createLong = `Create ProwJob yaml configuration

# Create ProwJobs using short options.
pj create -g ./examples/global1.yaml -i ./examples/jobs.yaml -o ./jobs

# Create ProwJobs using long options.
pj create --global ./examples/global1.yaml --input ./examples/jobs.yaml --output ./jobs

# Create ProwJobs using input from stdin and ouput to stdout.
pj create
`

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: createShort,
	Long:  createLong,
	RunE:  create,
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringSliceP("global", "g", []string{}, "Global configuration files.")
	createCmd.Flags().StringSliceP("input", "i", []string{"/dev/stdin"}, "Input files and/or directories.")
	createCmd.Flags().StringP("output", "o", "/dev/stdout", "Output directory.")
}

func create(cmd *cobra.Command, args []string) error {
	var globalConfig api.Job
	var prowjobs = make(map[string]*prow.ProwJobConfig)

	global, err := cmd.Flags().GetStringSlice("global")
	if err != nil {
		return errors.Wrapf(err, "getting global flag")
	}

	input, err := cmd.Flags().GetStringSlice("input")
	if err != nil {
		return errors.Wrapf(err, "getting input flag")
	}

	output, err := cmd.Flags().GetString("output")
	if err != nil {
		return errors.Wrapf(err, "getting output flag")
	}

	// Process output directory.
	if output, err = filepath.Abs(output); err != nil {
		return errors.Wrapf(err, "getting output path: %s", output)
	}

	if err = os.MkdirAll(output, os.ModePerm); err != nil {
		err = multierror.Append(errors.Wrapf(err, "creating output directory: %s", output))
	}

	// Process global configuration files.
	for i, g := range global {

		if global[i], err = filepath.Abs(g); err != nil {
			err = multierror.Append(errors.Wrapf(err, "getting global path: %s", global[i]))
			continue
		}

		if !osutil.Exists(global[i]) {
			err = multierror.Append(errors.Wrapf(err, "global path exists: %s", global[i]))
			continue
		}

		f, err := ioutil.ReadFile(global[i])
		if err != nil {
			err = multierror.Append(errors.Wrapf(err, "reading global path: %s", global[i]))
			continue
		}

		var gc api.JobConfiguration
		if err := yaml.Unmarshal(f, &gc); err != nil {
			err = multierror.Append(errors.Wrapf(err, "unmarshal global config: %s", global[i]))
			continue
		}

		if err := mergo.Merge(&globalConfig, gc.GlobalDefaults); err != nil {
			err = multierror.Append(errors.Wrapf(err, "merge global config: %s", global[i]))
			continue
		}
	}

	// Process input configuration files and defaults.
	for i, j := range input {
		if input[i], err = filepath.Abs(j); err != nil {
			err = multierror.Append(errors.Wrapf(err, "getting input path: %s", input[i]))
			continue
		}

		if !osutil.Exists(input[i]) {
			err = multierror.Append(errors.Wrapf(err, "input path exists: %s", input[i]))
			continue
		}

		// Process job configuration.
		if err = filepath.Walk(input[i], func(inPath string, info os.FileInfo, err error) error {

			if !osutil.HasExtension(inPath, prow.YamlExt) {
				return nil
			}

			f, err := ioutil.ReadFile(inPath)
			if err != nil {
				err = multierror.Append(errors.Wrapf(err, "reading input path: %s", inPath))
				return nil
			}

			var jc api.JobConfiguration
			if err := yaml.Unmarshal(f, &jc); err != nil {
				err = multierror.Append(errors.Wrapf(err, "unmarshal input config: %s", inPath))
				return nil
			}

			for i := range jc.Jobs {
				job := &jc.Jobs[i]

				for _, m := range []interface{}{jc.LocalDefaults, jc.GlobalDefaults, globalConfig} {
					if err := mergo.Merge(job, m); err != nil {
						err = multierror.Append(errors.Wrapf(err, "merge input config: %s", inPath))
						return nil
					}
				}

				for _, req := range job.Require {
					if err := mergo.Merge(job, job.Requirements[req]); err != nil {
						err = multierror.Append(errors.Wrapf(err, "merge requirement: %s", req))
						return nil
					}
				}

				var outPath = output

				if osutil.IsDirectory(output) {
					if job.OutputTemplate != "" {
						tmpl := prow.ResolveTemplate(job.OutputTemplate, job)
						outPath = filepath.Join(output, tmpl)
					} else {
						outPath = filepath.Join(output, prow.DefaultOutput)
					}
				}

				if _, exists := prowjobs[outPath]; !exists {
					prowjobs[outPath] = prow.NewProwJobConfig()
				}

				// Default to presubmit if unspecified.
				if len(job.Types) == 0 {
					job.Types = []api.JobType{api.Presubmit}
				}

				for _, jobType := range job.Types {
					switch jobType {
					case api.Postsubmit:
						prowjobs[outPath].AddPresubmit(job.OrgRepo, job)
					case api.Periodic:
						prowjobs[outPath].AddPeriodic(job)
					case api.Presubmit:
						fallthrough
					default:
						prowjobs[outPath].AddPresubmit(job.OrgRepo, job)
					}
				}
			}
			return nil
		}); err != nil {
			err = multierror.Append(errors.Wrapf(err, "walking input path: %s", input[i]))
		}
	}

	for path, jobs := range prowjobs {
		if jobs.Empty() {
			continue
		}

		jobConfig := prowapi.JobConfig{}

		if err = jobConfig.SetPresubmits(jobs.Presubmits); err != nil {
			err = multierror.Append(errors.Wrapf(err, "settings presubmits: %s", path))
		}

		if err = jobConfig.SetPostsubmits(jobs.Postsubmits); err != nil {
			err = multierror.Append(errors.Wrapf(err, "settings postsubmits: %s", path))
		}

		jobConfig.Periodics = jobs.Periodics

		jobConfigYaml, err := yaml.Marshal(jobConfig)
		if err != nil {
			err = multierror.Append(errors.Wrapf(err, "marshal job config: %s", path))
			continue
		}

		dir := filepath.Dir(path)
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			err = multierror.Append(errors.Wrapf(err, "creating directory: %s", path))
			continue
		}

		outBytes := []byte(prow.AutogenHeader)
		outBytes = append(outBytes, jobConfigYaml...)

		if err = ioutil.WriteFile(path, outBytes, 0644); err != nil {
			err = multierror.Append(errors.Wrapf(err, "writing job config: %s", path))
		}
	}

	return err
}
