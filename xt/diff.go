//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-16

package xt

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/xanygo/anygo/cli/xcolor"
	"github.com/xanygo/anygo/internal/zreflect"
)

func doDiff[T any](actual T, expected T) Labeleds {
	fn := xcolor.SetColorEnabled(true)
	defer fn()

	strExpected := prettyGoValue(expected)
	strActual := prettyGoValue(actual)

	var items Labeleds
	items = append(items, Labeled{Label: "Actual", Content: strActual})
	items = append(items, Labeled{Label: "Expected", Content: strExpected})

	// 0x 开头的是指针的地址，这类信息不可直接观察到
	if strExpected != strActual && len(strActual) < 60 && !strings.Contains(strActual, "(0x") {
		return items
	}

	strExpectedDump := sprintLineNo(zreflect.DumpString(expected))
	strActualDump := sprintLineNo(zreflect.DumpString(actual))

	diffValue := Labeled{Label: "Diff"}
	diffIndex2 := getDiffIndex(strExpectedDump, strActualDump)
	{
		bf1 := &bytes.Buffer{}
		bf1.WriteString(xcolor.GreenString(strExpectedDump[:diffIndex2]))

		s1, s2, f1 := strings.Cut(strExpectedDump[diffIndex2:], "\n")
		if f1 {
			bf1.WriteString(xcolor.RedString(s1))
			bf1.WriteString("\t←────🟢 diff here\n")
			bf1.WriteString(xcolor.RedString(cutDiffAfter(s2)))
		} else {
			bf1.WriteString(xcolor.RedString(strExpectedDump[diffIndex2:]))
		}
		c1 := Labeled{
			Label:   "Expected",
			Content: bf1.String(),
		}
		diffValue.Children = append(diffValue.Children, c1)
	}

	{
		bf2 := &bytes.Buffer{}
		bf2.WriteString(xcolor.GreenString(strActualDump[:diffIndex2]))

		s3, s4, f2 := strings.Cut(strActualDump[diffIndex2:], "\n")
		if f2 {
			bf2.WriteString(xcolor.RedString(s3))
			bf2.WriteString("\t←────🔴 diff here\n")
			bf2.WriteString(xcolor.RedString(cutDiffAfter(s4)))
		} else {
			bf2.WriteString(xcolor.RedString(strActualDump[diffIndex2:]))
		}
		c2 := Labeled{
			Label:   "Actual",
			Content: bf2.String(),
		}
		diffValue.Children = append(diffValue.Children, c2)
	}
	items = append(items, diffValue)
	return items
}

func prettyGoValue(v any) string {
	if v == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%#v", v)
}

func getDiffIndex(str1, str2 string) int {
	ml := min(len(str1), len(str2))
	var index int
	for i := range ml {
		if str1[i] != str2[i] {
			index = i
			break
		}
	}
	return index
}

func sprintLineNo(str string) string {
	lines := strings.Split(str, "\n")
	result := make([]string, len(lines))
	format := "%0" + strconv.Itoa(len(strconv.Itoa(len(lines)))) + "d"
	for i, line := range lines {
		result[i] = fmt.Sprintf(format+" %s", i+1, line)
	}
	return strings.Join(result, "\n")
}

func cutDiffAfter(s string) string {
	index := strings.Index(s, "\n")
	if index == -1 {
		return s
	}

	for range 30 {
		next := strings.Index(s[index+1:], "\n")
		if next == -1 {
			return s
		}
		index += next + 1
	}

	return s[:index]
}
