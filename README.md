# pj

pj is a command line tool for managing ProwJobs (PJs).

## Installing

## Getting Started

## Subcommands

### `create`

Create ProwJob yaml configuration.

#### `-g, --global <file1,file2,...>`

Global configuration files containing *default* values inherited by all jobs.

```yaml
# All valid `metav1.ObjectMeta`, `corev1.Container`, and `corev1.PodSpec` fields are accepted.
labels: {app.kubernetes.io/part-of: prow}
namespace: test-pods
resources: {requests: {cpu: 1}, limits: {cpu: 3}}
aliases: {istio: istio.io}
clusterName: default

# Requirements are a `map[string]Job` that a job can `require` as part of its definition. 
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
# Local defaults inherited by all the below jobs.
repo: istio/istio
nodeSelector: {testing: test-pool}
clone_tmpl: "https://github.com/{{.Org}}/{{.Repo}}.git"
output_tmpl: "{{.Org}}/{{.Repo}}/{{.Org}}.{{.Repo}}.gen"

jobs:
- name: job_1
  types: [presubmit]
  require: [gcp]  # a corresponding `requirements` item must exist.
  labels: {app.kubernetes.io/version: 1.0.0}
  branches: [master]
  image: alpine:latest
  command: [echo, test1]
  extra_repos: [istio/test-infra@master]
```

#### `-o, --ouput <directory>`

Output directory to write jobs to. Subdirectory structure is determined by the `output_tmpl` field. 
