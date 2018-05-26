package main

import (
	"os"
	"encoding/binary"
	"bufio"
	"fmt"
)

type BinaryFlashDescriptor struct {
	HeaderOffset uint32
	Header BinaryFlashDescriptorHeader
	OEM    [0x40]uint8
	FR     BinaryFR
	FC     BinaryFC
	FPS    BinaryFPS
	FM     BinaryFM
	FMS    BinaryFMS
	VSCC []BinaryVSCC
}


type BinaryFlashDescriptorHeader struct {
	Flvalsig uint32
	Flmap0   uint32
	Flmap1   uint32
	Flmap2   uint32
	Reserved [3804/4]uint32
	Flumap1  uint32
}

type BinaryFR struct {
	Flreg [9]uint32
}

type BinaryFC struct {
	Flcomp uint32
	Flill  uint32
	Flpb   uint32
}


type BinaryFPS struct {
	Pchstrp [18]uint32
}

type BinaryFM struct {
	Flmstr1 uint32
	Flmstr2 uint32
	Flmstr3 uint32
	Flmstr4 uint32
	Flmstr5 uint32
}

type BinaryFMS struct {
	Data [8]uint32
}
type BinaryVSCC struct {
	Jid  uint32
	Vscc uint32
}

func readBinaryIFD(f *os.File, offset int64) BinaryFlashDescriptor{

	var FlashDescriptor BinaryFlashDescriptor

	fmt.Println("Reading FD")
	f.Seek(offset, 0)

	FlashDescriptor.HeaderOffset = uint32(offset)
	FlashDescriptor.Header = readIFDHeader(f)

	frba := ((FlashDescriptor.Header.Flmap0 >> 16) & 0xFF) << 4
	var fr BinaryFR
	f.Seek(int64(frba), 0)
	binary.Read(bufio.NewReader(f), binary.LittleEndian, &fr)
	FlashDescriptor.FR = fr

	fcba:= (FlashDescriptor.Header.Flmap0 & 0xFF) << 4
	var fc BinaryFC
	f.Seek(int64(fcba), 0)
	binary.Read(bufio.NewReader(f), binary.LittleEndian, &fc)
	FlashDescriptor.FC = fc

	fpsba:= ((FlashDescriptor.Header.Flmap1 >> 16) & 0xFF) << 4
	var fps BinaryFPS
	f.Seek(int64(fpsba), 0)
	binary.Read(bufio.NewReader(f), binary.LittleEndian, &fps)
	FlashDescriptor.FPS=fps

	fmba:=  ((FlashDescriptor.Header.Flmap1) & 0xFF) << 4
	var fm BinaryFM
	f.Seek(int64(fmba), 0)
	binary.Read(bufio.NewReader(f), binary.LittleEndian, &fm)
	FlashDescriptor.FM = fm

	fmsba := ((FlashDescriptor.Header.Flmap2) & 0xFF) << 4
	var fms BinaryFMS
	f.Seek(int64(fmsba), 0)
	binary.Read(bufio.NewReader(f), binary.LittleEndian, &fms)
	FlashDescriptor.FMS = fms

	vtl :=  (FlashDescriptor.Header.Flumap1 >> 8) & 0xFF
	vtba := (FlashDescriptor.Header.Flumap1 & 0xFF) << 4

	f.Seek(int64(vtba), 0)
	reader := bufio.NewReader(f)
	for i := uint32(0); i < vtl; i++ {
		var Vscc BinaryVSCC
		binary.Read(reader, binary.LittleEndian, &Vscc)
		FlashDescriptor.VSCC = append(FlashDescriptor.VSCC, Vscc)
	}

	f.Seek(0xF00, 0)
	binary.Read(bufio.NewReader(f), binary.LittleEndian, &FlashDescriptor.OEM)

	return FlashDescriptor
}

func readIFDHeader(f *os.File) BinaryFlashDescriptorHeader {

	var fd BinaryFlashDescriptorHeader
	binary.Read(bufio.NewReader(f), binary.LittleEndian, &fd)
	return fd
}