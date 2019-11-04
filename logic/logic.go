package logic

import (
	"errors"
	"fmt"
	"sort"
	"time"
)

const FlagHalfLength int32 = 7

type Vector2 struct {
	X int32
	Y int32
}

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
	return Orientations[(o.Code+times)%8]
}

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

func (t *Tile) IsFlag() bool {
	return t.ownerFlag.Tile == t
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
	for i, o := range Orientations {
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
	for subCode, vertexType := range VertexTypes {
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
		} else if tile.Y == maxY {
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
