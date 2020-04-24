module github.com/clarketm/pjcli

go 1.14

replace (
	k8s.io/client-go => k8s.io/client-go v0.17.3
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.5.0
)

require (
	cloud.google.com/go/storage v1.1.2 // indirect
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/huandu/xstrings v1.3.1 // indirect
	github.com/imdario/mergo v0.3.8
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/spf13/cobra v0.0.6
	github.com/spf13/viper v1.6.2
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/test-infra v0.0.0-20200307225934-f04d2034f147
	sigs.k8s.io/yaml v1.1.0
)
