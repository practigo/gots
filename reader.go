package gots

import (
	"bufio"
	"encoding/binary"
	"io"
	"os"
)

const (
	// MPEG TS constants
	PacketSize = 188
	SyncByte   = 0x47 // Bit pattern of 0x47 (ASCII char 'G')
)

const (
	// packet header constants
	NotScrambled       = "NotScrambled"
	ScrambledReserved  = "ScrambledReserved"
	EvenScrambled      = "EvenScrambled"
	OddScrambled       = "OddScrambled"
	AdaptationReserved = "AdaptationReserved"
	PayloadOnly        = "PayloadOnly"
	AdaptationOnly     = "AdaptationOnly"
	AdaptationPayload  = "AdaptationPayload"
)

func convertScrambled(i uint32) string {
	switch i {
	case 0:
		return NotScrambled
	case 0x40:
		return ScrambledReserved
	case 0x80:
		return EvenScrambled
	case 0xc0:
		return OddScrambled
	default:
		return ""
	}
}

func convertAdaption(i uint32) string {
	switch i {
	case 0:
		return AdaptationReserved
	case 0x10:
		return PayloadOnly
	case 0x20:
		return AdaptationOnly
	case 0x30:
		return AdaptationPayload
	default:
		return ""
	}
}

// IsSynced tells if a byte is a sync byte.
func IsSynced(b byte) bool {
	return b == SyncByte
}

// Header is the parsed packet header.
type Header struct {
	TEI                    bool
	PUSI                   bool
	Priority               bool
	PID                    uint32 // 13 bits, use uint32 here for simplicity
	TSC                    string // 2 bits
	AdaptationFieldControl string // 2 bits
	ContinuityCounter      uint32 // 4 bits
}

// Packet is the basic unit of data in a transport stream.
type Packet [PacketSize]byte

// Synced checks if the first byte of a Packet is a sync byte.
func (p *Packet) Synced() bool {
	return IsSynced(p[0])
}

// ParseHeader parses the first 4 bytes of a Packet into a Header.
func (p *Packet) ParseHeader() *Header {
	// 4-byte header
	h := binary.BigEndian.Uint32(p[:4])
	return &Header{
		TEI:                    h&0x800000 != 0,
		PUSI:                   h&0x400000 != 0,
		PID:                    (h & 0x1fff00) >> 8,
		TSC:                    convertScrambled(h & 0xc0),
		AdaptationFieldControl: convertAdaption(h & 0x30),
		ContinuityCounter:      h & 0xf,
	}
}

// Reader reads the packets from a ts file.
type Reader interface {
	Next() (*Packet, error)
}

type reader struct {
	br *bufio.Reader
}

func (r *reader) Next() (*Packet, error) {
	for {
		b, err := r.br.ReadByte()
		if err != nil {
			return nil, err
		}
		if !IsSynced(b) {
			continue
		}
		// back one byte
		if err = r.br.UnreadByte(); err != nil {
			return nil, err
		}
		var p Packet
		if _, err = io.ReadFull(r.br, p[:]); err != nil {
			return nil, err
		}
		return &p, nil
	}
}

// NewReader opens the file with filename and returns a Reader
// for reading packets from the file.
func NewReader(filename string) (Reader, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return &reader{bufio.NewReader(f)}, nil
}
