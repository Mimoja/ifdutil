package main

type FlashDescriptor struct {
	HeaderOffset uint32
	HEADER       FlashDescriptorHeader
	OEM          [0x40]uint8
	FLASHCONTROL []MEFlashControl
	REGION       RegionSection
	COMPONENT    ComponentSection
	PCHSTRAP     [18]string
	MASTER       MasterSection
	STRAP        [8]string
}

type FlashDescriptorHeader struct {
	FLVALSIG string
	FLMAP0   FlashDescriptorHeaderFLMAP0
	FLMAP1   FlashDescriptorHeaderFLMAP1
	FLMAP2   FlashDescriptorHeaderFLMAP2
	RESERVED [(0xefc - 0x20) / 4]uint32
	FLUMAP1  FlashDescriptorHeaderFLUMAP1
}

type FlashDescriptorHeaderFLMAP0 struct {
	RESERVED0 uint32
	NR        uint32
	FRBA      string
	RESERVED2 uint32
	RESERVED3 uint32
	RESERVED4 uint32
	RESERVED5 uint32
	NC        uint32
	FCBA      string
}

type FlashDescriptorHeaderFLMAP1 struct {
	ISL       string
	FPSBA     string
	RESERVED0 uint32
	NM        uint32
	FMBA      string
}

type FlashDescriptorHeaderFLMAP2 struct {
	RIL     string
	ICCRIBA string
	PSL     string
	FMSBA   string
}
type FlashDescriptorHeaderFLUMAP1 struct {
	RESERVED0 uint32
	VTL       uint32
	VTBA      string
}
type MEFlashControl struct {
	COMPONENT MEFlashControlComponent
	CONTROL   MEFlashControlControl
}

type MEFlashControlComponent struct {
	DeviceID0 string
	DeviceID1 string
	VendorID  string
}

type MEFlashControlControl struct {
	LowerEraseOpcode              string
	LowerWriteEnableOnWriteStatus string
	LowerWriteStatusRequired      bool
	LowerWriteGranularity         uint32
	LowerBlockAndSectorEraseSize  string
	UpperEraseOpcode              string
	UpperWriteEnableOnWriteStatus string
	UpperWriteStatusRequired      bool
	UpperWriteGranularity         uint32
	UpperBlockAndSectorEraseSize  string
}

type MasterSection struct {
	BIOS     MasterSectionEntry
	ME       MasterSectionEntry
	ETHERNET MasterSectionEntry
	RESERVED MasterSectionEntry
	EC       MasterSectionEntry
}

type MasterSectionEntry struct {
	FlashDescriptorReadAccess     bool
	FlashDescriptorWriteAccess    bool
	HostCPUBIOSRegionReadAccess   bool
	HostCPUBIOSRegionWriteAccess  bool
	IntelMERegionReadAccess       bool
	IntelMERegionWriteAccess      bool
	GbERegionReadAccess           bool
	GbERegionWriteAccess          bool
	PlatformDataRegionReadAccess  bool
	PlatformDataRegionWriteAccess bool
	ECRegionReadAccess            bool
	ECRegionWriteAccess           bool

	RequesterID string
}

type RegionSection struct {
	FLASH     RegionSectionEntry
	BIOS      RegionSectionEntry
	ME        RegionSectionEntry
	ETHERNET  RegionSectionEntry
	PLATFORM  RegionSectionEntry
	EXPANSION RegionSectionEntry
	RESERVED2 RegionSectionEntry
	RESERVED3 RegionSectionEntry
	EC        RegionSectionEntry
	//RESERVED3 RegionSectionEntry 9 or 10?
}

type RegionSectionEntry struct {
	START string
	END   string
}

type ComponentSection struct {
	FLCOMP ComponentSectionFLCOMP
	FLILL  ComponentSectionFLILL
	FLPB   ComponentSectionFLPB
}

type ComponentSectionFLCOMP struct {
	DualOutputFastReadSupport  bool
	ReadIDStatusClockFrequency uint32
	WriteEraseClockFrequency   uint32
	FastReadClockFrequency     uint32
	FastReadSupport            bool
	ReadClockFrequency         uint32
	Component1Density          uint32
	Component2Density          uint32
}

type ComponentSectionFLILL struct {
	InvalidInstruction0 string
	InvalidInstruction1 string
	InvalidInstruction2 string
	InvalidInstruction3 string
}

type ComponentSectionFLPB struct {
	FlashPartitionBoundaryAddress string
}

func getRegionByNumber(pfd FlashDescriptor, index int) (RegionSectionEntry, string, string, string) {
	switch index {
	case 0:
		return pfd.REGION.FLASH, "fd", "flashdescriptor", "Flash Descriptor"
	case 1:
		return pfd.REGION.BIOS, "bios", "bios", "Bios"
	case 2:
		return pfd.REGION.ME, "me", "intel_me", "Intel ME"
	case 3:
		return pfd.REGION.ETHERNET, "gbe", "gbe", "Ethernet"
	case 4:
		return pfd.REGION.PLATFORM, "pd", "platform_data", "Platform Data"
	case 5:
		return pfd.REGION.EXPANSION, "res1", "reserved1", "Expansion / Reserved1"
	case 6:
		return pfd.REGION.RESERVED2, "res2", "reserved2", "Reserved 2"
	case 7:
		return pfd.REGION.RESERVED3, "res3", "reserved3", "Reserved 3"
	case 8:
		return pfd.REGION.EC, "ec", "ec", "Embedded Controler"
	}
	panic("Unknown region index")
	return pfd.REGION.FLASH, "", "", ""
}

func getMasterSectionByNumber(pfd FlashDescriptor, index int) (MasterSectionEntry, string) {
	switch index {
	case 0:
		return pfd.MASTER.BIOS, "Host CPU/BIOS"
	case 1:
		return pfd.MASTER.ME, "Intel ME"
	case 2:
		return pfd.MASTER.ETHERNET, "Ethernet"
	case 3:
		return pfd.MASTER.RESERVED, "RESERVED"
	case 4:
		return pfd.MASTER.EC, "EC"

	}
	panic("Unknown region index")
	return pfd.MASTER.RESERVED, ""
}
