package kenv

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/quasilyte/ktest/internal/fileutil"
)

type Info struct {
	KphpRoot string
}

func NewInfo() *Info {
	return &Info{}
}

func (info *Info) KphpBinary() string {
	return filepath.Join(info.KphpRoot, "objs", "bin", "kphp2cpp")
}

func (info *Info) FindRoot() error {
	envRoot := os.Getenv("KPHP_ROOT")
	if envRoot != "" {
		if !fileutil.FileExists(envRoot) {
			return fmt.Errorf("KPHP_ROOT points to a non-existing directory")
		}
		info.KphpRoot = envRoot
		return nil
	}

	home, err := os.UserHomeDir()
	homeKphp := filepath.Join(home, "kphp")
	if err == nil && fileutil.FileExists(homeKphp) {
		info.KphpRoot = homeKphp
		return nil
	}

	return fmt.Errorf("$KPHP_ROOT is not set")
}
