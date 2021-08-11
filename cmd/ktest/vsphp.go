package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

func benchmarkVsPHP(args []string) error {
	fs := flag.NewFlagSet("ktest bench-vs-php", flag.ExitOnError)
	flagCount := fs.Int("count", 10, `run each benchmark n times`)
	flagPhpCommand := fs.String("php", "php", `PHP command to run the benchmarks`)
	flagKphpCommand := fs.String("kphp2cpp-binary", "", `kphp binary path; if empty, $KPHP_ROOT/objs/kphp2cpp is used`)
	fs.Parse(args)

	if len(fs.Args()) == 0 {
		// TODO: print command help here?
		log.Printf("Expected at least 1 positional argument, the benchmarking target")
		return nil
	}

	benchTarget := fs.Args()[0]

	_ = flagPhpCommand

	var createdFiles []string
	defer func() {
		for _, f := range createdFiles {
			os.Remove(f)
		}
	}()
	createTempFile := func(data []byte) (string, error) {
		f, err := ioutil.TempFile("", "ktest-bench")
		if err != nil {
			return "", err
		}
		createdFiles = append(createdFiles, f.Name())
		if _, err := f.Write(data); err != nil {
			return "", err
		}
		return f.Name(), nil
	}

	// 1. Run `ktest bench ... > kphpResultsFile`
	// 2. Run `ktest bench-php ... > phpResultsFile`
	// 3. Run `ktest benchstat phpResultsFile kphpResultsFile`

	var kphpResultsFile string
	var phpResultsFile string

	{
		args := []string{
			"bench",
			"--count", fmt.Sprint(*flagCount),
		}
		if *flagKphpCommand != "" {
			args = append(args, "--kphp2cpp-binary", *flagKphpCommand)
		}
		args = append(args, benchTarget)
		out, err := exec.Command(os.Args[0], args...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("run KPHP benchmarks: %v: %s", err, out)
		}
		filename, err := createTempFile(out)
		if err != nil {
			return err
		}
		kphpResultsFile = filename
	}

	{
		args := []string{
			"bench-php",
			"--count", fmt.Sprint(*flagCount),
		}
		if *flagPhpCommand != "" {
			args = append(args, "--php", *flagPhpCommand)
		}
		args = append(args, benchTarget)
		out, err := exec.Command(os.Args[0], args...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("run PHP benchmarks: %v: %s", err, out)
		}
		filename, err := createTempFile(out)
		if err != nil {
			return err
		}
		phpResultsFile = filename
	}

	{
		args := []string{
			"benchstat",
			phpResultsFile,
			kphpResultsFile,
		}
		out, err := exec.Command(os.Args[0], args...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("run benchstat: %v: %s", err, out)
		}
		out = bytes.Replace(out, []byte("old time"), []byte("PHP time"), 1)
		out = bytes.Replace(out, []byte("new time"), []byte("KPHP time"), 1)
		fmt.Print(string(out))
	}

	return nil
}
