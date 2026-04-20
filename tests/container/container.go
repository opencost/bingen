package container

import (
	"fmt"
	"strings"
)

type Container struct {
	Name     string
	Children []string
	Value    float64
}

func (c *Container) String() string {
	var sb strings.Builder
	sb.WriteString("[Container\n")
	sb.WriteString("  Name: " + c.Name + "\n")
	sb.WriteString("  Children: [")
	if len(c.Children) > 0 {
		sb.WriteRune('\n')
		for _, child := range c.Children {
			sb.WriteString("    " + child + "\n")
		}
		sb.WriteString("  ]\n")
	} else {
		sb.WriteString("]\n")
	}
	sb.WriteString(fmt.Sprintf("  Value: %f\n", c.Value))
	sb.WriteString("]\n")

	return sb.String()
}
