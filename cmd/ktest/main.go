package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/cespare/subcmd"
	"github.com/quasilyte/ktest/internal/kenv"
	"github.com/quasilyte/ktest/internal/phpunit"
)

func main() {
	log.SetFlags(0)

	cmds := []subcmd.Command{
		{
			Name:        "phpunit",
			Description: "",
			Do:          phpunitMain,
		},

		{
			Name:        "env",
			Description: "print ktest-related env variables information",
			Do:          envMain,
		},
	}

	subcmd.Run(cmds)
}

func envMain(args []string) {
	kphpVars := []string{
		"KPHP_ROOT",
		"KPHP_TESTS_POLYFILLS_REPO",
	}

	for _, name := range kphpVars {
		v := os.Getenv(name)
		fmt.Printf("%s=%q\n", name, v)
	}
}

func phpunitMain(args []string) {
	if err := cmdPhpunit(args); err != nil {
		log.Fatalf("ktest phpunit: error: %v", err)
	}
}

func cmdPhpunit(args []string) error {
	conf := &phpunit.RunConfig{}

	workdir, err := os.Getwd()
	if err != nil {
		return err
	}

	fs := flag.NewFlagSet("ktest phpunit", flag.ExitOnError)
	debug := fs.Bool("debug", false,
		`print debug info`)
	fs.BoolVar(&conf.NoCleanup, "no-cleanup", false,
		`whether to keep temp build directory`)
	fs.StringVar(&conf.ProjectRoot, "project-root", workdir,
		`project root directory`)
	fs.StringVar(&conf.KphpCommand, "kphp-binary", "",
		`kphp binary path; if empty, $KPHP_ROOT/objs/kphp2cpp is used`)
	fs.Parse(args)

	if len(fs.Args()) == 0 {
		// TODO: print command help here?
		log.Printf("Expected at least 1 positional argument, the test target")
		return nil
	}

	testTarget, err := filepath.Abs(fs.Args()[0])
	if err != nil {
		return fmt.Errorf("resolve test target path: %v", err)
	}

	conf.ProjectRoot, err = filepath.Abs(conf.ProjectRoot)
	if err != nil {
		return fmt.Errorf("resolve project root path: %v", err)
	}
	if !strings.HasSuffix(conf.ProjectRoot, "/") {
		conf.ProjectRoot += "/"
	}

	conf.TestTarget = testTarget
	conf.TestArgv = fs.Args()[1:]
	conf.Output = os.Stdout

	if *debug {
		conf.DebugPrint = func(msg string) {
			log.Print(msg)
		}
	}

	if conf.KphpCommand == "" {
		kphpEnv := kenv.NewInfo()
		if err := kphpEnv.FindRoot(); err != nil {
			return err
		}
		conf.KphpCommand = kphpEnv.KphpBinary()
	}

	result, err := phpunit.Run(conf)
	if err != nil {
		return err
	}

	formatConfig := &phpunit.FormatConfig{
		PrintTime: true,
	}
	phpunit.FormatResult(os.Stdout, formatConfig, result)

	return nil
}
