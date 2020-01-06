package main

import (
	"fmt"
	"github.com/yenkeia/mirgo/common"
	"sync"
)

// Map ...
type Map struct {
	Env    *Environ
	Width  uint16 // 测试用
	Height uint16 // 测试用
	Info   *common.MapInfo
	AOI    *AOIManager
	Cells  *sync.Map // key=Cell.Coordinate  value=*Cell
}

func (m *Map) Submit(t *Task) {
	m.Env.Game.Pool.EntryChan <- t
}

func (m *Map) GetAllPlayers() []*Player {
	players := make([]*Player, 0)
	m.AOI.grids.Range(func(k, v interface{}) bool {
		g := v.(*Grid)
		g.Players.Range(func(k, v interface{}) bool {
			players = append(players, v.(*Player))
			return true
		})
		return true
	})
	return players
}

// Broadcast send message to all players in this map
func (m *Map) Broadcast(msg interface{}) {
	players := m.GetAllPlayers()
	for i := range players {
		players[i].Enqueue(msg)
	}
}

func (m *Map) GetCell(coordinate string) *Cell {
	v, ok := m.Cells.Load(coordinate)
	if !ok {
		return nil
	}
	return v.(*Cell)
}

func (m *Map) AddObject(obj interface{}) {
	switch o := obj.(type) {
	case *Player:
		coordinate := o.Point().Coordinate()
		grid := m.AOI.GetGridByCoordinate(coordinate)
		grid.AddPlayer(o)
		m.GetCell(o.Point().Coordinate()).SetObject(o)
	case *NPC:
		coordinate := o.Point().Coordinate()
		grid := m.AOI.GetGridByCoordinate(coordinate)
		grid.AddNPC(o)
		m.GetCell(o.Point().Coordinate()).SetObject(o)
	case *Monster:
		coordinate := o.Point().Coordinate()
		grid := m.AOI.GetGridByCoordinate(coordinate)
		grid.AddMonster(o)
		m.GetCell(o.Point().Coordinate()).SetObject(o)
	}
}

// UpdateObject 更新对象在 Cells, AOI 中的数据, 如果更新成功返回 true
func (m *Map) UpdateObject(obj interface{}, points ...*common.Point) bool {
	switch o := obj.(type) {
	case *Player:
		for i := range points {
			c := m.GetCell(points[i].Coordinate())
			if c == nil || !c.IsValid() {
				return false
			}
		}
		old := m.GetCell(o.Point().Coordinate())
		old.SetObject(nil)
		last := m.GetCell(points[len(points)-1].Coordinate())
		last.SetObject(o)
		// TODO change AOI
		return true
	case *NPC:
	case *Monster:

	}
	return false
}

// InitNPCs 初始化地图上的 NPC
func (m *Map) InitNPCs() error {
	for _, ni := range m.Env.GameDB.NpcInfos {
		ni := ni
		if ni.MapID == m.Info.ID {
			m.AddObject(NewNPC(m, &ni))
		}
	}
	return nil
}

// InitMonsters 初始化地图上的怪物
func (m *Map) InitMonsters() error {
	for _, ri := range m.Env.GameDB.RespawnInfos {
		if ri.MapID == m.Info.ID {
			r, err := NewRespawn(m, &ri)
			if err != nil {
				return err
			}
			for _, a := range r.AliveMonster {
				a := a
				m.AddObject(a)
			}
		}
	}
	return nil
}

// GetValidPoint
func (m *Map) GetValidPoint(x int, y int, spread int) (*common.Point, error) {
	if spread == 0 {
		//log.Debugf("GetValidPoint: (x: %d, y: %d), spread: %d\n", x, y, spread)
		c := m.GetCell(common.Point{X: uint32(x), Y: uint32(y)}.Coordinate())
		if c != nil && c.IsValid() {
			return common.NewPointByCoordinate(c.Coordinate), nil
		}
		return nil, fmt.Errorf("GetValidPoint: (x: %d, y: %d), spread: %d\n", x, y, spread)
	}
	minX := x - spread
	maxX := x + spread
	minY := y - spread
	maxY := y + spread
	//log.Debugf("(%d,%d,%d)(%d,%d,%d,%d)\n", x, y, spread, minX, maxX, minY, maxY)
	cnt := 0
	for {
		if cnt == 100 {
			return nil, fmt.Errorf("no valid point in (%d,%d) spread: %d", x, y, spread)
		}
		tryX := G_Rand.RandInt(minX, maxX)
		tryY := G_Rand.RandInt(minY, maxY)
		c := m.GetCell(common.Point{X: uint32(tryX), Y: uint32(tryY)}.Coordinate())
		if c != nil && c.IsValid() {
			return common.NewPointByCoordinate(c.Coordinate), nil
		}
		cnt += 1
	}
}

func (m *Map) GetNextCell(cell *Cell, direction common.MirDirection, step uint32) *Cell {
	p := cell.Point().NextPoint(direction, step)
	return m.GetCell(p.Coordinate())
}
