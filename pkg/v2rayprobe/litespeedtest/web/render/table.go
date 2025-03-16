package render

import (
	"sort"
)


type Node struct {
	Id       int    `json:"id"`
	Group    string `en:"Group" cn:"群组名" json:"group"`
	Remarks  string `en:"Remarks" cn:"备注" json:"remarks"`
	Protocol string `en:"Protocol" cn:"协议" json:"protocol"`
	Ping     string `en:"Ping" cn:"Ping" json:"ping"`
	AvgSpeed int64  `en:"AvgSpeed" cn:"平均速度" json:"avg_speed"`
	MaxSpeed int64  `en:"MaxSpeed" cn:"最大速度" json:"max_speed"`
	IsOk     bool   `json:"isok"`
	Traffic  int64  `json:"traffic"`
	Link     string `json:"link,omitempty"` // api only
}
type Nodes []Node

func (nodes Nodes) Sort(sortMethod string) {
	sort.Slice(nodes[:], func(i, j int) bool {
		switch sortMethod {
		case "speed":
			return nodes[i].MaxSpeed < nodes[j].MaxSpeed
		case "rspeed":
			return nodes[i].MaxSpeed > nodes[j].MaxSpeed
		case "ping":
			return nodes[i].Ping < nodes[j].Ping
		case "rping":
			return nodes[i].Ping > nodes[j].Ping
		default:
			return true
		}
	})
}