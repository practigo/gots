package gots

import "fmt"

func ShowPackets(filename string, n int) error {
	r, err := NewReader(filename)
	if err != nil {
		return err
	}

	pat := PAT(make(map[uint16]uint16))
	pmt := PMTs(make(map[uint16]*PMT))
	pidCounts := make(map[uint16]int)

	for i := 0; i < n; i++ {
		p, err := r.Next()
		if err != nil {
			panic(err)
		}

		pd := p.ParseAll()
		pid := pd.H.PID

		switch {
		case pid == PIDPAT:
			psi := ParsePSI(pd.Payload, pd.H.PUSI)
			pat.Update(psi)
		case IsPMT(pat, pid):
			psi := ParsePSI(pd.Payload, pd.H.PUSI)
			pmt.Update(psi, pid)
		default:
			if c, ok := pidCounts[pid]; ok {
				pidCounts[pid] = c + 1
			} else {
				pidCounts[pid] = 1
			}
		}
	}

	fmt.Println("== PAT:")
	for k, v := range pat {
		fmt.Printf("program #%d: %d\n", k, v)
	}
	fmt.Println("\n== PMTs:")
	for k, v := range pmt {
		fmt.Printf("program #%d: PCR = %d, Descriptor = [%d] \n", k, v.PCR, v.PInfoLen)
		for _, s := range v.Streams {
			fmt.Printf("--> PID %d: type = %d, descriptor = [%d]\n", s.PID, s.Type, s.ESInfoLen)
		}
	}
	fmt.Println("\n== PID counts:")
	for k, v := range pidCounts {
		fmt.Printf("total %d packsts of pid %d\n", v, k)
	}

	return nil
}
