package main

import (
	"fmt"
	"os"
	"encoding/json"
	"flag"
	"strconv"
	"io/ioutil"
)

func main() {

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", " ")


	//legacyDump := flag.Bool("fork", false, "a bool")
	layout := flag.String("layout", "", "dump regions into a flashrom layout file")
	extract := flag.Bool("extract", false, "extract intel fd modules)

	flag.Parse()


	argsWithProg := flag.Args()
	fmt.Println("Supplied arguments:")
	fmt.Println(argsWithProg)


	if(len(argsWithProg) == 0){
		panic("Please supply a flashimage to open")
	}

	f, err := os.Open(argsWithProg[0])
	if err != nil {
		panic(err)
	}

	fd := readBinaryIFD(f, 0x10)
	pfd := parseBinary(fd)


	if(*layout != ""){
		var layoutString string
		layoutString+= printLayout(pfd.REGION.FLASH, "fd")
		layoutString+= printLayout(pfd.REGION.BIOS, "bios")
		layoutString+= printLayout(pfd.REGION.ME, "me")
		layoutString+= printLayout(pfd.REGION.ETHERNET, "gbe")
		layoutString+= printLayout(pfd.REGION.PLATFORM, "pd")
		layoutString+= printLayout(pfd.REGION.EXPANSION, "res1")
		layoutString+= printLayout(pfd.REGION.RESERVED1, "res2")
		layoutString+= printLayout(pfd.REGION.RESERVED2, "res3")
		layoutString+= printLayout(pfd.REGION.EC, "ec")

		ioutil.WriteFile(*layout, []byte(layoutString), 0644)
	}

	if(*extract){

	}

	//enc.Encode(fd)
	//enc.Encode(pfd)
}
func getRegionLimits(region RegionSectionEntry) (int64, int64, int64){
	start, _ := strconv.ParseInt(region.START, 0, 64)
	end, _ := strconv.ParseInt(region.END, 0, 64)
	error, _ := strconv.ParseInt("0x00FFF000", 0, 64)

	return (start, end, error)
}

func printLayout(region RegionSectionEntry, name string) string{

	start, end, error := getRegionLimits(region)


	if(start < end && start < error && end < error){
		return fmt.Sprintf("%08x:%08x %s\n", start, end, name)
	}
	return ""
}

func readRegion(file *os.File, region RegionSectionEntry) []byte {

	start, end, error := getRegionLimits(region)


	if(start > end || start > error || end > error){
		return make([]byte, 0)
	}

	bytes := make([]byte, end-start)

	file.Seek(start, 0)
	_, err := file.Read(bytes)
	if err != nil {
		panic(err)
	}

	return bytes
}
