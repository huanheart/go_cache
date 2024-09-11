package Day2

//Group 是 GeeCache 最核心的数据结构，负责与用户的交互，并且控制缓存值存储和获取的流程
type Group struct {
	name      string
	getter    Getter
	mainCache cache
}
