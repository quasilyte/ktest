package bench

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/quasilyte/ktest/internal/fileutil"
	"github.com/z7zmey/php-parser/pkg/conf"
	"github.com/z7zmey/php-parser/pkg/errors"
	"github.com/z7zmey/php-parser/pkg/parser"
	"github.com/z7zmey/php-parser/pkg/version"
	"github.com/z7zmey/php-parser/pkg/visitor/traverser"
)

type runner struct {
	conf *RunConfig

	benchFiles []*benchFile

	composerMode bool

	buildDir string
}

type benchFile struct {
	id int

	fullName  string
	shortName string

	info *benchParsedInfo

	generatedMain []byte
}

type benchMethod struct {
	Name string
	Key  string
}

type benchParsedInfo struct {
	ClassName    string
	BenchMethods []benchMethod
}

func newRunner(conf *RunConfig) *runner {
	return &runner{conf: conf}
}

func (r *runner) debugf(format string, args ...interface{}) {
	if r.conf.DebugPrint != nil {
		r.conf.DebugPrint(fmt.Sprintf(format, args...))
	}
}

func (r *runner) Run() error {
	defer func() {
		if r.buildDir == "" || r.conf.NoCleanup {
			return
		}
		if err := os.RemoveAll(r.buildDir); err != nil {
			log.Printf("remove temp build dir: %v", err)
		}
	}()

	steps := []struct {
		name string
		fn   func() error
	}{
		{"find bench files", r.stepFindBenchFiles},
		{"prepare temp build dir", r.stepPrepareTempBuildDir},
		{"parse bench files", r.stepParseBenchFiles},
		{"filter only parsed files", r.stepFilterOnlyParsedFiles},
		{"sort bench files", r.stepSortBenchFiles},
		{"generate bench main", r.stepGenerateBenchMain},
		{"run bench", r.stepRunBench},
	}

	for _, step := range steps {
		if err := step.fn(); err != nil {
			return fmt.Errorf("%s: %w", step.name, err)
		}
	}

	return nil
}

func (r *runner) stepFindBenchFiles() error {
	var testDir string
	var benchFiles []string
	if strings.HasSuffix(r.conf.BenchTarget, ".php") {
		benchFiles = []string{r.conf.BenchTarget}
	} else {
		var err error
		benchFiles, err = findBenchFiles(r.conf.BenchTarget)
		if err != nil {
			return err
		}
	}
	if !strings.HasSuffix(testDir, "/") {
		testDir += "/"
	}

	r.benchFiles = make([]*benchFile, len(benchFiles))
	for i, f := range benchFiles {
		r.benchFiles[i] = &benchFile{
			fullName:  f,
			shortName: strings.TrimPrefix(f, testDir),
		}
	}

	if r.conf.DebugPrint != nil {
		for _, f := range r.benchFiles {
			r.debugf("test file: %q", f.fullName)
		}
	}

	return nil
}

func (r *runner) stepPrepareTempBuildDir() error {
	tempDir, err := ioutil.TempDir("", "kphpbench-build")
	if err != nil {
		return err
	}
	r.buildDir = tempDir
	r.debugf("temp build dir: %q", tempDir)

	return nil
}

func (r *runner) stepParseBenchFiles() error {
	for _, f := range r.benchFiles {
		src, err := ioutil.ReadFile(f.fullName)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}
		var parserErrors []*errors.Error
		errorHandler := func(e *errors.Error) {
			parserErrors = append(parserErrors, e)
		}
		rootNode, err := parser.Parse(src, conf.Config{
			Version:          &version.Version{Major: 5, Minor: 6},
			ErrorHandlerFunc: errorHandler,
		})
		if err != nil || len(parserErrors) != 0 {
			for _, parseErr := range parserErrors {
				log.Printf("%s: parse error: %v", f.fullName, parseErr)
			}
			return err
		}
		f.info = &benchParsedInfo{}
		visitor := &astVisitor{out: f.info}
		traverser.NewTraverser(visitor).Traverse(rootNode)

		if f.info.ClassName == "" {
			return fmt.Errorf("%s: can't find a benchmark class inside a file", f.shortName)
		}
	}

	return nil
}

func (r *runner) stepFilterOnlyParsedFiles() error {
	parsedFiles := make([]*benchFile, 0, len(r.benchFiles))
	for _, f := range r.benchFiles {
		if f.info != nil {
			parsedFiles = append(parsedFiles, f)
		}
	}
	r.benchFiles = parsedFiles

	return nil
}

func (r *runner) stepSortBenchFiles() error {
	sort.Slice(r.benchFiles, func(i, j int) bool {
		return r.benchFiles[i].fullName < r.benchFiles[j].fullName
	})

	for i, f := range r.benchFiles {
		f.id = i
	}

	return nil
}

func (r *runner) stepGenerateBenchMain() error {
	r.composerMode = fileutil.FileExists(filepath.Join(r.conf.ProjectRoot, "composer.json"))

	for _, f := range r.benchFiles {
		var generated bytes.Buffer
		templateData := map[string]interface{}{
			"BenchFilename":  f.fullName,
			"BenchClassName": f.info.ClassName,
			"BenchMethods":   f.info.BenchMethods,
			"Unroll":         make([]struct{}, 20),
		}
		if r.composerMode {
			templateData["Bootstrap"] = filepath.Join(r.conf.ProjectRoot, "vendor", "autoload.php")
		}
		if err := benchMainTemplate.Execute(&generated, templateData); err != nil {
			return fmt.Errorf("%s: %w", f.fullName, err)
		}
		f.generatedMain = generated.Bytes()
	}

	return nil
}

var benchMainTemplate = template.Must(template.New("bench_main").Parse(`<?php

require_once '{{.BenchFilename}}';

{{if .Bootstrap}}
require_once '{{.Bootstrap}}';
{{end}}

function __bench_main() {
  $bench = new {{.BenchClassName}}();

  {{range $bench := .BenchMethods}}

  fprintf(STDERR, "{{$bench.Key}}\t");
  $run1_start = hrtime(true);
  $bench->{{$bench.Name}}();
  $run1_end = hrtime(true);
  $op_time_approx = $run1_end - $run1_start;
  $num_tries = (int)(1000000000 / $op_time_approx);
  if ($num_tries < 40) {
    $num_tries = 40;
  }
  $time_total = 0;
  for ($i = 0; $i < $num_tries; $i += {{len $.Unroll}}) {
    $start = hrtime(true);
    {{ range $.Unroll}}
    $bench->{{$bench.Name}}();
    {{- end}}
    $elapsed = hrtime(true) - $start;
    if ($elapsed > 0) {
      $time_total += $elapsed;
    }
  }
  $avg_time = (int)($time_total / $num_tries);
  fprintf(STDERR, "$num_tries\t$avg_time.0 ns/op\n");

  {{- end}}
}

__bench_main();
`))

func (r *runner) runPhpBench() error {
	for _, f := range r.benchFiles {
		mainFilename := filepath.Join(r.buildDir, "main.php")
		if err := fileutil.WriteFile(mainFilename, f.generatedMain); err != nil {
			return err
		}

		for i := 0; i < r.conf.Count; i++ {
			args := []string{
				"-f", mainFilename,
			}
			runCommand := exec.Command(r.conf.PhpCommand, args...)
			runCommand.Dir = r.buildDir
			var runStdout bytes.Buffer
			runCommand.Stderr = r.conf.Output
			runCommand.Stdout = &runStdout
			if err := runCommand.Run(); err != nil {
				log.Printf("%s: run error: %v", f.fullName, err)
				continue
			}
		}
	}

	return nil
}

func (r *runner) stepRunBench() error {
	if r.conf.PhpCommand != "" {
		return r.runPhpBench()
	}

	for _, f := range r.benchFiles {
		mainFilename := filepath.Join(r.buildDir, "main.php")
		if err := fileutil.WriteFile(mainFilename, f.generatedMain); err != nil {
			return err
		}

		// 1. Build.
		args := []string{
			"--mode", "cli",
			"--destination-directory", r.buildDir,
		}
		if r.composerMode {
			args = append(args, "--composer-root", r.conf.ProjectRoot)
		}
		args = append(args, mainFilename)
		buildCommand := exec.Command(r.conf.KphpCommand, args...)
		buildCommand.Dir = r.buildDir
		out, err := buildCommand.CombinedOutput()
		if err != nil {
			log.Printf("%s: build error: %v: %s", f.fullName, err, out)
			continue
		}

		// 2. Run.
		for i := 0; i < r.conf.Count; i++ {
			executableName := filepath.Join(r.buildDir, "cli")
			runCommand := exec.Command(executableName)
			runCommand.Dir = r.buildDir
			var runStdout bytes.Buffer
			runCommand.Stderr = r.conf.Output
			runCommand.Stdout = &runStdout
			if err := runCommand.Run(); err != nil {
				log.Printf("%s: run error: %v", f.fullName, err)
				continue
			}
		}
	}

	return nil
}
