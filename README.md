# pj-cli (ProwJob CLI)

Command line tool for working with ProwJobs (PJ)

## Subcommands

### `create`

Create ProwJob yaml configuration.

#### `-g, --global <file1,file2,...>`

Global configuration files containing *default* values inherited by all jobs.

```yaml
global_defaults:
  namespace: test-pods
  resources: {requests: {cpu: 1}, limits: {cpu: 3}}
  aliases: {istio: istio.io}
  clusterName: default
  requirements:
    gcp:
      labels:
        preset-service-account: "true"
    root:
      securityContext:
        privileged: true
```

#### `-i, --input <file1,file2,...>`

Job configuration files with *optional* file-level *defaults*.

```yaml
defaults:
  repo: istio/istio
  nodeSelector: {testing: test-pool}
  clone_tmpl: "https://github.com/{{.Org}}/{{.Repo}}.git"

jobs:
- name: job_1
  types: [presubmit]
  require: [github]
  labels: {app.kubernetes.io/version: 1.0.0}
  branches: [master]
  image: alpine:latest
  command: [echo, test1]
  extra_repos: [istio/test-infra]
  nodeSelector: {testing: build-pool}
  output_tmpl: "{{.Org}}/{{.Repo}}/{{.Org}}.{{.Repo}}.gen"
```

#### `-o, --ouput <directory>`

Output directory to write jobs to. Subdirectory structure is determined by the `output_tmpl` field. 
