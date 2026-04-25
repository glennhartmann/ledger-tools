package common

import (
	"log"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

const (
	ledgerToolsName = "ledger_tools"
)

var (
	DefaultConfigDir = ""
	DefaultDataDir   = ""

	defaultBaseDir          = filepath.Join(os.ExpandEnv("${HOME}"), ".ledger_tools")
	defaultDefaultConfigDir = filepath.Join(defaultBaseDir, "config")
	defaultDefaultDataDir   = filepath.Join(defaultBaseDir, "data")
)

func init() {
	var err error
	DefaultConfigDir, err = xdg.ConfigFile(ledgerToolsName)
	if err != nil {
		DefaultConfigDir = defaultDefaultConfigDir
		log.Printf("xdg.ConfigFile() = {%+v}; using DefaultConfigDir = %q", DefaultConfigDir)
	}

	DefaultDataDir, err = xdg.DataFile(ledgerToolsName)
	if err != nil {
		DefaultDataDir = defaultDefaultDataDir
		log.Printf("xdg.DataFile() = {%+v}; using DefaultDataDir = %q", DefaultDataDir)
	}
}
