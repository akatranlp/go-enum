//go:generate go run . string EnumType string int
package main

import (
	_ "embed"
	"flag"
	"os"
	"path"
	"runtime/debug"
	"strings"
	"text/template"

	"github.com/akatranlp/go-enum/enumtype"
)

var templateFuncs = template.FuncMap{
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
}

//go:embed enum.str.gen.go.tmpl
var enumStrTemplateFile string
var enumStrTemplate = template.Must(template.New("enum").Funcs(templateFuncs).Parse(enumStrTemplateFile))

//go:embed enum.int.gen.go.tmpl
var enumIntTemplateFile string
var enumIntTemplate = template.Must(template.New("enum").Funcs(templateFuncs).Parse(enumIntTemplateFile))

type TemplateData struct {
	App        string
	Enum       string
	Values     []string
	EmptyValid bool
}

const (
	enumType = iota
	enumName
	enumValues
	argMinCount
)

func main() {
	emptyValid := flag.Bool("empty", false, "allow empty value")
	rootDir := flag.String("dir", ".", "root directory")
	flag.Parse()

	if flag.NArg() < argMinCount {
		panic("not enough arguments")
	}
	args := flag.Args()

	dirInfo, err := os.Stat(*rootDir)
	if err != nil {
		panic(err)
	} else if !dirInfo.IsDir() {
		panic("not a directory")
	}

	enumType, err := enumtype.Parse(args[enumType])
	if err != nil {
		panic(err)
	}

	enumName := args[enumName]
	packageName := strings.ToLower(enumName)
	enumValues := args[enumValues:]

	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		panic("could not read build info")
	}

	appName, _ := strings.CutPrefix(buildInfo.Path, "github.com/")
	origArgs := append([]string{appName}, os.Args[1:]...)

	data := TemplateData{
		App:        strings.Join(origArgs, " "),
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

	switch enumType {
	case enumtype.String:
		if err := enumStrTemplate.Execute(f, data); err != nil {
			panic(err)
		}
	case enumtype.Int:
		if err := enumIntTemplate.Execute(f, data); err != nil {
			panic(err)
		}
	}
}
