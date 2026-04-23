package main

import (
	"fmt"
	"charm.land/lipgloss/v2"
	"image/color"
)

func main() {
	c := lipgloss.Color("#ff0000")
	fmt.Printf("%T\n", c)
	var _ color.Color = c
}
