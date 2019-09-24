package main

import (
	"../.."
	"flag"
)

func main(){
	targetFile := flag.String("file", "TestFile5.plist", "Target plist file to decode.")

	flag.Parse()

	iChatTool.ReadPList(*targetFile)
}