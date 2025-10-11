//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-05

package xmetric

import "github.com/xanygo/anygo/xerror"

func Dump(s SpanReader) map[string]any {
	if IsNoop(s) {
		return nil
	}
	children := s.Children()
	cs := make([]map[string]any, 0, len(children))
	for _, child := range children {
		cs = append(cs, Dump(child))
	}
	result := make(map[string]any, 6)
	result["Name"] = s.Name()
	result["Cost"] = s.EndTime().Sub(s.StartTime()).String()
	if attrs := s.Attributes(); len(attrs) > 0 {
		result["Attrs"] = attrs
	}
	if es := xerror.String(s.Error()); es != "" {
		result["Error"] = es
	}
	if n := s.AttemptCount(); n > 0 {
		result["Attempt"] = n
	}
	if len(cs) > 0 {
		result["Children"] = cs
	}
	return result
}
