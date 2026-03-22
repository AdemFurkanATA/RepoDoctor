package languages

import (
	"bufio"
	"bytes"
	"strings"
)

type pythonSignalCounts struct {
	imports int
	funcs   int
	classes int
}

func collectPythonSignalsBounded(content []byte) pythonSignalCounts {
	counts := pythonSignalCounts{}
	scanner := bufio.NewScanner(bytes.NewReader(content))
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 256*1024)

	inTripleSingle := false
	inTripleDouble := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if inTripleSingle {
			if strings.Contains(line, "'''") {
				inTripleSingle = false
			}
			continue
		}
		if inTripleDouble {
			if strings.Contains(line, `"""`) {
				inTripleDouble = false
			}
			continue
		}

		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "'''") {
			if !strings.Contains(line[3:], "'''") {
				inTripleSingle = true
			}
			continue
		}
		if strings.HasPrefix(line, `"""`) {
			if !strings.Contains(line[3:], `"""`) {
				inTripleDouble = true
			}
			continue
		}

		if strings.HasPrefix(line, "import ") || strings.HasPrefix(line, "from ") {
			counts.imports++
		}
		if strings.HasPrefix(line, "def ") || strings.HasPrefix(line, "async def ") {
			counts.funcs++
		}
		if strings.HasPrefix(line, "class ") {
			counts.classes++
		}
	}

	return counts
}
