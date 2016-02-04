package main

import (
	"os"
	"strings"
	"text/template"
	"path"

	flag "github.com/spf13/pflag"
	log "github.com/glerchundi/logrus"
	"fmt"
)

const (
	cliName        = "tmpl-renderer"
	cliDescription = "tmpl-renderer renders a template file"
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
		log.Fatal(err)
	}

	// template funcs
	funcMap := make(map[string]interface{})
	funcMap["getenv"] = os.Getenv

	// create template
	tmpl, err := template.New(path.Base(tmplFile)).Funcs(funcMap).ParseFiles(tmplFile)
	if err != nil {
		log.Fatal(fmt.Errorf("Unable to process template %s, %s", tmplFile, err))
	}

	// and then execute ;)
	if err = tmpl.Execute(os.Stderr, nil); err != nil {
		log.Fatal(err)
	}
}
