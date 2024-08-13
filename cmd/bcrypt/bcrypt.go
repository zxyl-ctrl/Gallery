package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

// 简化的命令行界面，aka CLI

// Do
// 1. Hash a password
// 2. Compare a password with a hash to see if it is correct

// 当使用go run，会默认创建一个二进制文件并带有一个随机的哈希值

func main() {
	switch os.Args[1] {
	case "hash":
		hash(os.Args[2])
	case "compare":
		compare(os.Args[2], os.Args[3])
	default:
		fmt.Printf("Invalid command: %v\n", os.Args[1])
	}

}

func hash(password string) {
	// 当散列密码时，bcrypt每次会生成不同的salts，并且salts存储在生成的哈希值中,因此，每次使用生成哈希值的函数都可以生成不同的哈希值
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("error hashing: %v\n", err)
		return
	}
	hash := string(hashedBytes)
	fmt.Println(hash)
}

// 在Linux中，可以使用backslash \ 来告诉终端我们仍然在创建需要执行的命令
func compare(password, hash string) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		fmt.Printf("Password is invalid: %v\n", password)
		return
	}
	fmt.Println("Password is correct!")
}

// 注意加密后的哈希值需要使用单引号，否则无法识别
// go run cmd/bcrypt/bcrypt.go compare "secret password" '$2a$10$lhuPorKkcEvpwDY6kzFlHOJUYr39cDW3hZ5r/8NkBLuk8KaZqThIO'
