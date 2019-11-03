package logic

//   |              |
// 0 |      N       | 1
//---|--------------|---
//   | 4          5 |
//   |              |
// W |     TILE     | E
//   |              |
//   | 7          6 |
//---+--------------+---
// 3 |      S       | 2
//   |              |

const (
	E  = 0
	SE = 1
	S  = 2
	SW = 3
	W  = 4
	NW = 5
	N  = 6
	NE = 7
)

const (
	VertexOuterNW = 1
	VertexOuterNE = 1 << 1
	VertexOuterSE = 1 << 2
	VertexOuterSW = 1 << 3
	VertexInnerNW = 1 << 4
	VertexInnerNE = 1 << 5
	VertexInnerSE = 1 << 6
	VertexInnerSW = 1 << 7
)

func init() {
	Orientations = make([]*Orientation, 8, 8)
	Orientations[E] = &Orientation{Vector2{1, 0}, E}
	Orientations[SE] = &Orientation{Vector2{1, 1}, SE}
	Orientations[S] = &Orientation{Vector2{0, 1}, S}
	Orientations[SW] = &Orientation{Vector2{-1, 1}, SW}
	Orientations[W] = &Orientation{Vector2{-1, 0}, W}
	Orientations[NW] = &Orientation{Vector2{-1, -1}, NW}
	Orientations[N] = &Orientation{Vector2{0, -1}, N}
	Orientations[NE] = &Orientation{Vector2{1, -1}, NE}

	VertexTypes = make(map[int]*VertexType)
	VertexTypes[VertexOuterNW] = &VertexType{
		Pos:          Vector2{0, 0},
		Orientation:  Orientations[E],
		IsOuter:      true,
		Code:         VertexOuterNW,
		ExpectedCode: VertexOuterNE | VertexInnerNW,
	}
	VertexTypes[VertexOuterNE] = &VertexType{
		Pos:          Vector2{1, 0},
		Orientation:  Orientations[S],
		IsOuter:      true,
		Code:         VertexOuterNE,
		ExpectedCode: VertexOuterSE | VertexInnerNE,
	}
	VertexTypes[VertexOuterSE] = &VertexType{
		Pos:          Vector2{1, 1},
		Orientation:  Orientations[W],
		IsOuter:      true,
		Code:         VertexOuterSE,
		ExpectedCode: VertexOuterSW | VertexInnerSE,
	}
	VertexTypes[VertexOuterSW] = &VertexType{
		Pos:          Vector2{0, 1},
		Orientation:  Orientations[N],
		IsOuter:      true,
		Code:         VertexOuterSW,
		ExpectedCode: VertexOuterNW | VertexInnerSW,
	}
	VertexTypes[VertexInnerNW] = &VertexType{
		Pos:          Vector2{0, 0},
		Orientation:  Orientations[N],
		IsOuter:      false,
		Code:         VertexInnerNW,
		ExpectedCode: VertexOuterNW | VertexInnerSW,
	}
	VertexTypes[VertexInnerNE] = &VertexType{
		Pos:          Vector2{1, 0},
		Orientation:  Orientations[E],
		IsOuter:      false,
		Code:         VertexInnerNE,
		ExpectedCode: VertexOuterNE | VertexInnerNW,
	}
	VertexTypes[VertexInnerSE] = &VertexType{
		Pos:          Vector2{1, 1},
		Orientation:  Orientations[S],
		IsOuter:      false,
		Code:         VertexInnerSE,
		ExpectedCode: VertexOuterSE | VertexInnerNE,
	}
	VertexTypes[VertexInnerSW] = &VertexType{
		Pos:          Vector2{0, 1},
		Orientation:  Orientations[W],
		IsOuter:      false,
		Code:         VertexInnerSW,
		ExpectedCode: VertexOuterSW | VertexInnerSE,
	}
}

var Orientations []*Orientation
var VertexTypes map[int]*VertexType
