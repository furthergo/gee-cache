package geecache

// 选择节点：根据key选择对应的peer
type PeerPicker interface {
	PickPeer(string) (PeerGetter, error)
}

// 从节点请求数据：根据group和key获取对应的ByteView
type PeerGetter interface {
	Get(group string, k string) ([]byte, error)
}

