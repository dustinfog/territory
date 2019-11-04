package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"territory/logic"
	"time"
)

func main() {
	m := logic.NewMap()

	points := [][]int32{
		//{
		//	1, 25, 25, 1,
		//},
		//{
		//	1, 40, 25, 0,
		//},
		//{
		//	1, 55, 25, 1,
		//},
		//
		//{
		//	1, 55, 40, 1,
		//},
		//
		//{
		//	1, 55, 55, 1,
		//},
		//
		//{
		//	1, 40, 55, 1,
		//},
		//
		//{
		//	1, 25, 55, 1,
		//},
		//{
		//	1, 25, 40, 1,
		//},

		{
			1, 25, 25, 1,
		},
		{
			2, 35, 26, 1,
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
		{
			1, 50, 45, 0,
		},
		{
			1, 65, 50, 0,
		},
		{
			1, 65, 65, 0,
		},
		{
			1, 53, 70, 0,
		},
		{
			1, 40, 68, 0,
		},
		{
			1, 25, 63, 0,
		},
		{
			1, 15, 73, 0,
		},
		{
			1, 14, 88, 0,
		},
		{
			1, 16, 100, 0,
		},
		{
			1, 30, 95, 0,
		},
		{
			1, 43, 90, 0,
		},
		{
			1, 55, 80, 0,
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

	newRgba := image.NewRGBA(image.Rect(0, 0, 500, 500))
	colors := make(map[int32]color.RGBA)
	colors[1] = color.RGBA{
		R: 255,
		G: 0,
		B: 0,
		A: 255,
	}
	colors[2] = color.RGBA{
		R: 0,
		G: 0,
		B: 255,
		A: 255,
	}

	logic.DrawImage(m, newRgba, colors)
	logic.DrawBoundaries(m, newRgba, colors)

	f, err := os.OpenFile("/Users/panzd/Documents/territory/test8.png", os.O_SYNC|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	err = png.Encode(f, newRgba)

	if err != nil {
		fmt.Printf("%v\n", err)
	}
}
