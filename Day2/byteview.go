package Day2

// 对这个Byte切片进行了一个封装，保证对应的数据不能被修改,即只读，不可修改
type ByteView struct {
	b []byte
}

// 封装对应的长度方法
func (v ByteView) Len() int {
	return len(v.b)
}

func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b) //复制对应的一个切片给到用户
}

// 封装一个string类型的
func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
