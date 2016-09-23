package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
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

func main() {
	// flags
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
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

	// check if file exists
	if _, err := os.Stat(tmplFile); os.IsNotExist(err) {
		log.Fatalf("%s", err)
	}

	// template funcs
	funcMap := make(map[string]interface{})
	funcMap["getenv"] = os.Getenv
	funcMap["getEnv"] = os.Getenv
	funcMap["getFileContent"] = func(i string) string {
		d, err := ioutil.ReadFile(i)
		if err != nil {
			log.Fatalf("%s", err)
		}
		return string(d)
	}
	funcMap["getFileContentBytes"] = func(i string) []byte {
		d, err := ioutil.ReadFile(i)
		if err != nil {
			log.Fatalf("%s", err)
		}
		return d
	}
	funcMap["gzip"] = func(i []byte) []byte {
		var b bytes.Buffer
		w := gzip.NewWriter(&b)
		if _, err := w.Write(i); err != nil {
			log.Fatalf("%s", err)
		}

		if err := w.Close(); err != nil {
			log.Fatalf("%s", err)
		}

		return b.Bytes()
	}
	funcMap["encodeBase64"] = func(i []byte) string {
		return base64.StdEncoding.EncodeToString(i)
	}
	funcMap["toInt"] = func(s string) int {
		i, err := strconv.Atoi(s)
		if err != nil {
			log.Fatalf("%s", err)
		}

		return i
	}
	funcMap["add"] = func(a, b int) int {
		return a + b
	}
	funcMap["sub"] = func(a, b int) int {
		return a - b
	}

	// create template
	tmpl, err := template.New(path.Base(tmplFile)).Funcs(funcMap).ParseFiles(tmplFile)
	if err != nil {
		log.Fatalf("Unable to process template %s, %s", tmplFile, err)
	}

	// and then execute ;)
	if err = tmpl.Execute(os.Stdout, nil); err != nil {
		log.Fatalf("%s", err)
	}
}
