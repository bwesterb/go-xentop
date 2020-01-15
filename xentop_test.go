package xentop

import (
	"testing"
)

func TestIssue3(t *testing.T) {
	header := []string{
		"NAME", "STATE", "CPU(sec)", "CPU(%)", "MEM(k)", "MEM(%)", "MAXMEM(k)",
		"MAXMEM(%)", "VCPUS", "NETS", "NETTX(k)", "NETRX(k)", "VBDS", "VBD_OO",
		"VBD_RD", "VBD_WR", "VBD_RSECT", "VBD_WSECT", "SSI",
	}
	fields, err := parseLine(
		"                              Windows7 (Dev) --b---      92470    0.0    8388600    1.6    8389632       1.6     4    0        0        0    1        0    93779   452880    4923585   12252390    0",
		header,
	)
	if err != nil {
		t.Fatalf("parseLine: %v", err)
	}
	if fields["NAME"] != "Windows7 (Dev)" {
		t.Fatal()
	}
}
