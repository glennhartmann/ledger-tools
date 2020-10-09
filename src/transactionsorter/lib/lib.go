package lib

import (
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
	"time"
)

func SortFile(path string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("ioutil.ReadFile: %v", err)
	}

	lines := strings.Split(string(b), "\n")
	if len(lines) == 0 {
		return nil
	}

	sortedLines, err := sortLines(lines)
	if err != nil {
		return fmt.Errorf("sortLines(): %v", err)
	}

	if err := ioutil.WriteFile(path, []byte(strings.Join(sortedLines, "\n")), 0644); err != nil {
		return fmt.Errorf("ioutil.WriteFile(): %v", err)
	}

	return nil
}

func sortLines(lines []string) ([]string, error) {
	hunks := make(sortableHunks, 0, 50 /* arbitrary */)

	curHunk, linesUsed := startFirstHunk(lines[0])
	line := ""
	i := linesUsed
	expectDate := true
	for {
		if i >= len(lines) {
			break
		}
		line = lines[i]

		if expectDate {
			var err error
			curHunk, i, hunks, expectDate, err = expectDateFlow(line, curHunk, i, hunks)
			if err != nil {
				return nil, fmt.Errorf("expectDateFlow(): %v", err)
			}
			continue
		}

		curHunk.lines = append(curHunk.lines, line)
		i++

		if isWhitespace(line) {
			expectDate = true
		}
	}
	hunks = append(hunks, curHunk)

	sort.Stable(hunks)
	return flattenHunks(hunks, len(lines)), nil
}

func expectDateFlow(line string, curHunk hunk, i int, hunks []hunk) (hunk, int, []hunk, bool, error) {
	expectDate := true

	if isWhitespaceOrCommentOrIgnorable(line) {
		curHunk.lines = append(curHunk.lines, line)
		i++
	} else {
		date, err := getDate(line)
		if err != nil {
			return hunk{}, 0, nil, false, fmt.Errorf("line %d: getDate(): %v", i, err)
		}

		hunks = append(hunks, curHunk)
		curHunk = hunk{date, make([]string, 0, 5 /* arbitrary */)}
		expectDate = false
	}

	return curHunk, i, hunks, expectDate, nil
}

func startFirstHunk(line string) (curHunk hunk, linesUsed int) {
	if isWhitespaceOrCommentOrIgnorable(line) {
		curHunk.date = time.Unix(0, 0)
		curHunk.lines = append(curHunk.lines, line)
		return curHunk, 1
	}
	return curHunk, 0
}

func getDate(line string) (time.Time, error) {
	i := strings.IndexAny(line, " \t")
	if i < 0 {
		return time.Time{}, fmt.Errorf("no whitespace found")
	}
	ds := line[:i]
	return time.Parse("2006/01/02", ds)
}

func flattenHunks(hunks []hunk, size int) []string {
	lines := make([]string, 0, size)
	for _, hunk := range hunks {
		lines = append(lines, hunk.lines...)
	}
	return lines
}

type hunk struct {
	date  time.Time
	lines []string
}

type sortableHunks []hunk

func (sh sortableHunks) Len() int {
	return len(sh)
}

func (sh sortableHunks) Less(i, j int) bool {
	return sh[i].date.Before(sh[j].date)
}

func (sh sortableHunks) Swap(i, j int) {
	t := sh[i]
	sh[i] = sh[j]
	sh[j] = t
}

func isWhitespace(s string) bool {
	return strings.TrimSpace(s) == ""
}

func isWhitespaceOrCommentOrIgnorable(s string) bool {
	s2 := strings.TrimSpace(s)
	return s2 == "" || strings.HasPrefix(s2, ";") ||
		strings.HasPrefix(s, "apply") || strings.HasPrefix(s, "end apply")
}
