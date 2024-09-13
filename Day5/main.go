package main

import (
	"awesomeProject2/Day5/geecache"
	"flag"
	"fmt"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *geecache.Group {
	return geecache.NewGroup("scores", 2<<10, geecache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist ", key)
		}))
}

// 用来启动缓存服务器：创建 HTTPPool，添加节点信息
// 注册到 gee 中，启动 HTTP 服务（共3个端口，8001/8002/8003），用户不感知
func startCacheServer(addr string, addrs []string, gee *geecache.Group) {
	peers := geecache.NewHTTPPool(addr)
	//将对应结点放入到哈希环上
	peers.Set(addrs...)
	//实现了一个多态，因为HHTTPPool实现了PeerPicker的方法
	gee.RegisterPeers(peers)
	log.Println("geecache is running at", addr)
	//进行监听，对应端口,开启服务
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

// APIServer用于与用户真正进行交互
func startAPIServer(apiAddr string, gee *geecache.Group) {
	//处理api这个接口上的所有内容,用于监听到对应内容所触发的回调
	http.Handle("api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key") //获取对应的key参数，即获取url上面的key参数
			//获取对应的缓存
			view, err := gee.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			//将缓存写入，返回给对应的客户
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
		}))
	log.Println("fontend server is running at", apiAddr)
	//7：通常表示去掉对应的http://这个前缀
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	var port int
	var api bool
	//定义一个整型的命令行标志。
	//&port: 指向一个整型变量的指针，用于存储解析后的值。
	//"port": 命令行中使用的标志名称。
	//8001: 默认值，如果用户没有提供该标志，则使用这个值。
	//"Geecache server port": 该标志的描述信息，通常用于帮助信息
	flag.IntVar(&port, "port", 8001, "Geecache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	//解析命令行参数。调用这个函数后，port 和 api 变量将被设置为用户在命令行中提供的值（如果有的话）。
	flag.Parse()
	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}
	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}
	gee := createGroup()
	if api {
		go startAPIServer(apiAddr, gee)
	}
	startCacheServer(addrMap[port], addrs, gee)
}
