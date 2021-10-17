package main

import (
    "fmt"
    "flag"
)

func checkfile(e error) {
    if e != nil {
        panic(e)
    }
}

type Node struct {
    Value int
    Left *Node
    Right *Node
}


func (n *Node) computeHash() int {
    var hash int = 1;
    var result[] int;
    var count int = 0;
    for _, value := range n.in_order_traversal(result, &count) {
        var new_value int = value + 2;
        hash = (hash * new_value + new_value) % 1000
    }
    return hash
}

func (n *Node) in_order_traversal(result []int, count *int) []int {
    
    if (n != nil){
        result = n.Left.in_order_traversal(result, count)
        //result[*count] = n.Value
        result = append(result, n.Value)
        (*count)++ //unnecessary but leave it, may be helpeful in optimization
        result = n.Right.in_order_traversal(result, count)
    }
    return result
}


func (n *Node) Insert (value int){
    if (value > n.Value) {
        //MOVE RIGHT, value larger
        if n.Right==nil {
            //NO RIGHT CHILD, insert
            n.Right = &Node{Value : value}
        } else {
            //repeat same method on right child
            n.Right.Insert(value)
        }
    } else if (value < n.Value) {
        if n.Left==nil {
            //NO RIGHT CHILD, insert
            n.Left = &Node{Value : value}
        } else {
            //repeat same method on right child
            n.Left.Insert(value)
        }
    } else {
        //ALREADY EXISTS
        fmt.Println("already exists")
    }
}
    

func (n *Node) Search (value int) bool {
    
    if (n==nil){
        return false
    }
    
    if (value > n.Value) {
        //MOVE RIGHT, value larger
        return n.Right.Search(value)
    } else if (value < n.Value) {
        return n.Left.Search(value)
    }
    return true
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
    
    //data, err := os.ReadFile(*input_file)
    //check(err)
    
    /***** BINARY SEARCH TREE *****/
    tree := &Node{Value: 100}
    tree.Insert(53)
    tree.Insert(203)
    tree.Insert(19)
    tree.Insert(76)
    tree.Insert(150)
    tree.Insert(310)
    tree.Insert(7)
    tree.Insert(24)
    tree.Insert(88)
    tree.Insert(276)
    
    fmt.Println(tree.Search(400))
    var test[] int
    var count int = 0
    test = tree.in_order_traversal(test, &count)
    var hash int = tree.computeHash()
    fmt.Printf("%v\n", test)
    fmt.Printf("hash value: %d\n", hash)
}