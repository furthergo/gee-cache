package geecache

import (
	"fmt"
	"github.com/futhergo/gee-cache/consistenthash"
	"github.com/futhergo/gee-cache/geecachepb"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"log"
	"net/http"
	url2 "net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 100
)

type HTTPPool struct {
	self string // 当前节点名
	basePath string // api path
	mu *sync.Mutex
	peers *consistenthash.Map // 一致性hash
	httpGetters map[string]*httpGetter // 节点名和HTTPGetter的映射
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self: self,
		basePath: defaultBasePath,
		mu: &sync.Mutex{},
	}
}

func (p *HTTPPool)Log(format string, v...interface{}) {
	log.Printf("[Server %s]: %s", p.self, fmt.Sprintf(format, v...))
}

// 实现Handle接口，用于出来Http请求，只响应basePath即_geecache
func (p *HTTPPool)ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if !strings.HasPrefix(path, p.basePath) {
		panic(fmt.Sprintf("serve unexpect path %s", path))
	}
	// path : /_geecahce/<groupName>/<key>
	p.Log("%s %s\n", r.Method, path)
	parts := strings.SplitN(path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	gn := parts[0]
	k := parts[1]
	g, ok := GetGroup(gn)
	if !ok {
		http.Error(w, "error group request", http.StatusBadRequest)
		return
	}
	// 解析group和key，用key调用Get获取Value
	bv, err := g.Get(k)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Value通过pb编码写回http body
	pbResp := &geecachepb.GetResponse{
		Value: bv.ByteSlice(),
	}
	pbs, err := proto.Marshal(pbResp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(pbs)
}

// 给当前节点设置分布式节点
func (p *HTTPPool)Set(peers...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.NewMap(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{
			baseURL: peer + p.basePath,
		}
	}
}

// 根据key选择在一致性hash环上的真实节点名（非本机）
func (p *HTTPPool)PickPeer(k string) (PeerGetter, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if node := p.peers.Get(k); node != "" && node != p.self {
		p.Log("Pick peer %s", node)
		return p.httpGetters[node], nil
	}
	return nil, fmt.Errorf("empty peer selected")
}

type httpGetter struct {
	baseURL string
}

// 向远端节点发送get请求，unmarshal并返回
func (g *httpGetter)Get(group, k string) ([]byte, error) {
	url := fmt.Sprintf("%v%v/%v", g.baseURL, url2.QueryEscape(group), url2.QueryEscape(k))
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server return error status: %d", resp.StatusCode)
	}

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body error %v", err)
	}

	pbResp := &geecachepb.GetResponse{}
	err = proto.Unmarshal(bs, pbResp)
	return pbResp.Value, err
}