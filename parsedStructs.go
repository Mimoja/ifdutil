package main

type FlashDescriptor struct {
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
	RESERVED [(0xefc - 0x20)/4]uint32
	FLUMAP1  FlashDescriptorHeaderFLUMAP1
}

type FlashDescriptorHeaderFLMAP0 struct {
	RESERVED0 uint32
	RESERVED1 uint32
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
	RIL       string
	ICCRIBA   string
	PSL       string
	FMSBA     string
}
type FlashDescriptorHeaderFLUMAP1 struct {
	VTBA string
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
	FlashDescriptorReadAccess bool
	FlashDescriptorWriteAccess bool
	HostCPUBIOSRegionReadAccess bool
	HostCPUBIOSRegionWriteAccess bool
	IntelMERegionReadAccess bool
	IntelMERegionWriteAccess bool
	GbERegionReadAccess bool
	GbERegionWriteAccess bool
	PlatformDataRegionReadAccess bool
	PlatformDataRegionWriteAccess bool
	ECRegionReadAccess bool
	ECRegionWriteAccess bool

	RequesterID string
}


type RegionSection struct {
	FLASH     RegionSectionEntry
	BIOS      RegionSectionEntry
	ME        RegionSectionEntry
	ETHERNET  RegionSectionEntry
	PLATFORM  RegionSectionEntry
	EXPANSION RegionSectionEntry
	RESERVED1 RegionSectionEntry
	RESERVED2 RegionSectionEntry
	EC        RegionSectionEntry
	//RESERVED3 RegionSectionEntry 9 or 10?
}

type RegionSectionEntry struct {
	START string
	END  string
}

type ComponentSection struct {
	FLCOMP ComponentSectionFLCOMP
	FLILL ComponentSectionFLILL
	FLPB ComponentSectionFLPB
}

type ComponentSectionFLCOMP struct {
	DualOutputFastReadSupport bool
	ReadIDStatusClockFrequency uint32
	WriteEraseClockFrequency uint32
	FastReadClockFrequency uint32
	FastReadSupport bool
	ReadClockFrequency uint32
	Component1Density uint32
	Component2Density uint32
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