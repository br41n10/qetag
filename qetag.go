// Package qetag implements the qiniu etag as shown in
// https://github.com/qiniu/qetag/tree/master
//
// ==== 对小于 4MB 文件 ====
// 文件内容做 sha1 计算，结果前面拼接 0x16，
// 对生成 21 字节的二进制数据做 url_safe_base64 计算
//
// ==== 大于 4MB 的文件 ====
// 对内容按 4MB 进行拆分，依次计算每个块的 sha1 值
// 生成 []byte{block1sha1, block2sha1, ...}
// 对上面的 bytes 在进行依次 sha1 计算，并在前面拼接 0x96
// 对生成 21 字节的二进制数据做 url_safe_base64 计算
//
// 更形象化的说明参见：https://www.jianshu.com/p/3785fc314fc5

package qetag

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"io"
)

const (
	BLOCK_BITS = 22              // Indicate that the blocksize is 4M
	BLOCK_SIZE = 1 << BLOCK_BITS // 4Mb
)

type digest struct {
	len          int              // 记录当前一共写入了多少字节
	x            [BLOCK_SIZE]byte // 存放 Copy 进来的 bytes
	nx           int              // 游标，记录 x 中存到什么位置了
	sha1BlockBuf []byte           // 所有的块 sha1，挨着放在一起
}

func New() *digest {
	return &digest{}
}

// CalSha1 计算 r 的 sha1，并返回 []byte{b..., sha1(r)}
func CalSha1(b []byte, r io.Reader) ([]byte, error) {
	h := sha1.New()
	_, err := io.Copy(h, r)
	if err != nil {
		return nil, err
	}
	return h.Sum(b), nil
}

// gt4m 返回底层数据是否大于 4MB
func (d *digest) gt4m() bool {
	return d.len > 4*1024*1024
}

func (d *digest) Sum(b []byte) []byte {

	// 如果 d.x 还未计算，计算一下
	if d.nx > 0 {
		d.sha1BlockBuf, _ = CalSha1(d.sha1BlockBuf, bytes.NewReader(d.x[0:d.nx]))
	}

	if d.gt4m() {
		// 大于 4MB, 需要再计算 sha1BlockBuf 的 sha1 并拼接到 0x96 后面
		sha1Buf := make([]byte, 0, 21) // 最终结果
		sha1Buf = append(sha1Buf, 0x96)
		sha1Buf, _ = CalSha1(sha1Buf, bytes.NewReader(d.sha1BlockBuf))
		return sha1Buf
	} else {
		// 小于4MB, 直接将 sha1 值拼接到 0x16 后面
		return append([]byte{0x16}, d.sha1BlockBuf...)
	}
}

func (d *digest) Reset() {
	d.len = 0
	d.nx = 0
	d.sha1BlockBuf = []byte{}
}

func (d *digest) Size() int {
	return 21
}

func (d *digest) BlockSize() int {
	return 524288
}

func (d *digest) Etag() string {
	sum := d.Sum(nil)
	return base64.URLEncoding.EncodeToString(sum)
}

// Writer is the interface that wraps the basic Write method.
//
// Write writes len(p) bytes from p to the underlying data stream.
// It returns the number of bytes written from p (0 <= n <= len(p))
// and any error encountered that caused the write to stop early.
// Write must return a non-nil error if it returns n < len(p).
// Write must not modify the slice data, even temporarily.
//
// Implementations must not retain p.
func (d *digest) Write(p []byte) (nn int, err error) {

	// 我只需要不停的往 d.x 中 copy p
	// 每次计算好 d.nx 和 p 即可
	// 如果
	for len(p) > 0 {
		n := copy(d.x[d.nx:], p) //
		nn += n                  // 计数
		d.len += n               //
		d.nx += n                // 游标往后移动 n 个
		p = p[n:]                // 把这次计算的n个字节去掉

		if d.nx == BLOCK_SIZE { // 满了，计算后进入下一次循环
			d.sha1BlockBuf, err = CalSha1(d.sha1BlockBuf, bytes.NewReader(d.x[:]))
			if err != nil {
				return
			}
			d.nx = 0
			continue
		}
	}
	return
}
