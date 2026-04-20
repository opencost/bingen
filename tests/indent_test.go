package test

import (
	"fmt"
	"testing"

	"github.com/kubecost/bingen/pkg/generator"
)

func TestBasicIndent(t *testing.T) {
	i := generator.NewIndent(4)
	i.OutN(2)
	i.In()
	fmt.Printf("%sTest\n", i)
}
