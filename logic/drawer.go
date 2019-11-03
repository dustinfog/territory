package logic

import (
	"fmt"
	"image"
	"image/color"
)

func DrawImage(m *Map, image *image.RGBA) {
	for x, row := range m.tiles {
		for y, tile := range row {
			var a uint8
			if tile.IsValid() {
				a = 255
			} else {
				a = 100
			}

			r := 255 / uint8(tile.GetAllianceId())

			for i := 0; i < 4; i++ {
				for j := 0; j < 4; j++ {
					//isV, _ := tile.IsVertex()
					//if isV {
					//	r = 0
					//}

					image.SetRGBA(int(x)*4+i, int(y)*4+j, color.RGBA{
						r,
						0,
						0,
						a,
					})
				}
			}
		}
	}
}

func DrawBoundaries(m *Map, image *image.RGBA, allianceId int32) {
	fmt.Printf("drawing %d\n", allianceId)
	cl := color.RGBA{
		255 / uint8(allianceId),
		0,
		0,
		255,
	}
	bs := NewBoundarySeeker(m, allianceId)
	vertex, _ := bs.Next()

	for !bs.Finished() {
		next, _ := bs.Next()

		startX := int(vertex.X)*4 + int(vertex.Type.Pos.X)*3
		startY := int(vertex.Y)*4 + int(vertex.Type.Pos.Y)*3

		endX := int(next.X)*4 + int(next.Type.Pos.X)*3
		endY := int(next.Y)*4 + int(next.Type.Pos.Y)*3

		if startX != endX {
			for x := startX; x != endX; x += int(vertex.Type.Orientation.X) {
				image.SetRGBA(x, startY, cl)
			}
		} else {
			for y := startY; y != endY; y += int(vertex.Type.Orientation.Y) {
				image.SetRGBA(startX, y, cl)
			}
		}

		if bs.IsTail() {
			vertex, _ = bs.Next()
		} else {
			vertex = next
		}
	}
}
