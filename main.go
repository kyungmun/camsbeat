package main

import (
	"os"
	"github.com/kyungmun/camsbeat/cmd"
)

func main() {

	//logWriter.Write([]byte("<128>lsbeat udp message send test"))
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
	/*
	err := beat.Run("camsbeat", "", beater.New)
	if err != nil {
		os.Exit(1)
	}
	*/
}
