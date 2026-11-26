package coverage

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// profileBuffer is the run of profile lines a write buffers before flushing.
const profileBuffer = 64 << 10

// DefaultCoverFile names the profile --cover writes when --coverfile names
// none.
const DefaultCoverFile = "phpscript.cov"

// The --cover modes. Line is the bare --cover mode: measure, write the
// profile, print the one-line percentage. Func and file also write the profile,
// then print a per-symbol report in the format go tool cover -func prints.
const (
	ModeLine = "line"
	ModeFunc = "func"
	ModeFile = "file"
)

// ProfileBlock is a line of a written profile: a coverage block with its
// columns resolved from the source text. The collector reports lines only; the
// columns exist so the profile is the format go test writes and go tool cover
// renders, spanning the statement text and not the indentation around it.
type ProfileBlock struct {
	File      string
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int
	NumStmt   int
	Count     int
}

// Columns resolves each block's columns against the source text of the file it
// came from. source returns the lines of a file, or nil when the host cannot
// read it, in which case the block keeps column 1 at both ends.
func Columns(blocks []Block, source func(file string) []string) []ProfileBlock {
	cache := map[string][]string{}
	lines := func(file string) []string {
		if cached, ok := cache[file]; ok {
			return cached
		}
		src := source(file)
		cache[file] = src
		return src
	}

	out := make([]ProfileBlock, 0, len(blocks))
	for _, b := range blocks {
		block := ProfileBlock{
			File:      b.File,
			StartLine: b.StartLine,
			StartCol:  1,
			EndLine:   b.EndLine,
			EndCol:    1,
			NumStmt:   b.NumStmt,
			Count:     b.Count,
		}
		if src := lines(b.File); src != nil {
			if b.StartLine >= 1 && b.StartLine <= len(src) {
				line := src[b.StartLine-1]
				block.StartCol = len(line) - len(strings.TrimLeft(line, " \t")) + 1
			}
			if b.EndLine >= 1 && b.EndLine <= len(src) {
				block.EndCol = len(src[b.EndLine-1]) + 1
			}
		}
		out = append(out, block)
	}
	SortProfile(out)
	return out
}

// SortProfile orders blocks by file and position, a profile's listing order.
func SortProfile(blocks []ProfileBlock) {
	sort.Slice(blocks, func(i, j int) bool {
		if blocks[i].File != blocks[j].File {
			return blocks[i].File < blocks[j].File
		}
		if blocks[i].StartLine != blocks[j].StartLine {
			return blocks[i].StartLine < blocks[j].StartLine
		}
		return blocks[i].EndLine < blocks[j].EndLine
	})
}

// WriteProfile writes blocks in the format go test -coverprofile writes, with
// the header naming the mode that recorded them. A block's Count is the number
// of executions under ModeCount and 0 or 1 under ModeAtomic.
func WriteProfile(w io.Writer, blocks []ProfileBlock, mode Mode) error {
	if mode != ModeCount && mode != ModeAtomic {
		mode = DefaultMode
	}
	// One buffer holds a run of lines and one allocation holds the buffer;
	// formatting each line through fmt cost three allocations a block.
	buf := make([]byte, 0, profileBuffer)
	flush := func() error {
		if len(buf) == 0 {
			return nil
		}
		_, err := w.Write(buf)
		buf = buf[:0]
		if err != nil {
			return fmt.Errorf("write coverfile: %w", err)
		}
		return nil
	}
	buf = append(buf, "mode: "...)
	buf = append(buf, mode...)
	buf = append(buf, '\n')
	for _, b := range blocks {
		count := b.Count
		if mode != ModeCount && count > 1 {
			count = 1
		}
		if len(buf) > profileBuffer-len(b.File)-64 {
			if err := flush(); err != nil {
				return err
			}
		}
		buf = append(buf, b.File...)
		buf = append(buf, ':')
		buf = strconv.AppendInt(buf, int64(b.StartLine), 10)
		buf = append(buf, '.')
		buf = strconv.AppendInt(buf, int64(b.StartCol), 10)
		buf = append(buf, ',')
		buf = strconv.AppendInt(buf, int64(b.EndLine), 10)
		buf = append(buf, '.')
		buf = strconv.AppendInt(buf, int64(b.EndCol), 10)
		buf = append(buf, ' ')
		buf = strconv.AppendInt(buf, int64(b.NumStmt), 10)
		buf = append(buf, ' ')
		buf = strconv.AppendInt(buf, int64(count), 10)
		buf = append(buf, '\n')
	}
	return flush()
}

// Percent is the statement-weighted percentage go test reports: statements in
// blocks that ran, over all registered statements.
func Percent(blocks []ProfileBlock) float64 {
	var covered, total int
	for _, b := range blocks {
		total += b.NumStmt
		if b.Count > 0 {
			covered += b.NumStmt
		}
	}
	if total == 0 {
		return 0
	}
	return float64(covered) / float64(total) * 100
}
