package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"territory/logic"
	"time"
)

func main() {
	m := logic.NewMap()

	points := [][]int32{
		{
			1, 25, 25, 1,
		},
		{
			1, 35, 26, 1,
		},
		{
			1, 29, 35, 0,
		},
		{
			2, 43, 27, 1,
		},
		{
			1, 35, 45, 0,
		},
	}

	flags := make([]*logic.Flag, 0, len(points))

	for i, point := range points {
		x, y, allianceId, fortress := point[1], point[2], point[0], point[3] == 1
		f, err := m.AddFlag(x, y, allianceId, fortress, time.Now().Add(time.Second*time.Duration(i)))
		if err != nil {
			fmt.Printf("%d,%d,%v\n", x, y, err)
		} else {
			fmt.Printf("=========== \n")
			flags = append(flags, f)
		}
	}

	//m.RemoveFlag(flags[1])
	//m.RemoveFlag(flags[2])

	newRgba := image.NewRGBA(image.Rect(0, 0, 250, 250))
	//DrawImage(m, newRgba)
	logic.DrawBoundaries(m, newRgba, 1)
	logic.DrawBoundaries(m, newRgba, 2)

	f, err := os.OpenFile("/Users/panzd/test_.png", os.O_SYNC|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	err = png.Encode(f, newRgba)

	if err != nil {
		fmt.Printf("%v\n", err)
	}
}
