package logic

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"sort"
	"time"

	rbt "github.com/emirpasic/gods/trees/redblacktree"
)

const FlagHalfLength int32 = 7

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
	orientations = make([]*Orientation, 8, 8)
	orientations[E] = &Orientation{Vector2{1, 0}, E}
	orientations[SE] = &Orientation{Vector2{1, 1}, SE}
	orientations[S] = &Orientation{Vector2{0, 1}, S}
	orientations[SW] = &Orientation{Vector2{-1, 1}, SW}
	orientations[W] = &Orientation{Vector2{-1, 0}, W}
	orientations[NW] = &Orientation{Vector2{-1, -1}, NW}
	orientations[N] = &Orientation{Vector2{0, -1}, N}
	orientations[NE] = &Orientation{Vector2{1, -1}, NE}

	vertexTypes = make(map[int]*VertexType)
	vertexTypes[VertexOuterNW] = &VertexType{
		Pos:          Vector2{0, 0},
		Orientation:  orientations[E],
		IsOuter:      true,
		Code:         VertexOuterNW,
		ExpectedCode: VertexOuterNE | VertexInnerNW,
	}
	vertexTypes[VertexOuterNE] = &VertexType{
		Pos:          Vector2{1, 0},
		Orientation:  orientations[S],
		IsOuter:      true,
		Code:         VertexOuterNE,
		ExpectedCode: VertexOuterSE | VertexInnerNE,
	}
	vertexTypes[VertexOuterSE] = &VertexType{
		Pos:          Vector2{1, 1},
		Orientation:  orientations[W],
		IsOuter:      true,
		Code:         VertexOuterSE,
		ExpectedCode: VertexOuterSW | VertexInnerSE,
	}
	vertexTypes[VertexOuterSW] = &VertexType{
		Pos:          Vector2{0, 1},
		Orientation:  orientations[N],
		IsOuter:      true,
		Code:         VertexOuterSW,
		ExpectedCode: VertexOuterNW | VertexInnerSW,
	}
	vertexTypes[VertexInnerNW] = &VertexType{
		Pos:          Vector2{0, 0},
		Orientation:  orientations[N],
		IsOuter:      false,
		Code:         VertexInnerNW,
		ExpectedCode: VertexOuterNW | VertexInnerSW,
	}
	vertexTypes[VertexInnerNE] = &VertexType{
		Pos:          Vector2{1, 0},
		Orientation:  orientations[E],
		IsOuter:      false,
		Code:         VertexInnerNE,
		ExpectedCode: VertexOuterNE | VertexInnerNW,
	}
	vertexTypes[VertexInnerSE] = &VertexType{
		Pos:          Vector2{1, 1},
		Orientation:  orientations[S],
		IsOuter:      false,
		Code:         VertexInnerSE,
		ExpectedCode: VertexOuterSE | VertexInnerNE,
	}
	vertexTypes[VertexInnerSW] = &VertexType{
		Pos:          Vector2{0, 1},
		Orientation:  orientations[W],
		IsOuter:      false,
		Code:         VertexInnerSW,
		ExpectedCode: VertexOuterSW | VertexInnerSE,
	}
}

type Vector2 struct {
	X int32
	Y int32
}

var orientations []*Orientation

type Orientation struct {
	Vector2
	Code int
}

func (o *Orientation) Right() *Orientation {
	return o.rotate(2)
}

func (o *Orientation) Back() *Orientation {
	return o.rotate(4)
}

func (o *Orientation) Left() *Orientation {
	return o.rotate(6)
}

// 以45度为单位顺时针旋转times次
func (o *Orientation) rotate(times int) *Orientation {
	if times < 0 {
		times = 8 - ((-times) % 8)
	}
	return orientations[(o.Code+times)%8]
}

var vertexTypes map[int]*VertexType

type VertexType struct {
	Pos          Vector2      //顶点在Tile内的位置
	Orientation  *Orientation //绘制走向
	IsOuter      bool         //外圈或内圈
	ExpectedCode int          //与之匹配的终点类型编码
	Code         int          //类型编码
}

type Tile struct {
	Vector2
	ownerFlag *Flag
}

func (t *Tile) GetAllianceId() int32 {
	if t.ownerFlag == nil {
		return 0
	}

	return t.ownerFlag.AllianceId
}

func (t *Tile) SetOwnerFlag(flag *Flag) {
	t.ownerFlag = flag
	flag.SetTileBit(t)
}

func (t *Tile) OwnerFlag() *Flag {
	return t.ownerFlag
}

func (t *Tile) IsValid() bool {
	return t.ownerFlag != nil && t.ownerFlag.IsValid
}

func (t *Tile) IsVertex() (bool, int) {
	return t.ownerFlag.IsVertex(t.X, t.Y)
}

func (t *Tile) IsEmpty() bool {
	//todo: 临时的
	return t.ownerFlag.Tile != t
}

type Flag struct {
	ID         int32
	AllianceId int32
	Tile       *Tile
	IsFortress bool
	Map        *Map
	IsValid    bool
	Neighbors  map[*Flag]*Flag //同盟的相邻旗子
	Overlaps   map[*Flag]*Flag //相交叠的旗子
	Bitmap     []int16
	Vertexes   map[int32]map[int32]int //联盟领地顶点
	MTime      time.Time
}

func NewFlag(x int32, y int32, allianceId int32, isFortress bool, mp *Map, mtime time.Time) *Flag {
	t, _ := mp.GetTile(x, y, true)

	f := &Flag{
		AllianceId: allianceId,
		Tile:       t,
		IsFortress: isFortress,
		Map:        mp,
		Neighbors:  make(map[*Flag]*Flag),
		Overlaps:   make(map[*Flag]*Flag),
		Bitmap:     make([]int16, FlagHalfLength*2+1, FlagHalfLength*2+1),
		MTime:      mtime,
	}
	t.SetOwnerFlag(f)
	return f
}

func (f *Flag) AddNeighbor(nf *Flag) {
	if f == nf || f.Neighbors[nf] == nf {
		return
	}
	f.Neighbors[nf] = nf
	nf.Neighbors[f] = f
}

func (f *Flag) RemoveNeighbor(nf *Flag) {
	if f.Neighbors[nf] == nil {
		return
	}

	delete(f.Neighbors, nf)
	delete(nf.Neighbors, f)
}

func (f *Flag) SetTileBit(tile *Tile) {
	r := tile.Y - f.Tile.Y + FlagHalfLength
	c := tile.X - f.Tile.X + FlagHalfLength
	f.Bitmap[r] |= 1 << uint(c)
}

func (f *Flag) ResetVertex() {
	f.Vertexes = make(map[int32]map[int32]int)
}

func (f *Flag) SetVertex(x int32, y int32, code int) {
	if code == 0 {
		return
	}

	row := f.Vertexes[x]
	if row == nil {
		row = make(map[int32]int)
		f.Vertexes[x] = row
	}

	row[y] = code

	fmt.Printf("Vertex: %d - %d:%d = %d\n", f.AllianceId, x, y, code)
}

func (f *Flag) IsVertex(x int32, y int32) (bool, int) {
	r := f.Vertexes[x]
	if r == nil {
		return false, 0
	}

	code, ok := r[y]
	return ok, code
}

func (f *Flag) CalcVertexes() {
	f.ResetVertex()

	m := f.Map
	for j, bitmap := range f.Bitmap {
		if bitmap == 0 {
			continue
		}

		for i := int32(0); i <= 2*FlagHalfLength; i++ {
			if (bitmap>>uint(i))&1 == 0 {
				continue
			}

			x := i + f.Tile.X - FlagHalfLength
			y := int32(j) + f.Tile.Y - FlagHalfLength
			tile, _ := m.GetTile(x, y, false)
			code := m.CalcVertexCode(tile)
			if code != 0 {
				f.SetVertex(x, y, code)
			}
		}
	}
}

func (f *Flag) AddOverlap(of *Flag) {
	if f == of || f.Overlaps[of] == of {
		return
	}

	f.Overlaps[of] = of
	of.Overlaps[f] = f
}

func (f *Flag) RemoveOverlap(of *Flag) {
	if f.Overlaps[of] == nil {
		return
	}

	delete(f.Overlaps, of)
	delete(of.Overlaps, f)
}

func (f *Flag) GetTiles() []*Tile {
	return nil
}

type Vertex struct {
	X    int32
	Y    int32
	Type *VertexType
}

type Map struct {
	tiles      map[int32]map[int32]*Tile
	fortresses map[int32]map[*Flag]*Flag
	flags      map[int32]map[*Flag]*Flag
}

func NewMap() *Map {
	return &Map{
		tiles:      make(map[int32]map[int32]*Tile),
		flags:      make(map[int32]map[*Flag]*Flag),
		fortresses: make(map[int32]map[*Flag]*Flag),
	}
}

func (m *Map) GetTile(x int32, y int32, createIfAbsent bool) (*Tile, bool) {
	row, ok := m.tiles[x]
	if row == nil {
		if !createIfAbsent {
			return nil, false
		}

		row = make(map[int32]*Tile)
		m.tiles[x] = row
	}

	t, ok := row[y]
	if t == nil && createIfAbsent {
		t = &Tile{
			Vector2: Vector2{
				x,
				y,
			},
		}

		row[y] = t
	}

	return t, ok
}

func (m *Map) removeTile(x int32, y int32) (*Tile, bool) {
	row, ok := m.tiles[x]
	if !ok {
		return nil, ok
	}

	t, ok := row[y]
	if !ok {
		return nil, ok
	}

	if len(row) == 1 {
		delete(m.tiles, x)
	} else {
		delete(row, y)
	}

	return t, ok
}

func (m *Map) CalcVertexCode(t *Tile) int {
	surround := make([]bool, 8, 8)
	for i, o := range orientations {
		x := t.X + o.X
		y := t.Y + o.Y

		tile, ex := m.GetTile(x, y, false)
		if ex && tile.GetAllianceId() == t.GetAllianceId() {
			surround[i] = true
		} else {
			surround[i] = false
		}
	}

	code := 0
	for subCode, vertexType := range vertexTypes {
		var orientation = vertexType.Orientation
		if vertexType.IsOuter {
			if !surround[orientation.Back().Code] && !surround[orientation.Left().Code] {
				code |= subCode
			}
		} else {
			if surround[orientation.Code] && !surround[orientation.rotate(-1).Code] && surround[orientation.rotate(-2).Code] {
				code |= subCode
			}
		}
	}

	return code
}

func (m *Map) AddFlag(x int32, y int32, allianceId int32, isFortress bool, tm time.Time) (*Flag, error) {
	t, ex := m.GetTile(x, y, false)

	if ex {
		if t.GetAllianceId() != allianceId {
			return nil, errors.New("not yours")
		}

		if !t.IsEmpty() {
			return nil, errors.New("occupied")
		}
	}

	if !isFortress && !m.checkFlagSettable(x, y, allianceId) {
		return nil, errors.New("no neighbor")
	}

	f := NewFlag(x, y, allianceId, isFortress, m, tm)

	flags := m.flags[f.AllianceId]
	if flags == nil {
		flags = make(map[*Flag]*Flag)
		m.flags[f.AllianceId] = flags
	}

	flags[f] = f

	if isFortress {
		fortresses := m.fortresses[f.AllianceId]

		if fortresses == nil {
			fortresses = make(map[*Flag]*Flag)
			m.fortresses[f.AllianceId] = fortresses
		}

		fortresses[f] = f
		f.IsValid = true
	}

	m.scanFlagArea(f)

	f.CalcVertexes()
	for neighbor := range f.Neighbors {
		neighbor.CalcVertexes()
	}

	m.scanAllianceArea(allianceId)

	return f, nil
}

func (m *Map) RemoveFlag(flag *Flag) {
	delete(m.flags[flag.AllianceId], flag)
	if len(m.flags[flag.AllianceId]) == 0 {
		delete(m.flags, flag.AllianceId)
	}
	if flag.IsFortress {
		delete(m.fortresses[flag.AllianceId], flag)
		if len(m.fortresses[flag.AllianceId]) == 0 {
			delete(m.fortresses, flag.AllianceId)
		}
	}

	for i := -FlagHalfLength; i <= FlagHalfLength; i++ {
		x := flag.Tile.X + i
		for j := -FlagHalfLength; j <= FlagHalfLength; j++ {
			y := flag.Tile.Y + j
			tile, ex := m.GetTile(x, y, false)
			if ex && tile.OwnerFlag() == flag {
				m.removeTile(x, y)
			}
		}
	}

	vertexDirty := make(map[*Flag]*Flag)

	for neighbor := range flag.Neighbors {
		neighbor.RemoveNeighbor(flag)
		vertexDirty[neighbor] = neighbor
	}
	allianceIds := make(map[int32]int32)

	sorter := &OverlapSorter{
		Flags: make([]*Flag, 0, len(flag.Overlaps)),
	}
	sorter.AddAll(flag.Overlaps)
	sort.Sort(sorter)

	for _, overlap := range sorter.Flags {
		overlap.RemoveOverlap(flag)
		m.scanFlagArea(overlap)
		allianceIds[overlap.AllianceId] = overlap.AllianceId
		vertexDirty[overlap] = overlap
	}

	for _, f1 := range vertexDirty {
		f1.CalcVertexes()
	}

	for allianceId := range allianceIds {
		m.scanAllianceArea(allianceId)
	}
}

func (m *Map) checkFlagSettable(x int32, y int32, allianceId int32) bool {
	minX := x - FlagHalfLength
	maxX := x + FlagHalfLength
	minY := y - FlagHalfLength
	maxY := y + FlagHalfLength

	type TileListNode struct {
		X    int32
		Y    int32
		Next *TileListNode
	}

	head := &TileListNode{
		X: x,
		Y: y,
	}
	tail := head

	marked := make(map[int32]map[int32]bool)

	scan := func(x int32, y int32) bool {
		tile, ex := m.GetTile(x, y, false)
		if ex && tile.GetAllianceId() == allianceId {
			return true
		}

		if m.markCoordinate(marked, x, y) && !ex && x >= minX && x <= maxX && y >= minY && y <= maxY {
			next := &TileListNode{
				X: x,
				Y: y,
			}
			tail.Next = next
			tail = next
		}

		return false
	}

	m.markCoordinate(marked, x, y)

	for head != nil {
		if head.X >= minX && scan(head.X-1, head.Y) {
			return true
		}

		if head.X <= maxX && scan(head.X+1, head.Y) {
			return true
		}

		if head.Y >= minY && scan(head.X, head.Y-1) {
			return true
		}

		if head.Y <= maxY && scan(head.X, head.Y+1) {
			return true
		}

		head = head.Next
	}

	return false
}

func (m *Map) scanFlagArea(flag *Flag) {
	t := flag.Tile
	x, y := t.X, t.Y
	minX := x - FlagHalfLength
	maxX := x + FlagHalfLength
	minY := y - FlagHalfLength
	maxY := y + FlagHalfLength

	type TileListNode struct {
		Tile *Tile
		Next *TileListNode
	}

	head := &TileListNode{
		Tile: t,
	}
	tail := head

	marked := make(map[int32]map[int32]bool)

	scan := func(x int32, y int32) {
		if !m.markCoordinate(marked, x, y) {
			return
		}

		tile, ex := m.GetTile(x, y, true)
		if ex {
			flag.AddOverlap(tile.OwnerFlag())

			if tile.GetAllianceId() != flag.AllianceId {
				return
			}

			flag.AddNeighbor(tile.OwnerFlag())
		} else {
			tile.SetOwnerFlag(flag)
		}

		next := &TileListNode{
			Tile: tile,
		}
		tail.Next = next
		tail = next
	}

	checkNeighbor := func(x int32, y int32) bool {
		tile, ex := m.GetTile(x, y, false)
		if ex && tile.GetAllianceId() == flag.AllianceId {
			flag.AddNeighbor(tile.OwnerFlag())
			return true
		}

		return false
	}

	m.markCoordinate(marked, x, y)
	for head != nil {
		tile := head.Tile

		if tile.X > minX {
			scan(tile.X-1, tile.Y)
		} else if tile.X == minX {
			checkNeighbor(tile.X-1, tile.Y)
		}

		if tile.X < maxX {
			scan(tile.X+1, tile.Y)
		} else if tile.X == maxX {
			checkNeighbor(tile.X+1, tile.Y)
		}

		if tile.Y > minY {
			scan(tile.X, tile.Y-1)
		} else if tile.Y == minY {
			checkNeighbor(tile.X, tile.Y-1)
		}

		if tile.Y < maxY {
			scan(tile.X, tile.Y+1)
		} else if tile.Y == minY {
			checkNeighbor(tile.X, tile.Y+1)
		}

		head = head.Next
	}

}

func (m *Map) markCoordinate(marked map[int32]map[int32]bool, x int32, y int32) bool {
	row := marked[x]
	exists := false
	if row == nil {
		row = make(map[int32]bool)
	}

	exists = row[y]
	if !exists {
		row[y] = true
		marked[x] = row
	}
	return !exists
}

func (m *Map) scanAllianceArea(allianceId int32) {
	type FlagListNode struct {
		Flag *Flag
		Next *FlagListNode
	}

	marked := make(map[*Flag]*Flag)

	for flag := range m.fortresses[allianceId] {
		if marked[flag] == flag {
			continue
		}

		head := &FlagListNode{
			Flag: flag,
		}
		tail := head

		for head != nil {
			f := head.Flag
			marked[f] = f

			for neighbor := range f.Neighbors {
				if marked[neighbor] == neighbor {
					continue
				}

				next := &FlagListNode{
					Flag: neighbor,
				}
				tail.Next = next
				tail = next
			}

			f.IsValid = true
			head = head.Next
		}
	}

	for flag := range m.flags[allianceId] {
		if marked[flag] == nil {
			flag.IsValid = false
		}
	}
}

func (m *Map) DrawImage(image *image.RGBA) {
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

func (m *Map) DrawBoundaries(image *image.RGBA, allianceId int32) {

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

type OverlapSorter struct {
	Flags []*Flag
}

func (f *OverlapSorter) AddAll(flags map[*Flag]*Flag) {
	for flag := range flags {
		f.Flags = append(f.Flags, flag)
	}
}

func (f *OverlapSorter) Len() int {
	return len(f.Flags)
}

func (f *OverlapSorter) Less(i, j int) bool {
	return f.Flags[i].MTime.UnixNano() < f.Flags[j].MTime.UnixNano()
}

func (f *OverlapSorter) Swap(i, j int) {
	fi := f.Flags[i]
	f.Flags[i] = f.Flags[j]
	f.Flags[j] = fi
}

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
		*pCode ^= next.Type.Code

		if *pCode == 0 {
			bs.remove(next.X, next.Y)
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

	for c, subVertex := range vertexTypes {
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
		nextVertexType := vertexTypes[innerCode]
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
			Type: vertexTypes[nextCode],
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
			Type: vertexTypes[nextCode],
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
