package lib

import (
	"testing"

	"reflect"
	"strings"
	"time"
)

func TestSortLines(t *testing.T) {
	tests := []struct {
		in      string
		wantOut string
		wantErr bool
	}{
		{
			overallTest1, overallTest1, false,
		},
		{
			overallTest2, overallTest1, false,
		},
		{
			overallTest3, overallTest3Want, false,
		},
	}
	for i, test := range tests {
		splitGot, err := sortLines(strings.Split(test.in, "\n"))
		if (err != nil) != test.wantErr {
			t.Errorf("%d: sortLines() = err(%v), want non-nil error %v", i, err, test.wantErr)
		}
		if err == nil {
			got := strings.Join(splitGot, "\n")
			if got != test.wantOut {
				t.Errorf("%d: sortLines() = %s, want %s", i, got, test.wantOut)
			}
		}
	}
}

func TestExpectDateFlow(t *testing.T) {
	tests := []struct {
		line           string
		curHunk        hunk
		i              int
		hunks          []hunk
		wantHunk       hunk
		wantI          int
		wantHunks      []hunk
		wantExpectDate bool
		wantErr        bool
	}{
		{
			line:           "",
			curHunk:        hunk{},
			i:              5,
			hunks:          []hunk{},
			wantHunk:       hunk{lines: []string{""}},
			wantI:          6,
			wantHunks:      []hunk{},
			wantExpectDate: true,
			wantErr:        false,
		},
		{
			line:           "   ;  ",
			curHunk:        hunk{time.Unix(5, 5), []string{"f"}},
			i:              8,
			hunks:          []hunk{},
			wantHunk:       hunk{time.Unix(5, 5), []string{"f", "   ;  "}},
			wantI:          9,
			wantHunks:      []hunk{},
			wantExpectDate: true,
			wantErr:        false,
		},
		{
			line:           "bad",
			curHunk:        hunk{time.Unix(5, 5), []string{"f"}},
			i:              8,
			hunks:          []hunk{},
			wantHunk:       hunk{},
			wantI:          0,
			wantHunks:      nil,
			wantExpectDate: false,
			wantErr:        true,
		},
		{
			line:           "2020/02/07 ",
			curHunk:        hunk{time.Unix(5, 5), []string{"f"}},
			i:              1,
			hunks:          []hunk{hunk{}},
			wantHunk:       hunk{time.Date(2020, time.February, 7, 0, 0, 0, 0, time.FixedZone("UTC", 0)), []string{}},
			wantI:          1,
			wantHunks:      []hunk{hunk{}, hunk{time.Unix(5, 5), []string{"f"}}},
			wantExpectDate: false,
			wantErr:        false,
		},
	}
	for j, test := range tests {
		hunk, i, hunks, expectDate, err := expectDateFlow(test.line, test.curHunk, test.i, test.hunks)
		if !hunk.date.Equal(test.wantHunk.date) || !reflect.DeepEqual(hunk.lines, test.wantHunk.lines) ||
			i != test.wantI || !reflect.DeepEqual(hunks, test.wantHunks) ||
			expectDate != test.wantExpectDate || (err != nil) != test.wantErr {
			t.Errorf("%d: expectDateFlow() = %v, %d, %v, %v, %v", j, hunk, i, hunks, expectDate, err)
		}
	}
}

func TestStartFirstHunk(t *testing.T) {
	tests := []struct {
		line  string
		wantH hunk
		wantI int
	}{
		{
			"", hunk{time.Unix(0, 0), []string{""}}, 1,
		},
		{
			"          ", hunk{time.Unix(0, 0), []string{"          "}}, 1,
		},
		{
			"    ;      ", hunk{time.Unix(0, 0), []string{"    ;      "}}, 1,
		},
		{
			"f", hunk{}, 0,
		},
	}
	for _, test := range tests {
		h, i := startFirstHunk(test.line)
		if !reflect.DeepEqual(h, test.wantH) {
			t.Errorf("startFirstHunk(%s) = h(%v), want h(%v)", test.line, h, test.wantH)
		}
		if i != test.wantI {
			t.Errorf("startFirstHunk(%s) = i(%d), want i(%d)", test.line, i, test.wantI)
		}
	}
}

func TestGetDate(t *testing.T) {
	tests := []struct {
		line    string
		want    time.Time
		wantErr bool
	}{
		{
			"", time.Time{}, true,
		},
		{
			" ", time.Time{}, true,
		},
		{
			"f", time.Time{}, true,
		},
		{
			"f ", time.Time{}, true,
		},
		{
			"2020/02/07", time.Time{}, true,
		},
		{
			"2020/02/07 ", time.Date(2020, time.February, 7, 0, 0, 0, 0, time.FixedZone("UTC", 0)), false,
		},
	}
	for _, test := range tests {
		d, err := getDate(test.line)
		if (err != nil) != test.wantErr {
			t.Errorf("getDate(%s) = err(%v), want non-nil error: %v", test.line, err, test.wantErr)
		}
		if err == nil {
			if !d.Equal(test.want) {
				t.Errorf("getDate(%s) = %v, want %v", test.line, d, test.want)
			}
		}
	}
}

func TestFlattenHunks(t *testing.T) {
	h := []hunk{
		hunk{time.Time{}, []string{"a", "b", "c"}},
		hunk{time.Time{}, []string{"d", "e"}},
		hunk{time.Time{}, []string{"f"}},
	}
	fh := flattenHunks(h, 6)
	if !reflect.DeepEqual(fh, []string{"a", "b", "c", "d", "e", "f"}) {
		t.Errorf("flattenHunks() failed, and I don't want to be helpful")
	}
}

func TestIsWhitespace(t *testing.T) {
	tests := []struct {
		s    string
		want bool
	}{
		{
			"", true,
		},
		{
			"    	  ", true,
		},
		{
			"    	  \n", true,
		},
		{
			"   ; 	  \n", false,
		},
		{
			";", false,
		},
		{
			"; sdkfj asldkf jsldkfj", false,
		},
		{
			" sdkfj asldkf jsldkfj", false,
		},
		{
			"apply tag asdfsdf", false,
		},
		{
			"end apply tag asdfsdf", false,
		},
	}

	for _, test := range tests {
		got := isWhitespace(test.s)
		if got != test.want {
			t.Errorf("isWhitespace(%s) = %v, want %v", test.s, got, test.want)
		}
	}
}

func TestIsWhitespaceOrCommentOrIgnorable(t *testing.T) {
	tests := []struct {
		s    string
		want bool
	}{
		{
			"", true,
		},
		{
			"    	  ", true,
		},
		{
			"    	  \n", true,
		},
		{
			"   ; 	  \n", true,
		},
		{
			";", true,
		},
		{
			"; sdkfj asldkf jsldkfj", true,
		},
		{
			" sdkfj asldkf jsldkfj", false,
		},
		{
			"apply tag asdfsdf", true,
		},
		{
			"end apply tag asdfsdf", true,
		},
	}

	for _, test := range tests {
		got := isWhitespaceOrCommentOrIgnorable(test.s)
		if got != test.want {
			t.Errorf("isWhitespaceOrCommentOrIgnorable(%s) = %v, want %v", test.s, got, test.want)
		}
	}
}

const (
	overallTest1 = `
; vim:filetype=ledger
; in this file, $ is used for CAD

2020/01/01 landlord
    Rent                           $2356
    Assets:Reimbursements:landlord = $2
    Liabilities:landlord           = $-2362.25

2020/01/01 what
    ; nothing
    whatever expense account    $232.64
    credit card                 = $-6.93
    ; :some:tags:

2020/01/16 mountain
    ; lift tickets
    Snowboarding          $51.98
    credit card
    ; comments are fun

2020/01/17 incomememememe
    ; income
    work Bonus         $-00
    work RRSP Match    $-1000
    ; deductions
    Tax                     $3352.3
    EI                      $2
    CPP                     $1
    ; assets
    bank                    $10000000 = $5
    ; comments:: $-236
    ; andtags:: $-247
`

	overallTest2 = `
; vim:filetype=ledger
; in this file, $ is used for CAD

2020/01/01 landlord
    Rent                           $2356
    Assets:Reimbursements:landlord = $2
    Liabilities:landlord           = $-2362.25

2020/01/17 incomememememe
    ; income
    work Bonus         $-00
    work RRSP Match    $-1000
    ; deductions
    Tax                     $3352.3
    EI                      $2
    CPP                     $1
    ; assets
    bank                    $10000000 = $5
    ; comments:: $-236
    ; andtags:: $-247

2020/01/01 what
    ; nothing
    whatever expense account    $232.64
    credit card                 = $-6.93
    ; :some:tags:

2020/01/16 mountain
    ; lift tickets
    Snowboarding          $51.98
    credit card
    ; comments are fun
`

	overallTest3 = `


2000/05/07 a
    ; cool
    stuff
    ;yeaaaaaahhh
    ;
    ;


;
1995/03/02 b
    stuff



2015/12/12 c
asdf
;





1992/02/02 d
    yo


`

	overallTest3Want = `


1992/02/02 d
    yo



1995/03/02 b
    stuff



2000/05/07 a
    ; cool
    stuff
    ;yeaaaaaahhh
    ;
    ;


;
2015/12/12 c
asdf
;




`
)
