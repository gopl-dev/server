// Package icon provides reusable SVG icon components for the frontend.
package icon

// Icons provided by Lucide (https://lucide.dev)

import "strings"

// classAttr joins default classes with user-provided classes.
func classAttr(classOpt ...string) string {
	if len(classOpt) == 0 {
		classOpt = []string{"w-5", "h-5"}
	}

	cls := make([]string, 2, 2+len(classOpt)) //nolint:mnd
	cls[0] = "inline-block"
	cls[1] = "align-middle"

	cls = append(cls, classOpt...)
	return strings.Join(cls, " ")
}
