package cli

import (
	"fmt"
	"io"
	"strings"
)

type NavNode struct {
	Id    string
	Link  string
	Child map[string]*NavNode
}

func (nn *NavNode) Get(path ...string) (target *NavNode) {
	for _, id := range path {
		if id == nn.Id {
			target = nn
			continue
		}
		if _, ok := target.Child[id]; !ok {
			target.Child[id] = &NavNode{Id: id, Child: make(map[string]*NavNode)}
		}
		target = target.Child[id]
	}
	return
}

func (nn *NavNode) Format(w io.Writer, depth int) {
	fmt.Fprintf(w, "%s* [%s](%s)\n", strings.Repeat(" ", depth*2), nn.Id, nn.Link)
	for _, n := range nn.Child {
		n.Format(w, depth+1)
	}
}
