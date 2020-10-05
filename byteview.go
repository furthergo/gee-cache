/*
实现Value接口，封装真实数据[]byte
 */
package geecache

type ByteView struct {
	b []byte
}

func (bv ByteView)Len() int {
	return len(bv.b)
}

func (bv ByteView)ByteSlice() []byte {
	return cloneBytes(bv.b)
}

func (bv ByteView)String() string {
	return string(bv.b)
}

func (bv ByteView)At(i int) byte {
	return bv.b[i]
}

func (bv ByteView)Slice(f, t int) ByteView {
	return ByteView{
		bv.b[f:t],
	}
}

func (bv ByteView)SliceFrom(f int) ByteView {
	return ByteView{
		bv.b[f:],
	}
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}