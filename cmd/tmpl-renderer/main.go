package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"text/template"

	flag "github.com/spf13/pflag"
)

const (
	cliName = "tmpl-renderer"
)

var (
	Version = "0.0.0"
	GitRev  = "----------------------------------------"
)

type tmplRendererConfig struct {
	outputPath string
}

func newTmplRendererConfig() tmplRendererConfig {
	return tmplRendererConfig{
		outputPath: "",
	}
}

func main() {
	trc := newTmplRendererConfig()

	// flags
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.StringVar(&trc.outputPath, "out", trc.outputPath, "")

	fs.SetNormalizeFunc(
		func(f *flag.FlagSet, name string) flag.NormalizedName {
			if strings.Contains(name, "_") {
				return flag.NormalizedName(strings.Replace(name, "_", "-", -1))
			}
			return flag.NormalizedName(name)
		},
	)

	// define usage
	fs.Usage = func() {
		goVersion := strings.TrimPrefix(runtime.Version(), "go")
		fmt.Fprintf(os.Stderr, "%s (version=%s, gitrev=%s, go=%s)\n", cliName, Version, GitRev, goVersion)
		/*
			fmt.Fprintf(os.Stderr, "Usage:\n")
			fs.PrintDefaults()
		*/
	}

	// parse
	fs.Parse(os.Args[1:])

	// set from env (if present)
	fs.VisitAll(func(f *flag.Flag) {
		if !f.Changed {
			key := strings.ToUpper(strings.Join(
				[]string{
					cliName,
					strings.Replace(f.Name, "-", "_", -1),
				},
				"_",
			))
			val := os.Getenv(key)
			if val != "" {
				fs.Set(f.Name, val)
			}
		}
	})

	// check if template was provided
	args := fs.Args()
	if len(args) < 1 {
		log.Fatal("Please provide a template")
	}

	// template file
	tmplFile := args[0]
	basePath := path.Dir(tmplFile)

	// check if file exists
	if _, err := os.Stat(tmplFile); os.IsNotExist(err) {
		log.Fatalf("%s", err)
	}

	// template funcs
	funcMap := make(map[string]interface{})

	funcMap["getenv"] = os.Getenv

	funcMap["getEnv"] = os.Getenv

	funcMap["encodeJSON"] = func(i interface{}) string {
		d, err := json.Marshal(i)
		if err != nil {
			log.Fatalf("%v", err)
		}
		return string(d)
	}

	funcMap["getFileContent"] = func(i string) string {
		d, err := ioutil.ReadFile(path.Join(basePath, i))
		if err != nil {
			log.Fatalf("%v", err)
		}
		return string(d)
	}

	funcMap["getFileContentBytes"] = func(i string) []byte {
		d, err := ioutil.ReadFile(path.Join(basePath, i))
		if err != nil {
			log.Fatalf("%v", err)
		}
		return d
	}

	funcMap["gzip"] = func(i []byte) []byte {
		var b bytes.Buffer
		w := gzip.NewWriter(&b)
		if _, err := w.Write(i); err != nil {
			log.Fatalf("%v", err)
		}

		if err := w.Close(); err != nil {
			log.Fatalf("%v", err)
		}

		return b.Bytes()
	}

	funcMap["encodeBase64"] = func(i []byte) string {
		return base64.StdEncoding.EncodeToString(i)
	}

	funcMap["toInt"] = func(s string) int {
		i, err := strconv.Atoi(s)
		if err != nil {
			log.Fatalf("%v", err)
		}

		return i
	}

	funcMap["add"] = func(a, b int) int {
		return a + b
	}

	funcMap["sub"] = func(a, b int) int {
		return a - b
	}

	var stdin []byte
	var stdinOnce sync.Once
	funcMap["getStdin"] = func() []byte {
		stdinOnce.Do(func() {
			fi, err := os.Stdin.Stat()
			if err != nil {
				return
			}

			if fi.Mode()&os.ModeNamedPipe == 0 {
				return
			}

			data, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				return
			}

			stdin = data
		})

		return stdin
	}

	funcMap["sopsDecrypt"] = func(i string) []byte {
		cmd := exec.Command("sops", "-d", path.Join(basePath, i))
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			log.Fatalf("%v", err)
		}

		return stdout.Bytes()
	}

	// create template
	tmpl, err := template.New(path.Base(tmplFile)).Funcs(funcMap).ParseFiles(tmplFile)
	if err != nil {
		log.Fatalf("Unable to process template %s, %s", tmplFile, err)
	}

	// choose writer between file and stdout
	var writer io.Writer
	if trc.outputPath != "" {
		f, err := os.Create(trc.outputPath)
		if err != nil {
			log.Fatalf("Unable to create %s: %v", trc.outputPath, err)
		}
		defer f.Close()

		writer = f
	} else {
		writer = os.Stdout
	}

	// and then execute ;)
	if err = tmpl.Execute(writer, nil); err != nil {
		log.Fatalf("%s", err)
	}
}
