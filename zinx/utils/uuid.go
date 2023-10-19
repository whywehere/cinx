package utils

import (
	"math/rand"
	"time"
)

func RandomID() uint32 {
	// 设置随机数种子，以确保每次生成的ID都不同
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// 生成一个随机的整数作为ID
	return uint32(rand.Intn(1000000))
}
