package main

/**
ifdutil Copyright (C) 2018 Mimoja <git@mimoja.de>
Based on ifdtool Copyright (C) 2011 Google Inc
*/
import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

func main() {

	//TODO Deal with ICCRIBA!!

	fmt.Printf("ifdutil v%s -- Copyright (C) 2018 Mimoja <git@mimoja.de>.\n\n", "0.1.0")
	fmt.Print("This program is free software: you can redistribute it and/or modify\n" +
		"it under the terms of the GNU General Public License as published by\n" +
		"the Free Software Foundation, version 2 of the License.\n\n" +
		"This program is distributed in the hope that it will be useful,\n" +
		"but WITHOUT ANY WARRANTY; without even the implied warranty of\n" +
		"MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the\n" +
		"GNU General Public License for more details.\n\n")

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", " ")

	legacyDump := flag.Bool("dump", false, "dump in original ifdtool format (includes some new fields)")
	layout := flag.String("layout", "", "dump regions into a flashrom layout file")
	extract := flag.Bool("extract", false, "extract intel fd modules")
	write := flag.Bool("write", false, "Write ifd to ifd.bin")
	findMagic := flag.Bool("magic", false, "Search image for ifd magic")

	flag.Parse()

	argsWithProg := flag.Args()

	if len(argsWithProg) == 0 {
		panic("Please supply a flashimage or configfile to open")
	}

	f, err := os.Open(argsWithProg[0])
	if err != nil {
		panic(err)
	}

	var magicOffset int64
	magicOffset = 0x10

	//TODO error handle if not found!
	for i := 0; ; i++ {
		f.Seek(int64(i), 0)
		header := readIFDHeader(f)
		if header.Flvalsig == 0x0FF0A55A {
			magicOffset = int64(i)
			if *findMagic {
				fmt.Printf("Found IFD Magic at 0x%08X\n", i)
			}
			break;
		}
	}


	fd := readBinaryIFD(f, magicOffset)
	pfd := parseBinary(fd)

	if *write {
		filename := "ifd.bin"
		writeFDtoFile(filename, pfd)
	}

	if *layout != "" {
		var layoutString string
		for i := 0; i < 9; i++ {
			region, shortname, _, _ := getRegionByNumber(pfd, i)
			layoutString += printLayout(region, shortname)
		}
		ioutil.WriteFile(*layout, []byte(layoutString), 0644)
	}

	if *extract {
		for i := 0; i < 9; i++ {
			region, _, longname, _ := getRegionByNumber(pfd, i)
			filename := fmt.Sprintf("flashregion_%d_%s.bin", i, longname)
			writeRegionToFile(filename, readRegion(f, region))
		}
	}

	if *legacyDump {
		fmt.Printf("FLMAP0:    0x%08x\n", fd.Header.Flmap0)
		fmt.Printf("  NR:      %d\n", pfd.HEADER.FLMAP0.NR)
		fmt.Printf("  FRBA:    %s\n", pfd.HEADER.FLMAP0.FRBA)
		fmt.Printf("  NC:      %d\n", pfd.HEADER.FLMAP0.NC)
		fmt.Printf("  FCBA:    %s\n", pfd.HEADER.FLMAP0.FCBA)

		fmt.Printf("FLMAP1:    0x%08x\n", fd.Header.Flmap1)
		fmt.Printf("  ISL:     %s\n", pfd.HEADER.FLMAP1.ISL)
		fmt.Printf("  FPSBA:   %s\n", pfd.HEADER.FLMAP1.FPSBA)
		fmt.Printf("  NM:      %d\n", pfd.HEADER.FLMAP1.NM)
		fmt.Printf("  FMBA:    %s\n", pfd.HEADER.FLMAP1.FMBA)

		fmt.Printf("FLMAP2:    0x%08x\n", fd.Header.Flmap2)
		fmt.Printf("  RIL:     %s\n", pfd.HEADER.FLMAP2.RIL)
		fmt.Printf("  ICCRIBA: %s\n", pfd.HEADER.FLMAP2.ICCRIBA)
		fmt.Printf("  PSL:     %s\n", pfd.HEADER.FLMAP2.PSL)
		fmt.Printf("  FMSBA:   %s\n", pfd.HEADER.FLMAP2.FMSBA)

		fmt.Printf("FLUMAP1:   0x%08x\n", fd.Header.Flumap1)
		fmt.Printf("  Intel ME VSCC Table Length (VTL):        %d\n",
			pfd.HEADER.FLUMAP1.VTL)
		fmt.Printf("  Intel ME VSCC Table Base Address (VTBA): %s\n\n",
			pfd.HEADER.FLUMAP1.VTBA)

		// dumpt VSCC
		fmt.Printf("ME VSCC table:\n")
		for i := uint32(0); i < pfd.HEADER.FLUMAP1.VTL; i++ {
			fmt.Printf("  JID%d:  0x%08x\n", i, fd.VSCC[i].Jid)
			fmt.Printf("    SPI Componend Device ID 1:          %s\n", pfd.FLASHCONTROL[i].COMPONENT.DeviceID1)
			fmt.Printf("    SPI Componend Device ID 0:          %s\n", pfd.FLASHCONTROL[i].COMPONENT.DeviceID0)
			fmt.Printf("    SPI Componend Vendor ID:            %s\n", pfd.FLASHCONTROL[i].COMPONENT.VendorID)

			fmt.Printf("  VSCC%d: 0x%08x\n", i, fd.VSCC[i].Vscc)
			fmt.Printf("    Lower Erase Opcode:                 %s\n", pfd.FLASHCONTROL[i].CONTROL.LowerEraseOpcode)
			fmt.Printf("    Lower Write Enable on Write Status: %s\n", pfd.FLASHCONTROL[i].CONTROL.LowerWriteEnableOnWriteStatus)
			fmt.Printf("    Lower Write Status Required:        %v\n", pfd.FLASHCONTROL[i].CONTROL.LowerWriteStatusRequired)
			fmt.Printf("    Lower Write Granularity:            %d bytes\n", pfd.FLASHCONTROL[i].CONTROL.LowerWriteGranularity)
			fmt.Printf("    Lower Block / Sector Erase Size:    %s\n", pfd.FLASHCONTROL[i].CONTROL.LowerBlockAndSectorEraseSize)
			fmt.Printf("    Upper Erase Opcode:                 %s\n", pfd.FLASHCONTROL[i].CONTROL.UpperEraseOpcode)
			fmt.Printf("    Upper Write Enable on Write Status: %s\n", pfd.FLASHCONTROL[i].CONTROL.UpperWriteEnableOnWriteStatus)
			fmt.Printf("    Upper Write Status Required:        %v\n", pfd.FLASHCONTROL[i].CONTROL.UpperWriteStatusRequired)
			fmt.Printf("    Upper Write Granularity:            %d bytes\n", pfd.FLASHCONTROL[i].CONTROL.UpperWriteGranularity)
			fmt.Printf("    Upper Block / Sector Erase Size:    %s\n", pfd.FLASHCONTROL[i].CONTROL.UpperBlockAndSectorEraseSize)
		}
		fmt.Printf("\n")

		// dump oem
		fmt.Printf("OEM Section:\n")
		for i := 0; i < 4; i++ {
			fmt.Printf("%02x:", i<<4)
			for j := 0; j < 16; j++ {
				fmt.Printf(" %02x", pfd.OEM[(i<<4)+j])
			}

			fmt.Printf("\n")
		}
		fmt.Printf("\n")

		// dump FR
		var maxRegions int
		if(fd.Version == 1){
			maxRegions = 5
		}else{
			maxRegions = 9
		}

		fmt.Printf("Found Region Section\n")
		for i := 0; i < maxRegions; i++ {
			fmt.Printf("FLREG%d:    0x%08x\n", i, fd.FR.Flreg[i])
			region, _, _, description := getRegionByNumber(pfd, i)
			base, limit, unused := getRegionLimits(region)
			unusedString := ""
			if unused {
				unusedString = "(unused)"
			}
			fmt.Printf("  Flash Region %d (%s): %08x - %08x %s\n",
				i, description, base, limit, unusedString)

		}
		fmt.Printf("\n")
		fmt.Printf("Found Component Section\n")
		fmt.Printf("FLCOMP     0x%08x\n", fd.FC.Flcomp)
		fmt.Printf("  Dual Output Fast Read Support:       %v\n", pfd.COMPONENT.FLCOMP.DualOutputFastReadSupport)
		fmt.Printf("  Read ID/Read Status Clock Frequency: %d\n", pfd.COMPONENT.FLCOMP.ReadIDStatusClockFrequency)
		fmt.Printf("  Write/Erase Clock Frequency:         %d\n", pfd.COMPONENT.FLCOMP.WriteEraseClockFrequency)
		fmt.Printf("  Fast Read Clock Frequency:           %d\n", pfd.COMPONENT.FLCOMP.FastReadClockFrequency)
		fmt.Printf("  Fast Read Support:                   %v\n", pfd.COMPONENT.FLCOMP.FastReadSupport)
		fmt.Printf("  Read Clock Frequency:                %d\n", pfd.COMPONENT.FLCOMP.ReadClockFrequency)

		fmt.Printf("  Component 2 Density:                 %d\n", pfd.COMPONENT.FLCOMP.Component2Density)
		fmt.Printf("  Component 1 Density:                 %d\n", pfd.COMPONENT.FLCOMP.Component1Density)

		fmt.Printf("FLILL      0x%08x\n", fd.FC.Flill)
		fmt.Printf("  Invalid Instruction 3: %s\n", pfd.COMPONENT.FLILL.InvalidInstruction3)
		fmt.Printf("  Invalid Instruction 2: %s\n", pfd.COMPONENT.FLILL.InvalidInstruction2)
		fmt.Printf("  Invalid Instruction 1: %s\n", pfd.COMPONENT.FLILL.InvalidInstruction1)
		fmt.Printf("  Invalid Instruction 0: %s\n", pfd.COMPONENT.FLILL.InvalidInstruction0)
		fmt.Printf("FLPB       0x%08x\n", fd.FC.Flpb)
		fmt.Printf("  Flash Partition Boundary Address: %s\n\n",
			pfd.COMPONENT.FLPB.FlashPartitionBoundaryAddress)

		fmt.Printf("Found PCH Strap Section\n")
		for i, elem := range pfd.PCHSTRAP {
			fmt.Printf("PCHSTRP%02d: %s\n", i, elem)
		}

		fmt.Printf("\n")

		fmt.Printf("Found Master Section\n")
		for i := 0; i < 5; i++ {
			section, name := getMasterSectionByNumber(pfd, i)
			fmt.Printf("FLMSTR%d:   (%s)\n", i, name)
			fmt.Printf("  EC Region Write Access:            %v\n", section.ECRegionWriteAccess)
			fmt.Printf("  Platform Data Region Write Access: %v\n", section.PlatformDataRegionWriteAccess)
			fmt.Printf("  GbE Region Write Access:           %v\n", section.GbERegionWriteAccess)
			fmt.Printf("  Intel ME Region Write Access:      %v\n", section.IntelMERegionWriteAccess)
			fmt.Printf("  Host CPU/BIOS Region Write Access: %v\n", section.HostCPUBIOSRegionWriteAccess)
			fmt.Printf("  Flash Descriptor Write Access:     %v\n", section.FlashDescriptorWriteAccess)

			fmt.Printf("  EC Region Read Access:             %v\n", section.ECRegionReadAccess)
			fmt.Printf("  Platform Data Region Read Access:  %v\n", section.PlatformDataRegionReadAccess)
			fmt.Printf("  GbE Region Read Access:            %v\n", section.GbERegionReadAccess)
			fmt.Printf("  Intel ME Region Read Access:       %v\n", section.IntelMERegionReadAccess)
			fmt.Printf("  Host CPU/BIOS Region Read Access:  %v\n", section.HostCPUBIOSRegionReadAccess)
			fmt.Printf("  Flash Descriptor Read Access:      %v\n", section.FlashDescriptorReadAccess)
			fmt.Printf("\n")
		}
		fmt.Printf("\n")

		fmt.Printf("Found Processor Strap Section\n")
		for _, elem := range pfd.STRAP {
			fmt.Printf("????:      %s\n", elem)
		}

	}

	//enc.Encode(fd)
	enc.Encode(pfd)
}
func getRegionLimits(region RegionSectionEntry) (int64, int64, bool) {
	start, _ := strconv.ParseInt(region.START, 0, 64)
	end, _ := strconv.ParseInt(region.END, 0, 64)
	error, _ := strconv.ParseInt("0x00FFF000", 0, 64)

	iserror := start >= error || start >= end

	//TODO is  (start | error) always 0x00FFFFFF when unused?

	return start, end, iserror
}

func printLayout(region RegionSectionEntry, name string) string {

	start, end, error := getRegionLimits(region)

	if !error {
		return fmt.Sprintf("%08x:%08x %s\n", start, end, name)
	}
	return ""
}

func readRegion(file *os.File, region RegionSectionEntry) []byte {

	start, end, error := getRegionLimits(region)

	if error {
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

func writeRegionToFile(filename string, data []byte) {
	if data != nil {
		ioutil.WriteFile(filename, data, 0644)
	}
}

func writeFDtoFile(filaname string, pfd FlashDescriptor) {

	fd := parseKomplex(pfd)

	fmt.Println("Writing testfd.bin")

	f, err := os.Create("testfd.bin")
	defer f.Close()

	if err != nil {
		panic(err)
	}

	b := make([]byte, 1)
	b[0] = 0xff

	// Fill with 0xff
	prefix := bytes.Repeat(b, int(fromString(pfd.REGION.FLASH.END)-fromString(pfd.REGION.FLASH.START)+1))
	f.Write(prefix)

	f.Seek(int64(fd.HeaderOffset), 0)
	writeField(f, fd.Header)

	f.Seek(0xF00, 0)
	writeField(f, fd.OEM)

	frba := fromString(pfd.HEADER.FLMAP0.FRBA)
	if fieldNeedsManualWrite(fd.HeaderOffset + frba) {
		f.Seek(int64(frba), 0)
		writeField(f, fd.FR)
	} else {
		fmt.Println("\tSkiping FR, already writen")
	}

	fcba := fromString(pfd.HEADER.FLMAP0.FCBA)
	if fieldNeedsManualWrite(fd.HeaderOffset + fcba) {
		f.Seek(int64(fcba), 0)
		writeField(f, fd.FC)
	} else {
		fmt.Println("\tSkiping FC, already writen")
	}

	fpsba := fromString(pfd.HEADER.FLMAP1.FPSBA)
	if fieldNeedsManualWrite(fd.HeaderOffset + fpsba) {
		f.Seek(int64(fcba), 0)
		writeField(f, fd.FPS)
	} else {
		fmt.Println("\tSkiping FPS, already writen")
	}

	fmba := fromString(pfd.HEADER.FLMAP1.FMBA)
	if fieldNeedsManualWrite(fd.HeaderOffset + fmba) {
		f.Seek(int64(fcba), 0)
		writeField(f, fd.FM)
	} else {
		fmt.Println("\tSkiping FM, already writen")
	}

	fmsba := fromString(pfd.HEADER.FLMAP2.FMSBA)
	if fieldNeedsManualWrite(fd.HeaderOffset + fmsba) {
		f.Seek(int64(fcba), 0)
		writeField(f, fd.FMS)
	} else {
		fmt.Println("\tSkiping FMS, already writen")
	}

	vsccba := fromString(pfd.HEADER.FLUMAP1.VTBA)
	if fieldNeedsManualWrite(fd.HeaderOffset + vsccba) {
		f.Seek(int64(vsccba), 0)
		for i := uint32(0); i < pfd.HEADER.FLUMAP1.VTL; i++ {
			writeField(f, fd.VSCC[i])
		}
	} else {
		fmt.Println("\tSkiping VSCC, already writen")
	}

	fmt.Println("Done. \n")

}

func fieldNeedsManualWrite(addres uint32) bool {
	return !(addres > 32 && addres < 32+3804)
}

func writeField(file *os.File, data interface{}) {

	var bin_buf bytes.Buffer

	binary.Write(&bin_buf, binary.LittleEndian, data)

	_, err := file.Write(bin_buf.Bytes())

	if err != nil {
		panic(err)
	}

}
