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

	str1 := sprintLineNo(zreflect.DumpString(expected))
	str2 := sprintLineNo(zreflect.DumpString(actual))

	ml := min(len(str1), len(str2))
	var diffIndex int
	for i := 0; i < ml; i++ {
		if str1[i] != str2[i] {
			diffIndex = i
			break
		}
	}
	var sb strings.Builder
	line := xcolor.CyanString(strings.Repeat("-", 120)) + "\n"
	sb.WriteString(line)
	const format1 = " %12s : "
	sb.WriteString(xcolor.GreenString(format1, "Expected"))
	a := fmt.Sprintf("%#v", actual)
	sb.WriteString(a)
	sb.WriteString("\n")
	sb.WriteString(xcolor.RedString(format1, "Actual"))
	b := fmt.Sprintf("%#v", expected)
	sb.WriteString(b)
	sb.WriteString("\n")
	sb.WriteString(line)
	sb.WriteString("\n")

	sb.WriteString(xcolor.CyanString("Diff:"))
	sb.WriteString("\n")
	sb.WriteString(xcolor.BgGreenString("Expected:"))
	sb.WriteString("\n")
	sb.WriteString(xcolor.GreenString(str1[:diffIndex]))

	s1, s2, f1 := strings.Cut(str1[diffIndex:], "\n")
	if f1 {
		sb.WriteString(xcolor.RedString(s1))
		sb.WriteString("\t←────🟢 diff here")
		sb.WriteString("\n")
		sb.WriteString(xcolor.RedString(s2))
	} else {
		sb.WriteString(xcolor.RedString(str1[diffIndex:]))
	}

	sb.WriteString("\n\n")
	sb.WriteString(xcolor.BgYellowString("Actual:"))
	sb.WriteString("\n")
	sb.WriteString(xcolor.GreenString(str2[:diffIndex]))

	s3, s4, f2 := strings.Cut(str2[diffIndex:], "\n")
	if f2 {
		sb.WriteString(xcolor.RedString(s3))
		sb.WriteString("\t←────🔴 diff here")
		sb.WriteString("\n")
		sb.WriteString(xcolor.RedString(s4))
	} else {
		sb.WriteString(xcolor.RedString(str2[diffIndex:]))
	}

	sb.WriteString("\n")
	return sb.String()
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
