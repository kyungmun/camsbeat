package cmd

import (
	"github.com/kyungmun/camsbeat/beater"

	cmd "github.com/elastic/beats/libbeat/cmd"
)

// Name of this beat
var Name = "camsbeat"

// RootCmd to handle beats cli
var RootCmd = cmd.GenRootCmd(Name, "1.0.0", beater.New)
