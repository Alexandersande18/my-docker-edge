package node

const (
	NodeStatusAlive = "alive"
	NodeStatusDead  = "dead"
)

type Node struct {
	NodeID  string
	LocalIP string
	Group   string
	Status  string
}
