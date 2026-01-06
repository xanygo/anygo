//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-11-16

package xt

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/xanygo/anygo/cli/xcolor"
	"github.com/xanygo/anygo/internal/zreflect"
)

func sprintfDiff[T any](expected, actual T) string {
	xcolor.SetColorable(true)
	defer xcolor.SetColorable(false)

	var sb strings.Builder
	line := xcolor.CyanString(strings.Repeat("-", 120)) + "\n"
	sb.WriteString(line)

	strExpected := prettyGoValue(expected)
	strActual := prettyGoValue(actual)
	diffIndex := getDiffIndex(strExpected, strActual)

	const format1 = " %12s : "
	sb.WriteString(xcolor.GreenString(format1, "Expected"))
	sb.WriteString(strExpected[:diffIndex])
	sb.WriteString(xcolor.GreenString(strExpected[diffIndex:]))
	sb.WriteString("\n")

	sb.WriteString(xcolor.RedString(format1, "Actual"))
	sb.WriteString(strActual[:diffIndex])
	sb.WriteString(xcolor.RedString(strActual[diffIndex:]))
	sb.WriteString("\n")
	sb.WriteString(line)
	sb.WriteString("\n")

	// 0x 开头的是指针的地址，这类信息不可直接观察到
	if strExpected != strActual && len(strActual) < 60 && !strings.Contains(strActual, "(0x") {
		return sb.String()
	}

	sb.WriteString(xcolor.CyanString("Diff:"))
	sb.WriteString("\n")
	sb.WriteString(xcolor.BgGreenString("Expected:"))
	sb.WriteString("\n")

	strExpectedDump := sprintLineNo(zreflect.DumpString(expected))
	strActualDump := sprintLineNo(zreflect.DumpString(actual))

	diffIndex2 := getDiffIndex(strExpectedDump, strActualDump)
	sb.WriteString(xcolor.GreenString(strExpectedDump[:diffIndex2]))

	s1, s2, f1 := strings.Cut(strExpectedDump[diffIndex2:], "\n")
	if f1 {
		sb.WriteString(xcolor.RedString(s1))
		sb.WriteString("\t←────🟢 diff here\n")
		sb.WriteString(xcolor.RedString(cutDiffAfter(s2)))
	} else {
		sb.WriteString(xcolor.RedString(strExpectedDump[diffIndex2:]))
	}

	sb.WriteString("\n\n")
	sb.WriteString(xcolor.BgYellowString("Actual:"))
	sb.WriteString("\n")
	sb.WriteString(xcolor.GreenString(strActualDump[:diffIndex2]))

	s3, s4, f2 := strings.Cut(strActualDump[diffIndex2:], "\n")
	if f2 {
		sb.WriteString(xcolor.RedString(s3))
		sb.WriteString("\t←────🔴 diff here\n")
		sb.WriteString(xcolor.RedString(cutDiffAfter(s4)))
	} else {
		sb.WriteString(xcolor.RedString(strActualDump[diffIndex2:]))
	}

	sb.WriteString("\n")
	return sb.String()
}

func prettyGoValue(v any) string {
	return fmt.Sprintf("%#v", v)
}

func getDiffIndex(str1, str2 string) int {
	ml := min(len(str1), len(str2))
	var index int
	for i := 0; i < ml; i++ {
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

	for i := 0; i < 30; i++ {
		next := strings.Index(s[index+1:], "\n")
		if next == -1 {
			return s
		}
		index += next + 1
	}

	return s[:index]
}
