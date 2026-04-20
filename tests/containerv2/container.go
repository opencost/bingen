package containerv2

import (
	"fmt"
	"strings"
)

type Container struct {
	Name     string
	Children []string
	oldValue float64
	Value    *float64 //@bingen:field[version=2]
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
	sb.WriteString(fmt.Sprintf("  (oldValue): %f\n", c.oldValue))
	if c.Value != nil {
		sb.WriteString(fmt.Sprintf("  Value: %f\n", *c.Value))
	} else {
		sb.WriteString("  Value: nil\n")
	}
	sb.WriteString("]\n")

	return sb.String()
}

func migrateContainer(c *Container, from uint8, to uint8) {
	fmt.Printf("Migrating Container from version %d to %d\n", from, to)
	if c.oldValue != 0.0 {
		v := c.oldValue
		c.Value = &v
	}
}
