package render

import (
	"fmt"
	"sort"
)


type Node struct {
	Id       int    `json:"id"`
	Group    string `en:"Group" json:"group"`
	Remarks  string `en:"Remarks" json:"remarks"`
	Protocol string `en:"Protocol" json:"protocol"`
	Ping     string `en:"Ping" json:"ping"`
	AvgSpeed int64  `en:"AvgSpeed" json:"avg_speed"`
	MaxSpeed int64  `en:"MaxSpeed" json:"max_speed"`
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

func (nodes Nodes) ChangeRemarks(remarkPrefix string) {
	for _, node := range nodes {node.Remarks = fmt.Sprintf("%s-%d", node.Id, remarkPrefix)}
}