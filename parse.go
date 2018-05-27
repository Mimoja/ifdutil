package main

import (
	"fmt"
	"strconv"
)

func parseBinary(descriptor BinaryFlashDescriptor) FlashDescriptor {
	var fd FlashDescriptor

	fd.HeaderOffset = descriptor.HeaderOffset
	fd.HEADER = FlashDescriptorHeader{
		FLVALSIG: toHexString(descriptor.Header.Flvalsig, 8),
		FLMAP0: FlashDescriptorHeaderFLMAP0{
			RESERVED0: getBits(descriptor.Header.Flmap0, 27, 31),
			NR:        getBits(descriptor.Header.Flmap0, 24, 26),
			FRBA:      toHexString(getBits(descriptor.Header.Flmap0, 16, 23)<<4, 0),
			RESERVED2: getBits(descriptor.Header.Flmap0, 13, 15),
			RESERVED3: getBits(descriptor.Header.Flmap0, 12, 12),
			RESERVED4: getBits(descriptor.Header.Flmap0, 11, 11),
			RESERVED5: getBits(descriptor.Header.Flmap0, 10, 10),
			NC:        getBits(descriptor.Header.Flmap0, 8, 9) + 1,
			FCBA:      toHexString(getBits(descriptor.Header.Flmap0, 0, 7)<<4, 0),
		},
		FLMAP1: FlashDescriptorHeaderFLMAP1{
			ISL:       toHexString(getBits(descriptor.Header.Flmap1, 24, 31), 8),
			FPSBA:     toHexString(getBits(descriptor.Header.Flmap1, 16, 23)<<4, 2),
			RESERVED0: getBits(descriptor.Header.Flmap1, 11, 15),
			NM:        getBits(descriptor.Header.Flmap1, 8, 10),
			FMBA:      toHexString(getBits(descriptor.Header.Flmap1, 0, 7)<<4, 0),
		},
		FLMAP2: FlashDescriptorHeaderFLMAP2{
			RIL:     toHexString(getBits(descriptor.Header.Flmap2, 24, 31), 8),
			ICCRIBA: toHexString(getBits(descriptor.Header.Flmap2, 16, 23), 4),
			PSL:     toHexString(getBits(descriptor.Header.Flmap2, 8, 15), 4),
			FMSBA:   toHexString(getBits(descriptor.Header.Flmap2, 0, 7)<<4, 0),
		},
		RESERVED: descriptor.Header.Reserved,
		FLUMAP1: FlashDescriptorHeaderFLUMAP1{
			RESERVED0: getBits(descriptor.Header.Flumap1, 16, 31),
			VTL:       getBits(descriptor.Header.Flumap1, 8, 15),
			VTBA:      toHexString(getBits(descriptor.Header.Flumap1, 0, 7)<<4, 6),
		},
	}

	fd.OEM = descriptor.OEM

	// TODO depend on IFD version
	// is it 9 or 10? or 5?
	maxRegions := 9

	/**
	if (ifd_version >= IFD_VERSION_2)
		base_mask = 0x7fff;
	else
		base_mask = 0xfff;
	*/
	base_mask := uint32(0xfff)
	limit_mask := uint32(base_mask << 16)

	for i := 0; i < maxRegions; i++ {
		var regionData = descriptor.FR.Flreg[i]

		rs := RegionSectionEntry{
			START: toHexString((regionData&base_mask)<<12, 8),
			END:   toHexString(((regionData&limit_mask)>>4)|0xfff, 8),
		}

		switch i {
		case 0:
			fd.REGION.FLASH = rs
			break
		case 1:
			fd.REGION.BIOS = rs
			break

		case 2:
			fd.REGION.ME = rs
			break

		case 3:
			fd.REGION.ETHERNET = rs
			break

		case 4:
			fd.REGION.PLATFORM = rs
			break

		case 5:
			fd.REGION.EXPANSION = rs
			break

		case 6:
			fd.REGION.RESERVED2 = rs
			break

		case 7:
			fd.REGION.RESERVED3 = rs
			break

		case 8:
			fd.REGION.EC = rs
			break

			//case 9:
			//	fd.REGION.RESERVED3 = rs
			//	break

		}
	}

	cs := ComponentSection{
		FLCOMP: ComponentSectionFLCOMP{
			DualOutputFastReadSupport:  isBitSet(descriptor.FC.Flcomp, 30),
			ReadIDStatusClockFrequency: getSPIFrequency(getBits(descriptor.FC.Flcomp, 27, 29)),
			WriteEraseClockFrequency:   getSPIFrequency(getBits(descriptor.FC.Flcomp, 24, 26)),
			FastReadClockFrequency:     getSPIFrequency(getBits(descriptor.FC.Flcomp, 21, 23)),
			FastReadSupport:            isBitSet(descriptor.FC.Flcomp, 20),
			ReadClockFrequency:         getSPIFrequency(getBits(descriptor.FC.Flcomp, 17, 19)),
			Component1Density:          getDensity(getBits(descriptor.FC.Flcomp, 0, 2)),
			Component2Density:          getDensity(getBits(descriptor.FC.Flcomp, 3, 5)),
		},
		//TODO deside based on IFD version
		FLILL: ComponentSectionFLILL{
			InvalidInstruction0: toHexString(getBits(descriptor.FC.Flill, 0, 7), 2),
			InvalidInstruction1: toHexString(getBits(descriptor.FC.Flill, 8, 15), 2),
			InvalidInstruction2: toHexString(getBits(descriptor.FC.Flill, 16, 23), 2),
			InvalidInstruction3: toHexString(getBits(descriptor.FC.Flill, 24, 31), 2),
		},
		FLPB: ComponentSectionFLPB{
			FlashPartitionBoundaryAddress: toHexString(getBits(descriptor.FC.Flpb, 0, 15)<<12, 8),
		},
	}
	fd.COMPONENT = cs

	for i := 0; i < len(descriptor.FPS.Pchstrp); i++ {
		fd.PCHSTRAP[i] = toHexString(descriptor.FPS.Pchstrp[i], 8)
	}

	fd.MASTER = MasterSection{
		BIOS:     parseFLMSTR(descriptor.FM.Flmstr1),
		ME:       parseFLMSTR(descriptor.FM.Flmstr2),
		ETHERNET: parseFLMSTR(descriptor.FM.Flmstr3),
		RESERVED: parseFLMSTR(descriptor.FM.Flmstr4),
		EC:       parseFLMSTR(descriptor.FM.Flmstr5),
	}

	for index, element := range descriptor.FMS.Data {
		fd.STRAP[index] = toHexString(element, 8)
	}

	for i := 0; i < len(descriptor.VSCC); i++ {
		Vscc := descriptor.VSCC[i]
		var mfc MEFlashControl

		mfc.COMPONENT.DeviceID0 = toHexString(getBits(Vscc.Jid, 8, 15), 2)
		mfc.COMPONENT.DeviceID1 = toHexString(getBits(Vscc.Jid, 16, 23), 2)
		mfc.COMPONENT.VendorID = toHexString(getBits(Vscc.Jid, 0, 7), 2)

		mfc.CONTROL.LowerEraseOpcode = toHexString(getBits(Vscc.Vscc, 24, 31), 2)
		if Vscc.Vscc&(1<<20) != 0 {
			mfc.CONTROL.LowerWriteEnableOnWriteStatus = "0x06"
		} else {
			mfc.CONTROL.LowerWriteEnableOnWriteStatus = "0x50"
		}

		if Vscc.Vscc&(1<<19) != 0 {
			mfc.CONTROL.LowerWriteStatusRequired = true
		} else {
			mfc.CONTROL.LowerWriteStatusRequired = false
		}

		if Vscc.Vscc&(1<<18) != 0 {
			mfc.CONTROL.LowerWriteGranularity = 64
		} else {
			mfc.CONTROL.LowerWriteGranularity = 1
		}
		switch (Vscc.Vscc >> 16) & 3 {
		case 0:
			mfc.CONTROL.LowerBlockAndSectorEraseSize = "0x00FF"
			break
		case 1:
			mfc.CONTROL.LowerBlockAndSectorEraseSize = "0x1000"
			break
		case 2:
			mfc.CONTROL.LowerBlockAndSectorEraseSize = "0x2000"
			break
		case 3:
			mfc.CONTROL.LowerBlockAndSectorEraseSize = "0x8000"
			break
		}

		mfc.CONTROL.UpperEraseOpcode = toHexString(getBits(Vscc.Vscc, 8, 15), 2)
		if Vscc.Vscc&(1<<4) != 0 {
			mfc.CONTROL.UpperWriteEnableOnWriteStatus = "0x06"
		} else {
			mfc.CONTROL.UpperWriteEnableOnWriteStatus = "0x50"
		}

		if Vscc.Vscc&(1<<3) != 0 {
			mfc.CONTROL.UpperWriteStatusRequired = true
		} else {
			mfc.CONTROL.UpperWriteStatusRequired = false
		}

		if Vscc.Vscc&(1<<2) != 0 {
			mfc.CONTROL.UpperWriteGranularity = 64
		} else {
			mfc.CONTROL.UpperWriteGranularity = 1
		}
		switch (Vscc.Vscc) & 3 {
		case 0:
			mfc.CONTROL.UpperBlockAndSectorEraseSize = "0x00FF"
			break
		case 1:
			mfc.CONTROL.UpperBlockAndSectorEraseSize = "0x1000"
			break
		case 2:
			mfc.CONTROL.UpperBlockAndSectorEraseSize = "0x2000"
			break
		case 3:
			mfc.CONTROL.UpperBlockAndSectorEraseSize = "0x8000"
			break
		}

		fd.FLASHCONTROL = append(fd.FLASHCONTROL, mfc)
	}

	return fd
}

func getSPIFrequency(freq uint32) uint32 {

	SPI_FREQUENCY_20MHZ := 0
	SPI_FREQUENCY_33MHZ := 1
	SPI_FREQUENCY_48MHZ := 2
	SPI_FREQUENCY_50MHZ_30MHZ := 4
	SPI_FREQUENCY_17MHZ := 6

	switch int(freq) {
	case SPI_FREQUENCY_20MHZ:
		return 20
		break
	case SPI_FREQUENCY_33MHZ:
		return 33
		break
	case SPI_FREQUENCY_48MHZ:
		return 48
		break
	case SPI_FREQUENCY_50MHZ_30MHZ:
		return 50
		/*
			//TODO fix IFD version check
			switch (ifd_version) {
				case IFD_VERSION_1:
				return 50
				case IFD_VERSION_2:
				return 30
			}
		*/
		break
	case SPI_FREQUENCY_17MHZ:
		return 17

	}
	return 0
}

func getDensity(density uint32) uint32 {
	COMPONENT_DENSITY_512KB := 0
	COMPONENT_DENSITY_1MB := 1
	COMPONENT_DENSITY_2MB := 2
	COMPONENT_DENSITY_4MB := 3
	COMPONENT_DENSITY_8MB := 4
	COMPONENT_DENSITY_16MB := 5
	COMPONENT_DENSITY_32MB := 6
	COMPONENT_DENSITY_64MB := 7
	COMPONENT_DENSITY_UNUSED := 0xf

	switch int(density) {
	case COMPONENT_DENSITY_512KB:
		return 1 << 19

	case COMPONENT_DENSITY_1MB:
		return 1 << 20

	case COMPONENT_DENSITY_2MB:
		return 1 << 21

	case COMPONENT_DENSITY_4MB:
		return 1 << 22

	case COMPONENT_DENSITY_8MB:
		return 1 << 23

	case COMPONENT_DENSITY_16MB:
		return 1 << 24

	case COMPONENT_DENSITY_32MB:
		return 1 << 25

	case COMPONENT_DENSITY_64MB:
		return 1 << 26

	case COMPONENT_DENSITY_UNUSED:
		return 0

	}
	return 0xFFFFFFFF
}

func parseFLMSTR(flmstr uint32) MasterSectionEntry {
	FLMSTR_WR_SHIFT_V1 := 24
	FLMSTR_RD_SHIFT_V1 := 16

	//FLMSTR_WR_SHIFT_V2 := 20
	//FLMSTR_RD_SHIFT_V2 := 8

	// TODO IFD Version
	/*
		if (ifd_version >= IFD_VERSION_2) {
			wr_shift = FLMSTR_WR_SHIFT_V2;
			rd_shift = FLMSTR_RD_SHIFT_V2;
		} else {
			wr_shift := FLMSTR_WR_SHIFT_V1;
			rd_shift := FLMSTR_RD_SHIFT_V1;
		}
	*/

	wr_shift := uint32(FLMSTR_WR_SHIFT_V1)
	rd_shift := uint32(FLMSTR_RD_SHIFT_V1)

	entry := MasterSectionEntry{
		FlashDescriptorReadAccess:     isBitSet(flmstr, rd_shift+0),
		FlashDescriptorWriteAccess:    isBitSet(flmstr, wr_shift+0),
		HostCPUBIOSRegionReadAccess:   isBitSet(flmstr, rd_shift+1),
		HostCPUBIOSRegionWriteAccess:  isBitSet(flmstr, wr_shift+1),
		IntelMERegionReadAccess:       isBitSet(flmstr, rd_shift+2),
		IntelMERegionWriteAccess:      isBitSet(flmstr, wr_shift+2),
		GbERegionReadAccess:           isBitSet(flmstr, rd_shift+3),
		GbERegionWriteAccess:          isBitSet(flmstr, wr_shift+3),
		PlatformDataRegionReadAccess:  isBitSet(flmstr, rd_shift+4),
		PlatformDataRegionWriteAccess: isBitSet(flmstr, wr_shift+4),
		ECRegionReadAccess:            isBitSet(flmstr, rd_shift+8),
		ECRegionWriteAccess:           isBitSet(flmstr, wr_shift+8),
		RequesterID:                   toHexString(getBits(flmstr, 0, 15), 8),
	}
	return entry
}

func isBitSet(val uint32, bit uint32) bool {
	return (val & (1 << bit)) != 0
}

func toHexString(val uint32, zeroFilll uint32) string {
	formatString := fmt.Sprintf("0x%%0%dX", zeroFilll)
	return fmt.Sprintf(formatString, val)
}

func getBits(val uint32, start uint8, end uint8) uint32 {
	var mask uint32

	for i := 0; i <= int(end-start); i++ {
		mask <<= 1
		mask |= 1
	}

	return (val >> start) & mask
}
func parseKomplex(descriptor FlashDescriptor) BinaryFlashDescriptor {
	var fd BinaryFlashDescriptor
	fd.HeaderOffset = descriptor.HeaderOffset

	flm0 := descriptor.HEADER.FLMAP0
	flm1 := descriptor.HEADER.FLMAP1
	flm2 := descriptor.HEADER.FLMAP2
	flum1 := descriptor.HEADER.FLUMAP1

	fd.Header = BinaryFlashDescriptorHeader{
		Flvalsig: fromString(descriptor.HEADER.FLVALSIG),
		Flmap0: flm0.RESERVED0<<27 | flm0.NR<<24 | fromString(flm0.FRBA)>>4<<16 | flm0.RESERVED2<<13 |
			flm0.RESERVED2<<12 | flm0.RESERVED2<<11 | flm0.RESERVED2<<10 | (flm0.NC-1)<<8 |
			fromString(flm0.FCBA)>>4,
		Flmap1: fromString(flm1.ISL)<<24 | fromString(flm1.FPSBA)>>4<<16 |
			flm1.RESERVED0<<11 | flm1.NM<<8 | fromString(flm1.FMBA)>>4,
		Flmap2: fromString(flm2.RIL)<<24 | fromString(flm2.ICCRIBA)<<16 | fromString(flm2.PSL)<<8 |
			fromString(flm2.FMSBA)>>4,
		Reserved: descriptor.HEADER.RESERVED,
		Flumap1:  flum1.RESERVED0<<16 | flum1.VTL<<8 | fromString(flum1.VTBA)>>4,
	}

	fd.OEM = descriptor.OEM

	//TODO dont rely on reserved area for all the other structs
	return fd
}

func fromString(hex string) uint32 {
	val, err := strconv.ParseInt(hex, 0, 64)

	if err != nil {
		panic(err)
	}

	return uint32(val)
}
