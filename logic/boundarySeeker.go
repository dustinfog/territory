package logic

import (
	"errors"
	"fmt"

	rbt "github.com/emirpasic/gods/trees/redblacktree"
)

type BoundarySeeker struct {
	xBaseYTree   map[int32]*rbt.Tree // {x => y => &code}
	yBaseXTree   map[int32]*rbt.Tree // {y => x => &code}
	mp           *Map
	head         *Vertex
	current      *Vertex
	pCurrentCode *int
}

func NewBoundarySeeker(m *Map, allianceId int32) *BoundarySeeker {
	xBaseYTree := make(map[int32]*rbt.Tree)
	yBaseXTree := make(map[int32]*rbt.Tree)

	flags := m.flags[allianceId]

	for f := range flags {
		flagVertexes := f.Vertexes

		for x, flagRow := range flagVertexes {
			yTree := xBaseYTree[x]
			if yTree == nil {
				yTree = rbt.NewWithIntComparator()
				xBaseYTree[x] = yTree
			}
			for y, code := range flagRow {
				fmt.Printf("Index: %d,%d,%d\n", x, y, code)
				code := code
				refCode := &code
				yTree.Put(int(y), refCode)

				xTree := yBaseXTree[y]
				if xTree == nil {
					xTree = rbt.NewWithIntComparator()
					yBaseXTree[y] = xTree
				}

				xTree.Put(int(x), refCode)
			}
		}
	}

	return &BoundarySeeker{
		xBaseYTree: xBaseYTree,
		yBaseXTree: yBaseXTree,
		mp:         m,
	}
}

func (bs *BoundarySeeker) Next() (next *Vertex, err error) {
	if len(bs.xBaseYTree) == 0 {
		if bs.head != nil {
			next = bs.head
			bs.head = nil
			bs.current = nil
			bs.pCurrentCode = nil
		}

		return
	}

	var pCode *int
	defer func() {
		if pCode != nil {
			*pCode ^= next.Type.Code

			if *pCode == 0 {
				bs.remove(next.X, next.Y)
			}
		}

		bs.current = next
		bs.pCurrentCode = pCode
	}()

	if bs.head == nil {
		next, pCode, err = bs.pickHead()
		return
	}

	vertex := bs.current
	if vertex.Type.IsOuter {
		next, pCode = bs.pickSelf()
		if next != nil {
			return
		}
	}

	orientation := vertex.Type.Orientation

	if orientation.X != 0 {
		next, pCode = bs.pickAlongX(orientation)
		if next != nil {
			return
		}
	} else {
		next, pCode = bs.pickAlongY(orientation)
		if next != nil {
			return
		}
	}

	next = bs.head
	bs.head = nil
	return
}

func (bs *BoundarySeeker) IsValid() bool {
	tile, _ := bs.mp.GetTile(bs.head.X, bs.head.Y, false)
	if tile != nil {
		return tile.IsValid()
	}

	return false
}

func (bs *BoundarySeeker) IsHead() bool {
	return bs.head == nil && bs.current == bs.head
}

func (bs *BoundarySeeker) IsTail() bool {
	return bs.head == nil && bs.current != nil
}

func (bs *BoundarySeeker) Finished() bool {
	return len(bs.xBaseYTree) == 0 && bs.current == nil
}

func (bs *BoundarySeeker) pickHead() (*Vertex, *int, error) {
	xBaseYTree := bs.xBaseYTree

	var vertex *Vertex
	var pCode *int

	for x, yTree := range xBaseYTree {
		it := yTree.Iterator()
		it.First()
		pCode = it.Value().(*int)

		vertex = &Vertex{
			X: x,
			Y: int32(it.Key().(int)),
		}
		break
	}

	if pCode == nil || vertex == nil {
		return nil, nil, errors.New("error")
	}

	for c, subVertex := range VertexTypes {
		if (*pCode)&c == c {
			vertex.Type = subVertex
			break
		}
	}

	bs.head = vertex
	return vertex, pCode, nil
}

func (bs *BoundarySeeker) pickSelf() (*Vertex, *int) {
	pCode := bs.pCurrentCode
	innerCode := bs.current.Type.ExpectedCode & *pCode
	if innerCode != 0 {
		nextVertexType := VertexTypes[innerCode]
		if nextVertexType.IsOuter {
			return &Vertex{
				X:    bs.current.X,
				Y:    bs.current.Y,
				Type: nextVertexType,
			}, pCode
		}
	}

	return nil, nil
}

func (bs *BoundarySeeker) pickAlongX(orientation *Orientation) (*Vertex, *int) {
	currentY := bs.current.Y
	xTree := bs.yBaseXTree[currentY]

	if xTree == nil {
		return nil, nil
	}

	var nextCode int
	var pCode *int
	var nextNode *rbt.Node

	expectedCode := bs.current.Type.ExpectedCode
	likelyClosed := currentY == bs.head.Y && expectedCode&bs.head.Type.Code != 0
	currentX := bs.current.X

	for nextCode == 0 {
		var found bool
		var nextX int32
		if orientation.X > 0 {
			nextNode, found = xTree.Ceiling(int(currentX) + 1)
			if !found {
				return nil, nil
			}
			nextX = int32(nextNode.Key.(int))
			if likelyClosed && currentX < bs.head.X && nextX > bs.head.X {
				return nil, nil
			}
		} else {
			nextNode, found = xTree.Floor(int(currentX) - 1)
			if !found {
				return nil, nil
			}
			nextX = int32(nextNode.Key.(int))
			if likelyClosed && currentX > bs.head.X && nextX < bs.head.X {
				return nil, nil
			}
		}

		pCode = nextNode.Value.(*int)
		nextCode = expectedCode & *pCode
		currentX = nextX
	}

	if nextNode != nil {
		return &Vertex{
			X:    currentX,
			Y:    currentY,
			Type: VertexTypes[nextCode],
		}, pCode
	}

	return nil, nil
}

func (bs *BoundarySeeker) pickAlongY(orientation *Orientation) (*Vertex, *int) {
	currentX := bs.current.X
	yTree := bs.xBaseYTree[currentX]

	if yTree == nil {
		return nil, nil
	}

	var nextCode int
	var pCode *int
	var nextNode *rbt.Node

	expectedCode := bs.current.Type.ExpectedCode
	likelyClosed := currentX == bs.head.X && expectedCode&bs.head.Type.Code != 0
	currentY := bs.current.Y

	for nextCode == 0 {
		var found bool
		var nextY int32
		if orientation.Y > 0 {
			nextNode, found = yTree.Ceiling(int(currentY) + 1)
			if !found {
				return nil, nil
			}
			nextY = int32(nextNode.Key.(int))
			if likelyClosed && currentY < bs.head.Y && nextY > bs.head.Y {
				return nil, nil
			}
		} else {
			nextNode, found = yTree.Floor(int(currentY) - 1)
			if !found {
				return nil, nil
			}
			nextY = int32(nextNode.Key.(int))
			if likelyClosed && currentY > bs.head.Y && nextY < bs.head.Y {
				return nil, nil
			}
		}

		pCode = nextNode.Value.(*int)
		nextCode = expectedCode & *pCode
		currentY = nextY
	}

	if nextNode != nil {
		return &Vertex{
			X:    currentX,
			Y:    currentY,
			Type: VertexTypes[nextCode],
		}, pCode
	}

	return nil, nil
}

func (bs *BoundarySeeker) remove(x int32, y int32) {
	xBaseYTree := bs.xBaseYTree
	yBaseXTree := bs.yBaseXTree
	xBaseYTree[x].Remove(int(y))
	if xBaseYTree[x].Empty() {
		delete(xBaseYTree, x)
	}
	yBaseXTree[y].Remove(int(x))
	if yBaseXTree[y].Empty() {
		delete(yBaseXTree, y)
	}
}
