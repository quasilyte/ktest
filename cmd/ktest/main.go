package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cespare/subcmd"
)

func main() {
	log.SetFlags(0)

	cmds := []subcmd.Command{
		{
			Name:        "phpunit",
			Description: "",
			Do:          cmdPhpunit,
		},

		{
			Name:        "env",
			Description: "print ktest-related env variables information",
			Do:          cmdEnv,
		},
	}

	subcmd.Run(cmds)
}

func cmdEnv(args []string) {
	kphpVars := []string{
		"KPHP_ROOT",
		"KPHP_TESTS_POLYFILLS_REPO",
	}

	for _, name := range kphpVars {
		v := os.Getenv(name)
		fmt.Printf("%s=%q\n", name, v)
	}
}

func cmdPhpunit(args []string) {

}
