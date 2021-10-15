package main

import (
    "fmt"
    "flag"
)

type node struct {
    value int
    left *node
    right *node
}

type bst struct {
    root *node
    len int
}

func main() {
    fmt.Println("Running go BST sequential")
    
    /*******   PARSING ARGUMENTS   *******/
    
    hash_workers := flag.Int("hash-workers", 1, "an int")
    data_workers := flag.Int("data-workers", 1, "an int")
    comp_workers := flag.Int("comp-workers", 1, "an int")
    input_file   := flag.String("input", "", "a string")
    
    flag.Parse()
    
    fmt.Println("hash workers:", *hash_workers)
    fmt.Println("data workers:", *data_workers)
    fmt.Println("comp workers:", *comp_workers)
    fmt.Println("input file:", *input_file)
    
}