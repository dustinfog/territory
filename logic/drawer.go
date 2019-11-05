package logic

import (
	"fmt"
	"image"
	"image/color"
)

func DrawImage(m *Map, image *image.RGBA, colors map[int32]color.RGBA) {
	for x, row := range m.tiles {
		for y, tile := range row {
			isV := tile.IsFlag()
			cl := colors[tile.GetAllianceId()]
			if isV {
				cl = color.RGBA{
					0, 0, 0, 255,
				}
			} else {
				cl.A = 8
			}
			for i := 0; i < 4; i++ {
				for j := 0; j < 4; j++ {
					image.SetRGBA(int(x)*4+i, int(y)*4+j, cl)
				}
			}
		}
	}
}

func DrawBoundaries(m *Map, image *image.RGBA, colors map[int32]color.RGBA) {
	for allianceId, _ := range m.fortresses {
		cl := colors[allianceId]

		bs := NewBoundarySeeker(m, allianceId)
		vertex, _ := bs.Next()

		for !bs.Finished() {
			fmt.Printf("draw %d:%d\n", vertex.X, vertex.Y)
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
}
