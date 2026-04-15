package main

import (
    "fmt"
    goversion "github.com/cumulus13/go-version"
)

func main() {
    info, err := goversion.Get()
    if err != nil {
        panic(err)
    }
    fmt.Println(info)
    v, err := goversion.Get()
    fmt.Println("version:", v.Authors[0])
}