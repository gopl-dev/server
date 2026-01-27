package icon

// Icons provided by Lucide (https://lucide.dev)

import "strings"

// classAttr joins default classes with user-provided classes.
func classAttr(classOpt ...string) string {
	cls := []string{"inline-block", "align-middle"}
	if len(classOpt) == 0 {
		classOpt = []string{"w-5", "h-5"}
	}

	cls = append(cls, classOpt...)
	return strings.Join(cls, " ")
}
