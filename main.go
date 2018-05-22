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
	extract := flag.Bool("extract", false, "extract intel fd modules")

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
		writeRegionToFile("_flashregion_0_flashdescriptor.bin", readRegion(f, pfd.REGION.FLASH))
		writeRegionToFile("_flashregion_1_bios.bin",readRegion(f,pfd.REGION.BIOS))
		writeRegionToFile("_flashregion_2_intel_me.bin",readRegion(f,pfd.REGION.ME))
		writeRegionToFile("_flashregion_3_gbe.bin",readRegion(f,pfd.REGION.ETHERNET))
		writeRegionToFile("_flashregion_4_platform_data.bin",readRegion(f,pfd.REGION.PLATFORM))
		writeRegionToFile("_flashregion_5_reserved.bin",readRegion(f,pfd.REGION.EXPANSION))
		writeRegionToFile("_flashregion_6_reserved.bin",readRegion(f,pfd.REGION.RESERVED1))
		writeRegionToFile("_flashregion_7_reserved.bin",readRegion(f,pfd.REGION.RESERVED2))
		writeRegionToFile("_flashregion_8_ec.bin",readRegion(f,pfd.REGION.EC))
	}

	//enc.Encode(fd)
	//enc.Encode(pfd)
}
func getRegionLimits(region RegionSectionEntry) (int64, int64, bool){
	start, _ := strconv.ParseInt(region.START, 0, 64)
	end, _ := strconv.ParseInt(region.END, 0, 64)
	error, _ := strconv.ParseInt("0x00FFF000", 0, 64)

	iserror := start >= error || start >= end

	return start, end, iserror
}

func printLayout(region RegionSectionEntry, name string) string{

	start, end, error := getRegionLimits(region)


	if(!error){
		return fmt.Sprintf("%08x:%08x %s\n", start, end, name)
	}
	return ""
}

func readRegion(file *os.File, region RegionSectionEntry, ) []byte {

	start, end, error := getRegionLimits(region)

	if(error){
		return nil
	}

	bytes := make([]byte, end-start+1)

	file.Seek(start, 0)
	_, err := file.Read(bytes)
	if err != nil {
		panic(err)
	}

	return bytes
}

func writeRegionToFile(filename string, data []byte){
	if(data != nil){
		ioutil.WriteFile(filename, data, 0644)
	}
}