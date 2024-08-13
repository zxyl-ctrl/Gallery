package main

import (
	"crypto/rand"
	"fmt"
)

func main() {
	n := 8
	b := make([]byte, n)
	nRead, err := rand.Read(b) // 根据字节数生成随机值
	if err != nil {
		panic(err)
	}
	if nRead < n {
		panic("didn't reaad enough random bytes")
	}
	fmt.Println(b)
}

// 字符串使用字符编码，UTF-8，这表明在打印时需要
// math.rand不是真正的随机(即使使用种子也不是)，crypto.rand是真正的随机
