package main

import (
	_ "embed"
	"flag"
	"os"
	"path"
	"runtime/debug"
	"strings"
	"text/template"
)

//go:embed enum.gen.go.tmpl
var enumTemplateStr string

var enumTemplate = template.Must(template.New("enum").
	Funcs(template.FuncMap{
		"lower": strings.ToLower,
		"upper": strings.ToUpper,
		"capitalize": func(s string) string {
			if len(s) == 0 {
				return ""
			}
			return strings.ToUpper(s[:1]) + s[1:]
		},
		"combine": func(a, b string) string { return a + b },
		"max":     func(a, b int) int { return max(a, b) },
		"firstChar": func(s string) string {
			if len(s) == 0 {
				return ""
			}
			return s[:1]
		},
	}).Parse(enumTemplateStr))

type TemplateData struct {
	App        string
	Enum       string
	Values     []string
	EmptyValid bool
}

func main() {
	emptyValid := flag.Bool("empty", false, "allow empty value")
	rootDir := flag.String("dir", ".", "root directory")
	flag.Parse()

	if flag.NArg() < 2 {
		panic("not enough arguments")
	}
	args := flag.Args()

	dirInfo, err := os.Stat(*rootDir)
	if err != nil {
		panic(err)
	} else if !dirInfo.IsDir() {
		panic("not a directory")
	}

	enumName := args[0]
	packageName := strings.ToLower(enumName)
	enumValues := args[1:]

	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		panic("could not read build info")
	}
	appName, _ := strings.CutPrefix(buildInfo.Path, "github.com/")

	data := TemplateData{
		App:        appName,
		Enum:       enumName,
		EmptyValid: *emptyValid,
		Values:     enumValues,
	}

	packageDir := path.Join(*rootDir, packageName)
	if err := os.MkdirAll(packageDir, os.ModePerm); err != nil {
		panic(err)
	}
	f, err := os.Create(path.Join(packageDir, packageName+".gen.go"))
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := enumTemplate.Execute(f, data); err != nil {
		panic(err)
	}
}
