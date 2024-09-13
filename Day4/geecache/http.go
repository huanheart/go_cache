package geecache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const defaultBasePath = "/_geecache/"

// 这里使用http的方法，任何类型都可以实现对应go中http包的接口
type HTTPPool struct {
	self     string
	basePath string //默认这个http服务所监听的对应接口路径是basePath
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
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
	//找到对应的缓存，如果没有内部会根据回调函数返回的数据返回对应的数据，然后将其数据放入到对应的缓存结构中
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
