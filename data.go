package gots

import (
	"encoding/binary"
)

const (
	// pre-defined
	PIDPAT       = uint16(0)
	DummyTableID = 0xff
)

// SyntaxSection is the table syntax section.
/*
Sizes in bytes

	2: IDExt
	1: bits etc.
	1: SectionNum
	1: LastSecNum
	N: table data
	4: CRC32
*/
type SyntaxSection struct {
	IDExt      uint16 // table ID extension
	SectionNum uint8  // section number
	LastSecNum uint8  // last section number
	DataLen    uint16 // N, computed
	Data       []byte
	CRC32      uint32
}

// TableHeader is the data of a PSI.
/*
Sizes in bytes

	1: ID
	2: bits + section length
	N: data
*/
type TableHeader struct {
	ID         byte
	SectionLen uint16 // N
	Section    SyntaxSection
}

func (t *TableHeader) Size() int {
	return int(t.SectionLen) + 3
}

func parsePSIHeader(payload []byte) (*TableHeader, bool) {
	l := len(payload)
	if l < 3 || payload[0] == DummyTableID {
		return nil, false
	}

	sectionLen := uint16(payload[1]&3)<<8 + uint16(payload[2])
	if sectionLen > 1021 || l < 3+int(sectionLen) {
		return nil, false
	}

	section := payload[3 : 3+sectionLen]
	return &TableHeader{
		ID:         payload[0],
		SectionLen: sectionLen,
		Section: SyntaxSection{
			IDExt:      binary.BigEndian.Uint16(section[:2]),
			SectionNum: uint8(section[3]),
			LastSecNum: uint8(section[4]),
			DataLen:    sectionLen - 4 - 5,
			Data:       section[5 : sectionLen-4], // 2+1+1+1
			CRC32:      binary.LittleEndian.Uint32(section[sectionLen-4:]),
		},
	}, true
}

// PSI is the program-specific information.
type PSI struct {
	Headers  []*TableHeader
	Residual []byte
}

// ParsePSI parses the PSI packet.
func ParsePSI(payload []byte, pusi bool) *PSI {
	// handle Payload Unit Start Indicator (PUSI)
	if pusi {
		offset := int(payload[0])
		payload = payload[1+offset:]
	}

	headers := make([]*TableHeader, 0)
	// table header repeated until end of TS packet payload
	for {
		t, ok := parsePSIHeader(payload)
		if ok {
			payload = payload[t.Size():]
			headers = append(headers, t)
		} else {
			break
		}
	}

	return &PSI{
		Headers:  headers,
		Residual: payload,
	}
}

// PAT essentially gives you a map of what PIDs are part of what programs.
type PAT map[uint16]uint16

// Update updates the PAT with the PSI packet.
func (p PAT) Update(psi *PSI) {
	for _, h := range psi.Headers {
		data := h.Section.Data
		l := len(data)
		for i := 0; i <= l-4; i += 4 {
			pNum := binary.BigEndian.Uint16(data[i : i+2])         // 16 bit program num
			pid := binary.BigEndian.Uint16(data[i+2:i+4]) & 0x1fff // 13 bit PID
			p[pNum] = pid
		}
	}
}

// IsPMT checks if a pid is a PMT packet.
func IsPMT(pat PAT, pid uint16) bool {
	for _, v := range pat {
		if v == pid {
			return true
		}
	}
	return false
}

// StreamInfo ...
type StreamInfo struct {
	Type        uint8
	PID         uint16
	ESInfoLen   uint16
	Descriptors []byte
}

// PMT is the Program Mapping Table.
type PMT struct {
	PCR         uint16
	PInfoLen    uint16
	Streams     []StreamInfo
	Descriptors []byte
}

// ParsePMT parses the PSI packet into a PMT, assuming there is only one PMT.
func ParsePMT(psi *PSI) PMT {
	if len(psi.Headers) > 0 {
		h := psi.Headers[0]
		data := h.Section.Data
		t := PMT{
			PCR:      binary.BigEndian.Uint16(data[:2]) & 0x1fff, // 13 bit PID
			PInfoLen: uint16(data[2]&3)<<8 + uint16(data[3]),     // 10 bit length
			Streams:  make([]StreamInfo, 0),
		}
		offset := 4 + t.PInfoLen
		if t.PInfoLen > 0 {
			t.Descriptors = data[4:offset]
		}
		for {
			s := StreamInfo{
				Type:      uint8(data[offset]),
				PID:       binary.BigEndian.Uint16(data[offset+1:offset+3]) & 0x1fff, // 13 bit PID
				ESInfoLen: uint16(data[offset+3]&3)<<8 + uint16(data[offset+4]),      // 10 bit length
			}
			offset += 5
			if s.ESInfoLen > 0 {
				s.Descriptors = data[offset : offset+s.ESInfoLen]
				offset += s.ESInfoLen
			}
			t.Streams = append(t.Streams, s)
			if offset >= h.Section.DataLen {
				break
			}
		}
		return t
	}
	return PMT{}
}

// PMTs is a map of all the PMT.
// For each program, there is one PMT.
type PMTs map[uint16]*PMT

func (p PMTs) Update(psi *PSI, pid uint16) {
	pmt := ParsePMT(psi)
	p[pid] = &pmt
}
