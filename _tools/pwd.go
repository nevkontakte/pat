package main

import (
	"flag"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var password = flag.String("password", "", "Password to hash")

func main() {
	flag.Parse()
	hash, _ := bcrypt.GenerateFromPassword([]byte(*password), 13)
	fmt.Println(string(hash))
}
