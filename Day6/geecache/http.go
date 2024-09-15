package geecache

import (
	"awesomeProject2/Day5/geecache/consistenthash"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)

// 这里使用http的方法，任何类型都可以实现对应go中http包的接口
type HTTPPool struct {
	self        string
	basePath    string //默认这个http服务所监听的对应接口路径是basePath
	mu          sync.Mutex
	peers       *consistenthash.Map    //对应的一致性哈希的map，用来根据具体的key选择对应的结点
	httpGetters map[string]*httpGetter //映射远程节点与对应的 httpGetter。每一个远程节点对应一个 httpGetter
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		//self保留自己的地址
		self:     self,
		basePath: defaultBasePath,
	}
}

// 对日志的进行一个封装,参数为一个接口，表示任何值都可以传递进来
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s ", p.self, fmt.Sprintf(format, v...)) //将对应的任意类型合并到format上,即相当于c++中的snprintf这个类型
}

//处理对应服务的函数

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path :" + r.URL.Path)
	}
	p.Log("%s %s ", r.Method, r.URL.Path)
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	//举例：http://localhost:9999/_geecache/scores/Tom这个url,
	// 会变成这个/scores/Tom,然后通过分割有两个对应的字符串scores,Tom
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest) //返回400
		return
	}
	groupName := parts[0]
	key := parts[1]
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}
	//本地方法组找到对应的缓存，如果没有内部会根据回调函数返回的数据返回对应的数据，然后将其数据放入到对应的缓存结构中
	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	//这意味着你告诉客户端（例如浏览器），返回的数据是二进制流，而不是特定格式的文本或其他类型的数据。这通常用于文件下载或传输未知类型的数据。
	w.Header().Set("Content-Type", "application/octet-stream")
	//复制一个切片给到用户
	//w.Write() 方法将这个字节切片的内容写入到 HTTP 响应体中。
	w.Write(view.ByteSlice())

}

// 表示将要访问的远程节点的地址，例如 http://example.com/_geecache/
type httpGetter struct {
	baseURL string
}

// 发送方法，并接收返回值进行返回
// baseURL 表示将要访问的远程节点的地址
func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v/%v", //这里 /不要漏掉了
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	//阻塞调用Get方法
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	//关闭方法体
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %v", res.Status)
	}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body %v", err)
	}
	return bytes, nil
}

// 定义一个没有用的对象，查看当前类型可以创建，即所有接口是否被正确实现
// 若没实现，这里就会报错，很常见的一种设计模式
var _ PeerGetter = (*httpGetter)(nil)

// 将一些真实结点进行设置，有种分布式存储那个项目地感觉，
//每个机器都有着其他结点的信息，即peer数组
// 以及通信的通道
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	//进行初始化map
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	//真正开始存入其他机器的信息
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}

}

// 这个函数应该是查找对应key存放在哪一个机器上，然后通过调用远程方法去获取这个缓存
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	//并不等于自己而且不能为空，那么就表示映射成功
	//通过哈希映射查询到对应结点应该存放到哪里
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil)
