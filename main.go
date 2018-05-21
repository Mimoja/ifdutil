package main

import (
	"fmt"
	"os"
	"encoding/json"
	"flag"
)

func main() {

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", " ")


	argsWithProg := os.Args[1:]
	fmt.Println("Supplied arguments:")
	fmt.Println(argsWithProg)

	//legacyDump := flag.Bool("fork", false, "a bool")
	//layout := flag.Bool("fork", false, "a bool")

	flag.Parse()



	if(len(argsWithProg) == 0){
		panic("Please supply a flashimage to open")
	}

	f, err := os.Open(argsWithProg[0])
	if err != nil {
		panic(err)
	}

	fd := readBinaryIFD(f, 0x10)
	pfd := parseBinary(fd)


	//enc.Encode(fd)
	enc.Encode(pfd)
}

