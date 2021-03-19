package phpunit

import (
	"bytes"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/quasilyte/ktest/internal/fileutil"
	"github.com/quasilyte/ktest/internal/kenv"
)

func TestPhpunit(t *testing.T) {
	testFiles, err := ioutil.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}

	kphpEnv := kenv.NewInfo()
	if err := kphpEnv.FindRoot(); err != nil {
		t.Fatal(err)
	}

	absFilepath := func(t *testing.T, filename string) string {
		abs, err := filepath.Abs(filename)
		if err != nil {
			t.Fatal(err)
		}
		return abs
	}

	initComposer := func(t *testing.T, workdir string) {
		if fileutil.FileExists(filepath.Join(workdir, "composer.lock")) {
			return
		}

		composerRequireCommand := exec.Command("composer", "require", "phpunit/phpunit")
		composerRequireCommand.Dir = workdir
		t.Log(composerRequireCommand.String())
		if err := composerRequireCommand.Run(); err != nil {
			t.Fatalf("run %s: %v", composerRequireCommand, err)
		}
		composerInstallCommand := exec.Command("composer", "install")
		composerInstallCommand.Dir = workdir
		t.Log(composerInstallCommand.String())
		if err := composerInstallCommand.Run(); err != nil {
			t.Fatalf("run %s: %v", composerInstallCommand, err)
		}
	}

	runTest := func(t *testing.T, filename string) {
		testDir := filepath.Join("testdata", filename)
		goldenData, err := ioutil.ReadFile(filepath.Join(testDir, "golden.txt"))
		if err != nil {
			t.Fatalf("read golden file: %v", err)
		}

		workdir := absFilepath(t, testDir)
		initComposer(t, workdir)

		var output bytes.Buffer
		result, err := Run(&RunConfig{
			ProjectRoot: workdir,
			TestTarget:  "tests",
			KphpCommand: kphpEnv.KphpBinary(),
			Output:      &output,
			NoCleanup:   true,
		})
		if err != nil {
			t.Fatal(err)
		}
		formatConfig := &FormatConfig{
			PrintTime: false,
		}
		FormatResult(&output, formatConfig, result)
		have := strings.TrimSpace(output.String())
		want := strings.TrimSpace(string(goldenData))
		if have != want {
			t.Errorf("output mismatches:\nhave:\n%s\nwant:\n%s", have, want)
		}
	}

	for i := range testFiles {
		f := testFiles[i]
		t.Run(f.Name(), func(t *testing.T) {
			runTest(t, f.Name())
		})
	}
}
