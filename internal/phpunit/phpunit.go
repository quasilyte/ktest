package phpunit

import (
	"io"
	"time"
)

// kphp test tests/
// 1. find all test scripts that match the argument
// 2. create a test main file
// 3. compile test main file
// 4. run test main file
// 5. print test main script results

type RunConfig struct {
	ProjectRoot string
	TestTarget  string
	TestArgv    []string

	KphpCommand string

	Output     io.Writer
	DebugPrint func(string)

	NoCleanup bool
}

type RunResult struct {
	Tests      int
	Assertions int
	Failures   []TestFailure
	Time       time.Duration
}

type TestFailure struct {
	Name     string
	Reason   string
	Message  string
	Location string
}

func Run(conf *RunConfig) (*RunResult, error) {
	startTime := time.Now()
	r := newRunner(conf)
	result, err := r.Run()
	if err != nil {
		return nil, err
	}
	result.Time = time.Since(startTime)
	return result, nil
}

type FormatConfig struct {
	PrintTime bool
}

func FormatResult(w io.Writer, conf *FormatConfig, result *RunResult) {
	formatResult(w, conf, result)
}
