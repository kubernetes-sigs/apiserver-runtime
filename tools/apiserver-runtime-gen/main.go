package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var bin, output string

var cmd = cobra.Command{
	Use:     "apiserver-runtime-gen",
	Short:   "run code generators",
	PreRunE: preRunE,
	RunE:    runE,
}

func preRunE(cmd *cobra.Command, args []string) error {
	if module == "" {
		return fmt.Errorf("must specify module")
	}
	if len(versions) == 0 {
		return fmt.Errorf("must specify versions")
	}
	return nil
}

func runE(cmd *cobra.Command, args []string) error {
	var err error

	// get the location the generators are installed
	bin = os.Getenv("GOBIN")
	if bin == "" {
		bin = filepath.Join(os.Getenv("HOME"), "go", "bin")
	}
	// install the generators
	for _, gen := range generators {
		// nolint:gosec
		err := run(exec.Command("go", "get", path.Join("k8s.io/code-generator/cmd", gen)))
		if err != nil {
			return err
		}
	}

	// setup the directory to generate the code to.
	// code generators don't work with go modules, and try the full path of the module
	output, err = ioutil.TempDir("", "gen")
	if err != nil {
		return err
	}
	if clean {
		// nolint:errcheck
		defer os.RemoveAll(output)
	}
	d, l := path.Split(module)                   // split the directory from the link we will create
	p := filepath.Join(strings.Split(d, "/")...) // convert module path to os filepath
	p = filepath.Join(output, p)                 // create the directory which will contain the link
	err = os.MkdirAll(p, 0700)
	if err != nil {
		return err
	}
	// link the tmp location to this one so the code generator writes to the correct path
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	err = os.Symlink(wd, filepath.Join(p, l))
	if err != nil {
		return err
	}

	return doGen()
}

func doGen() error {
	inputs := strings.Join(versions, ",")

	gen := map[string]bool{}
	for _, g := range generators {
		gen[g] = true
	}

	if gen["deepcopy-gen"] {
		err := run(getCmd("deepcopy-gen",
			"--input-dirs", inputs,
			"-O", "zz_generated.deepcopy",
			"--bounding-dirs", path.Join(module, "pkg/apis")))
		if err != nil {
			return err
		}
	}

	if gen["openapi-gen"] {
		err := run(getCmd("openapi-gen",
			"--input-dirs", "k8s.io/apimachinery/pkg/api/resource,"+
				"k8s.io/apimachinery/pkg/apis/meta/v1,"+
				"k8s.io/apimachinery/pkg/runtime,"+
				"k8s.io/apimachinery/pkg/version,"+
				inputs,
			"-O", "zz_generated.openapi",
			"--output-package", path.Join(module, "pkg/generated/openapi")))
		if err != nil {
			return err
		}
	}

	if gen["client-gen"] {
		err := run(getCmd("client-gen",
			"--clientset-name", "versioned", "--input-base", "",
			"--input", inputs, "--output-package", path.Join(module, "pkg/generated/clientset")))
		if err != nil {
			return err
		}
	}

	if gen["lister-gen"] {
		err := run(getCmd("lister-gen",
			"--input-dirs", inputs, "--output-package", path.Join(module, "pkg/generated/listers")))
		if err != nil {
			return err
		}
	}

	if gen["informer-gen"] {
		err := run(getCmd("informer-gen",
			"--input-dirs", inputs,
			"--versioned-clientset-package", path.Join(module, "pkg/generated/clientset/versioned"),
			"--listers-package", path.Join(module, "pkg/generated/listers"),
			"--output-package", path.Join(module, "pkg/generated/informers")))
		if err != nil {
			return err
		}
	}

	return nil
}

var (
	generators []string
	header     string
	module     string
	versions   []string
	clean      bool
)

func main() {
	cmd.Flags().BoolVar(&clean, "clean", true, "Delete temporary directory for code generation.")

	defaultGen := []string{"client-gen", "deepcopy-gen", "informer-gen", "lister-gen", "openapi-gen"}
	cmd.Flags().StringSliceVar(&generators, "generators",
		defaultGen, "Code generators to install and run.")
	defaultBoilerplate := filepath.Join("hack", "boilerplate.go.txt")
	cmd.Flags().StringVar(&header, "go-header-file", defaultBoilerplate,
		"File containing boilerplate header text. The string YEAR will be replaced with the current 4-digit year.")

	var defaultModule string
	why := exec.Command("go", "mod", "why")
	why.Stderr = os.Stderr
	if m, err := why.Output(); err == nil {
		parts := strings.Split(string(m), "\n")
		if len(parts) > 1 {
			defaultModule = parts[1]
		}
	} else {
		fmt.Fprintf(os.Stderr, "cannot parse go module: %v\n", err)
	}
	cmd.Flags().StringVar(&module, "module", defaultModule, "Go module of the apiserver.")

	// calculate the versions
	var defaultVersions []string
	if files, err := ioutil.ReadDir(filepath.Join("pkg", "apis")); err == nil {
		for _, f := range files {
			if f.IsDir() {
				versionFiles, err := ioutil.ReadDir(filepath.Join("pkg", "apis", f.Name()))
				if err != nil {
					log.Fatal(err)
				}
				for _, v := range versionFiles {
					if v.IsDir() {
						match := versionRegexp.MatchString(v.Name())
						if !match {
							continue
						}
						defaultVersions = append(defaultVersions, path.Join(module, "pkg", "apis", f.Name(), v.Name()))
					}
				}
			}
		}
	} else {
		fmt.Fprintf(os.Stderr, "cannot parse api versions: %v\n", err)
	}
	cmd.Flags().StringSliceVar(
		&versions, "versions", defaultVersions, "Go packages of API versions to generate code for.")

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

var versionRegexp = regexp.MustCompile("^v[0-9]+((alpha|beta)?[0-9]+)?$")

func run(cmd *exec.Cmd) error {
	fmt.Println(strings.Join(cmd.Args, " "))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func getCmd(cmd string, args ...string) *exec.Cmd {
	// nolint:gosec
	e := exec.Command(filepath.Join(bin, cmd), "--output-base", output, "--go-header-file", header)

	e.Args = append(e.Args, args...)
	return e
}
