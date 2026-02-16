package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "EdunexAdmin2026@@" // Troque pela senha que você quer
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	fmt.Println(string(hash))
}
