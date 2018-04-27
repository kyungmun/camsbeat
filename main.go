package main

import (
	"os"
	"github.com/kyungmun/camsbeat/cmd"

	//"net"
	//"log"
)

func main() {

	//logWriter, error := net.Dial("udp", "10.165.60.64:514")
	//if error != nil {
	//	log.Fatal("error")
	//}
	//defer logWriter.Close()
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
