package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dreamsxin/go-kitcli/generator"
	mstrings "github.com/dreamsxin/go-kitcli/generator/strings"
	"github.com/dreamsxin/go-kitcli/generator/template"
	lg "github.com/dreamsxin/go-kitcli/logger"
	"github.com/vetcher/go-astra"
	"github.com/vetcher/go-astra/types"
)

const (
	Version = generator.Version
)

var (
	flagFileName     = flag.String("file", "", "Path to input file with interface.")
	flagPbGoFileName = flag.String("pb-go", "", "Path to XXX_service.pb.go file with protobuf implementation of interface structs.")
	flagOutputDir    = flag.String("out", "", "Output directory.")
	flagPackageName  = flag.String("package", "", "Package name for imports")
	flagHelp         = flag.Bool("help", false, "Show help.")
	flagVerbose      = flag.Int("v", 1, "Sets microgen verbose level.")
	flagDebug        = flag.Bool("debug", false, "Print all microgen messages. Equivalent to -v=100.")
	flagGenProtofile = flag.String(".proto", "", "Package field in protobuf file. If not empty, service.proto file will be generated.")
	flagGenMain      = flag.Bool(generator.MainTag, false, "Generate main.go file.")
)

func init() {
	flag.Parse()
}

func readFromInput(prefix string, delim byte) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prefix)
	input, err := reader.ReadString(delim)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(input, "\n \t\r\f\v"), nil
}

func findPackageNameFromGoModFile(filePath string) (string, error) {
	buffer, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	r, err := regexp.Compile(`module\s+(.*)`)
	if err != nil {
		return "", err
	}

	result := r.FindStringSubmatch(string(buffer))
	if len(result) > 0 {
		return result[1], nil
	}

	return "", errors.New("could not find package name")
}

const (
	goModFileName = "go.mod"
)

func main() {
	lg.Logger.Level = *flagVerbose
	if *flagDebug {
		lg.Logger.Level = lg.DEBUG
	}
	lg.Logger.Logln(1, "@microgen", Version)
	if *flagHelp {
		flag.Usage()
		os.Exit(0)
	}

	if *flagFileName == "" {
		val, err := readFromInput("file path with interfaces: ", '\n')
		if err != nil {
			lg.Logger.Logln(lg.ERROR, "fatal:", err)
			os.Exit(1)
		}
		if val == "" {
			lg.Logger.Logln(lg.ERROR, "fatal:", "interfaces file path required")
			flag.Usage()
			os.Exit(0)
		}

		*flagFileName = val
	}
	if *flagOutputDir == "" {
		defaultDir := filepath.Dir(*flagFileName)
		*flagOutputDir = defaultDir
		printLine := fmt.Sprintf("output directory [%v]: ", defaultDir)
		val, err := readFromInput(printLine, '\n')
		if err != nil {
			lg.Logger.Logln(lg.ERROR, "fatal:", err)
			os.Exit(1)
		}
		if val != "" {
			*flagOutputDir = val
		}
		if *flagOutputDir == "" {
			lg.Logger.Logln(lg.ERROR, "fatal:", "output directory required")
			flag.Usage()
			os.Exit(0)
		}
	}
	if *flagPackageName == "" {
		goModFilePath := filepath.Join(*flagOutputDir, goModFileName)
		defaultPackageName, _ := findPackageNameFromGoModFile(goModFilePath)
		*flagPackageName = defaultPackageName
		printLine := fmt.Sprintf("pacakge name for imports [%v]: ", defaultPackageName)
		val, err := readFromInput(printLine, '\n')
		if err != nil {
			lg.Logger.Logln(lg.ERROR, "fatal:", err)
			os.Exit(1)
		}
		if val != "" {
			*flagPackageName = val
		}
		if *flagPackageName == "" {
			lg.Logger.Logln(lg.ERROR, "fatal:", "package name required")
			flag.Usage()
			os.Exit(0)
		}
	}

	if *flagPbGoFileName == "" {
		val, err := readFromInput("path to XXX_service.pb.go (leave empty for no pb validation): ", '\n')
		if err != nil {
			lg.Logger.Logln(lg.ERROR, "fatal:", err)
			os.Exit(1)
		}

		*flagPbGoFileName = val
	}

	lg.Logger.Logln(4, "Source file:", *flagFileName)
	info, err := astra.ParseFile(*flagFileName)
	if err != nil {
		lg.Logger.Logln(lg.ERROR, "fatal:", err)
		os.Exit(1)
	}
	var pbGoFile *types.File = nil
	if *flagPbGoFileName != "" {
		pbGoFile, err = astra.ParseFile(*flagPbGoFileName)
		if err != nil {
			lg.Logger.Logln(lg.ERROR, "fatal:", err)
			os.Exit(1)
		}
	}

	i := findInterface(info)
	if i == nil {
		lg.Logger.Logln(lg.ERROR, "fatal: could not find interface with @microgen tag")
		lg.Logger.Logln(lg.DEBUG, "All founded interfaces:")
		lg.Logger.Logln(lg.DEBUG, listInterfaces(info.Interfaces))
		os.Exit(1)
	}

	if err := generator.ValidateInterface(i, pbGoFile); err != nil {
		lg.Logger.Logln(lg.ERROR, "validation:", err)
		os.Exit(1)
	}

	ctx, err := prepareContext(*flagFileName, i)
	if err != nil {
		lg.Logger.Logln(lg.ERROR, "fatal:", err)
		os.Exit(1)
	}

	absOutputDir, err := filepath.Abs(*flagOutputDir)
	if err != nil {
		lg.Logger.Logln(lg.ERROR, "fatal:", err)
		os.Exit(1)
	}
	units, err := generator.ListTemplatesForGen(ctx, i, absOutputDir, *flagFileName, *flagPackageName, *flagGenProtofile, *flagGenMain)
	if err != nil {
		lg.Logger.Logln(lg.ERROR, "fatal:", err)
		os.Exit(1)
	}
	for _, unit := range units {
		err := unit.Generate(ctx)
		if err != nil && err != generator.EmptyStrategyError {
			lg.Logger.Logln(lg.ERROR, "fatal:", unit.Path(), err)
			os.Exit(1)
		}
	}
	lg.Logger.Logln(1, "all files successfully generated")
}

func listInterfaces(ii []types.Interface) string {
	var s string
	for _, i := range ii {
		s = s + fmt.Sprintf("\t%s(%d methods, %d embedded interfaces)\n", i.Name, len(i.Methods), len(i.Interfaces))
	}
	return s
}

func prepareContext(filename string, iface *types.Interface) (context.Context, error) {
	ctx := context.Background()
	ctx = template.WithSourcePackageImport(ctx, filename)

	set := template.TagsSet{}
	genTags := mstrings.FetchTags(iface.Docs, generator.TagMark+generator.MicrogenMainTag)
	for _, tag := range genTags {
		set.Add(tag)
	}
	ctx = template.WithTags(ctx, set)
	return ctx, nil
}

func findInterface(file *types.File) *types.Interface {
	for i := range file.Interfaces {
		if docsContainMicrogenTag(file.Interfaces[i].Docs) {
			return &file.Interfaces[i]
		}
	}
	return nil
}

func docsContainMicrogenTag(strs []string) bool {
	for _, str := range strs {
		if strings.HasPrefix(str, generator.TagMark+generator.MicrogenMainTag) {
			return true
		}
	}
	return false
}
