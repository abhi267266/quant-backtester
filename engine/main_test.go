package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestInspectHead(t *testing.T) {
	csvContent := `Timestamp,Open,High,Low,Close,Volume
2023-01-01T15:04:05Z,100.0,105.0,99.0,104.5,1000
2023-01-02,104.5,106.0,103.0,105.0,2000`

	// 1. Mock Stdin
	r, w, _ := os.Pipe()
	w.Write([]byte(csvContent))
	w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	// 2. Mock Stdout
	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	// 3. Mock CLI Arguments
	oldArgs := os.Args
	os.Args = []string{"engine", "inspect", "head", "-n", "1"}
	defer func() { os.Args = oldArgs }()

	// 4. Run main application
	main()

	// 5. Read output
	wOut.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, rOut)
	output := buf.String()

	// 6. Assertions
	if !strings.Contains(output, "2023-01-01") {
		t.Errorf("expected output to contain Timestamp string")
	}
	if !strings.Contains(output, "104.50000000") {
		t.Errorf("expected output to correctly format Close price 104.5 as 104.50000000, got:\n%s", output)
	}
	// Ensure we only parsed the first row
	if strings.Contains(output, "106.00000000") {
		t.Errorf("head -n 1 should only return the first row. Row 2 was detected.")
	}
}
