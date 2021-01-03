package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	yaml "gopkg.in/yaml.v2"
)

func helperLoadConfig(contents []byte) (*Config, error) {
	cfg := &Config{}
	err := yaml.Unmarshal(contents, cfg)
	return cfg, err
}

func TestLoadFile(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		file string
		cfg  Config
		ok   bool
	}{
		{
			file: "../example.tfnotify.yaml",
			cfg: Config{
				CI: "circleci",
				Notifier: Notifier{
					Github: GithubNotifier{
						Token: "$GITHUB_TOKEN",
						Repository: Repository{
							Owner: "mercari",
							Name:  "tfnotify",
						},
					},
				},
				Terraform: Terraform{
					Default: Default{
						Template: "",
					},
					Fmt: Fmt{
						Template: "",
					},
					Plan: Plan{
						Template:    "{{ .Title }}\n{{ .Message }}\n{{if .Result}}\n<pre><code>{{ .Result }}\n</pre></code>\n{{end}}\n<details><summary>Details (Click me)</summary>\n\n<pre><code>{{ .Body }}\n</pre></code></details>\n",
						WhenDestroy: WhenDestroy{},
					},
					Apply: Apply{
						Template: "",
					},
					UseRawOutput: false,
				},
				path: "../example.tfnotify.yaml",
			},
			ok: true,
		},
		{
			file: "../example-with-destroy-and-result-labels.tfnotify.yaml",
			cfg: Config{
				CI: "circleci",
				Notifier: Notifier{
					Github: GithubNotifier{
						Token: "$GITHUB_TOKEN",
						Repository: Repository{
							Owner: "mercari",
							Name:  "tfnotify",
						},
					},
				},
				Terraform: Terraform{
					Default: Default{
						Template: "",
					},
					Fmt: Fmt{
						Template: "",
					},
					Plan: Plan{
						Template: "{{ .Title }}\n{{ .Message }}\n{{if .Result}}\n<pre><code>{{ .Result }}\n</pre></code>\n{{end}}\n<details><summary>Details (Click me)</summary>\n\n<pre><code>{{ .Body }}\n</pre></code></details>\n",
						WhenAddOrUpdateOnly: WhenAddOrUpdateOnly{
							Label: "add-or-update",
						},
						WhenDestroy: WhenDestroy{
							Label:    "destroy",
							Template: "## :warning: WARNING: Resource Deletion will happen :warning:\n\nThis plan contains **resource deletion**. Please check the plan result very carefully!\n",
						},
						WhenPlanError: WhenPlanError{
							Label: "error",
						},
						WhenNoChanges: WhenNoChanges{
							Label: "no-changes",
						},
					},
					Apply: Apply{
						Template: "",
					},
					UseRawOutput: false,
				},
				path: "../example-with-destroy-and-result-labels.tfnotify.yaml",
			},
			ok: true,
		},
		{
			file: "no-such-config.yaml",
			cfg: Config{
				CI: "circleci",
				Notifier: Notifier{
					Github: GithubNotifier{
						Token: "$GITHUB_TOKEN",
						Repository: Repository{
							Owner: "mercari",
							Name:  "tfnotify",
						},
					},
				},
				Terraform: Terraform{
					Default: Default{
						Template: "",
					},
					Fmt: Fmt{
						Template: "",
					},
					Plan: Plan{
						Template:    "{{ .Title }}\n{{ .Message }}\n{{if .Result}}\n<pre><code>{{ .Result }}\n</pre></code>\n{{end}}\n<details><summary>Details (Click me)</summary>\n\n<pre><code>{{ .Body }}\n</pre></code></details>\n",
						WhenDestroy: WhenDestroy{},
					},
					Apply: Apply{
						Template: "",
					},
				},
				path: "no-such-config.yaml",
			},
			ok: false,
		},
	}

	for _, testCase := range testCases {
		var cfg Config

		err := cfg.LoadFile(testCase.file)
		if err == nil {
			if !testCase.ok {
				t.Error("got no error but want error")
			} else if !reflect.DeepEqual(cfg, testCase.cfg) {
				t.Errorf("got %#v but want: %#v", cfg, testCase.cfg)
			}
		} else {
			if testCase.ok {
				t.Errorf("got error %q but want no error", err)
			}
		}
	}
}

func TestValidation(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		contents []byte
		expected string
	}{
		{
			contents: []byte(""),
			expected: "ci: need to be set",
		},
		{
			contents: []byte("ci: rare-ci\n"),
			expected: "rare-ci: not supported yet",
		},
		{
			contents: []byte("ci: circleci\n"),
			expected: "notifier is missing",
		},
		{
			contents: []byte("ci: codebuild\n"),
			expected: "notifier is missing",
		},
		{
			contents: []byte("ci: cloudbuild\n"),
			expected: "notifier is missing",
		},
		{
			contents: []byte("ci: cloud-build\n"),
			expected: "notifier is missing",
		},
		{
			contents: []byte("ci: circleci\nnotifier:\n  github:\n"),
			expected: "notifier is missing",
		},
		{
			contents: []byte("ci: circleci\nnotifier:\n  github:\n    token: token\n"),
			expected: "repository owner is missing",
		},
		{
			contents: []byte(`
ci: circleci
notifier:
  github:
    token: token
    repository:
      owner: owner
`),
			expected: "repository name is missing",
		},
		{
			contents: []byte(`
ci: circleci
notifier:
  github:
    token: token
    repository:
      owner: owner
      name: name
`),
			expected: "",
		},
	}
	for _, testCase := range testCases {
		cfg, err := helperLoadConfig(testCase.contents)
		if err != nil {
			t.Fatal(err)
		}
		err = cfg.Validation()
		if err == nil {
			if testCase.expected != "" {
				t.Errorf("got no error but want %q", testCase.expected)
			}
		} else {
			if err.Error() != testCase.expected {
				t.Errorf("got %q but want %q", err.Error(), testCase.expected)
			}
		}
	}
}

func TestGetNotifierType(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		contents []byte
		expected string
	}{
		{
			contents: []byte("repository:\n  owner: a\n  name: b\nci: circleci\nnotifier:\n  github:\n    token: token\n"),
			expected: "github",
		},
	}
	for _, testCase := range testCases {
		cfg, err := helperLoadConfig(testCase.contents)
		if err != nil {
			t.Fatal(err)
		}
		actual := cfg.GetNotifierType()
		if actual != testCase.expected {
			t.Errorf("got %q but want %q", actual, testCase.expected)
		}
	}
}

func createDummy(file string) {
	validConfig := func(file string) bool {
		for _, c := range []string{
			"tfnotify.yaml",
			"tfnotify.yml",
			".tfnotify.yaml",
			".tfnotify.yml",
		} {
			if file == c {
				return true
			}
		}
		return false
	}
	if !validConfig(file) {
		return
	}
	if _, err := os.Stat(file); err == nil {
		return
	}
	f, err := os.OpenFile(file, os.O_RDONLY|os.O_CREATE, 0o666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
}

func removeDummy(file string) {
	os.Remove(file)
}

func TestFind(t *testing.T) { //nolint:paralleltest
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	testCases := []struct {
		file   string
		expect string
		ok     bool
	}{
		{
			// valid config
			file:   ".tfnotify.yaml",
			expect: ".tfnotify.yaml",
			ok:     true,
		},
		{
			// valid config
			file:   "tfnotify.yaml",
			expect: "tfnotify.yaml",
			ok:     true,
		},
		{
			// valid config
			file:   ".tfnotify.yml",
			expect: ".tfnotify.yml",
			ok:     true,
		},
		{
			// valid config
			file:   "tfnotify.yml",
			expect: "tfnotify.yml",
			ok:     true,
		},
		{
			// invalid config
			file:   "codecov.yml",
			expect: "",
			ok:     false,
		},
		{
			// in case of no args passed
			file:   "",
			expect: filepath.Join(wd, "tfnotify.yaml"),
			ok:     true,
		},
	}
	var cfg Config
	for _, testCase := range testCases { //nolint:paralleltest
		testCase := testCase
		t.Run(testCase.file, func(t *testing.T) {
			createDummy(testCase.file)
			actual, err := cfg.Find(testCase.file)
			if (err == nil) != testCase.ok {
				t.Errorf("got error %q", err)
			}
			if actual != testCase.expect {
				t.Errorf("got %q but want %q", actual, testCase.expect)
			}
		})
		defer removeDummy(testCase.file)
	}
}
