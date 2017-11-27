// Go wrapper of the xentop utility.
package xentop

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"reflect"
	"strconv"
)

var fieldNames [][]byte

// A line in xentop
type Line struct {
	Name               string  `field:"NAME"`
	State              string  `field:"STATE"`
	CpuTime            int64   `field:"CPU(sec)"`
	CpuFraction        float32 `field:"CPU(%)"`
	Memory             int64   `field:"MEM(k)"`
	MaxMemory          int64   `field:"MAXMEM(k)"`
	MemoryFraction     float32 `field:"MEM(%)"`
	MaxMemoryFraction  float32 `field:"MAXMEM(%)"`
	VirtualCpus        int64   `field:"VCPUS"`
	NetworkInterfaces  int64   `field:"NETS"`
	NetworkTx          int64   `field:"NETTX(k)"`
	NetworkRx          int64   `field:"NETRX(k)"`
	VirtualDisks       int64   `field:"VBDS"`
	DiskBlockedIO      int64   `field:"VBD_OO"`
	DiskReadOps        int64   `field:"VBD_RD"`
	DiskWriteOps       int64   `field:"VBD_WR"`
	DiskSectorsRead    int64   `field:"VBD_RSEC"`
	DiskSectorsWritten int64   `field:"VBD_WSEC"`
	SSID               int64   `field:"SSID"`
}

// Fills a Line struct with the values from parseLine
func fillLine(data map[string][]byte) (ret Line, errs []error) {
	errs = []error{}
	pRet := &Line{}
	sv := reflect.Indirect(reflect.ValueOf(pRet))
	st := sv.Type()
	for i := 0; i < st.NumField(); i++ {
		fieldType := st.Field(i)
		fieldName, ok := fieldType.Tag.Lookup("field")
		if !ok {
			continue
		}
		val, ok := data[fieldName]
		if !ok {
			errs = append(errs, fmt.Errorf("Missing field  %s", fieldName))
			continue
		}
		delete(data, fieldName)
		field := sv.FieldByIndex(fieldType.Index)
		if string(val) == "n/a" || string(val) == "no limit" {
			continue
		}
		switch fieldType.Type.Kind() {
		case reflect.String:
			field.SetString(string(val))
		case reflect.Float32:
			pVal, err := strconv.ParseFloat(string(val), 32)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s: could not parse: %s", fieldName, err))
				continue
			}
			field.SetFloat(float64(pVal))
		case reflect.Int64:
			pVal, err := strconv.ParseInt(string(val), 10, 64)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s: could not parse: %s", fieldName, err))
				continue
			}
			field.SetInt(pVal)
		default:
			panic("Encountered unexpected fieldtype in Line struct")
		}
	}
	ret = *pRet
	return
}

// Represents a field in the headerline
type headerField struct {
	offset int
	field  string
}

// Parse a line returned by "xentop -b"
func parseLine(line []byte, header []headerField) (ret map[string][]byte) {
	ret = make(map[string][]byte)
	for i, hf := range header {
		var nextOffset int
		if i == len(header)-1 {
			nextOffset = len(line)
		} else {
			nextOffset = header[i+1].offset
		}
		ret[hf.field] = bytes.TrimSpace(line[hf.offset:nextOffset])
	}
	return
}

// Parse a headerline returned by "xentop -b"
func parseHeader(line []byte) (ret []headerField) {
	originalLength := len(line)
	ret = make([]headerField, 0)
	for {
		fieldOffset := originalLength - len(line)
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			return
		}
		ok := false
		for _, fieldName := range fieldNames {
			if !bytes.HasPrefix(line, fieldName) {
				continue
			}
			ret = append(ret, headerField{
				offset: fieldOffset,
				field:  string(fieldName),
			})
			line = line[len(fieldName):]
			ok = true
			break
		}
		if !ok {
			bits := bytes.SplitN(line, []byte(" "), 2)
			ret = append(ret, headerField{
				offset: fieldOffset,
				field:  string(bits[0]),
			})
			if len(bits) == 1 {
				return
			}
			line = bits[1]
		}
	}
}

// Runs xentop and writes lines and errors back over the provided channels.
func XenTop(lines chan<- Line, errs chan<- error) {
	cmd := exec.Command("xentop", "-b")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		errs <- fmt.Errorf("fatal: %s", err)
		return
	}
	defer stdout.Close()

	if err := cmd.Start(); err != nil {
		errs <- fmt.Errorf("fatal: %s", err)
		return
	}

	r := bufio.NewReader(stdout)

	var header []headerField

	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			errs <- fmt.Errorf("fatal: %s", err)
			return
		}

		isHeader := bytes.HasPrefix(bytes.TrimSpace(line), []byte("NAME"))

		if isHeader {
			header = parseHeader(line)
			continue
		}

		if header == nil {
			errs <- fmt.Errorf("Missing header")
			return
		}

		pLine, pErrs := fillLine(parseLine(line, header))
		for _, err := range pErrs {
			errs <- err
		}
		lines <- pLine
	}
}

func init() {
	// fill the fieldNames slice
	fieldNames = make([][]byte, 0)
	lineType := reflect.ValueOf(Line{}).Type()
	for i := 0; i < lineType.NumField(); i++ {
		fieldType := lineType.Field(i)
		if fieldName, ok := fieldType.Tag.Lookup("field"); ok {
			fieldNames = append(fieldNames, []byte(fieldName))
		}
	}
}
