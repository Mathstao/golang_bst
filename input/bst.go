package main

import (
    "fmt"
    "flag"
    //"io/ioutil"
    "bufio"
    "os"
    "log"
    "strconv"
    "text/scanner"
    "strings"
)

type InputArgs struct {
    hash_workers *int
    data_workers *int
    comp_workers *int
    input_file *string
}

type Node struct {
    Value int
    Left *Node
    Right *Node
}

func checkfile(e error) {
    if e != nil {
        panic(e)
    }
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
    
func test_tree() *Node {
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
    return tree
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



func parse_args() *InputArgs {
    hash_workers := flag.Int("hash-workers", 1, "an int")
    data_workers := flag.Int("data-workers", 1, "an int")
    comp_workers := flag.Int("comp-workers", 1, "an int")
    input_file   := flag.String("input", "", "a string")
    
    flag.Parse()
    
    args := InputArgs{hash_workers: hash_workers,
                      data_workers: data_workers,
                      comp_workers: comp_workers,
                      input_file: input_file}
    return &args
}

func main() {
    
    bst_hashmap := make(map[int][]int)
    
    fmt.Println("Running go BST sequential")
    
    /*******   PARSING ARGUMENTS   *******/ 
    var args *InputArgs = parse_args()
    
    fmt.Println("hash workers:", *args.hash_workers)
    fmt.Println("data workers:", *args.data_workers)
    fmt.Println("comp workers:", *args.comp_workers)
    fmt.Println("input file:", *args.input_file)
    
    file, err := os.Open(*args.input_file)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    
    file_scanner := bufio.NewScanner(file)
    // optionally, resize scanner's capacity for lines over 64K ... needed?
    
    var bst_id int = 0;
    //bst_chan := make(chan *Node)
    for file_scanner.Scan() {
        fmt.Println("new tree parsing:")
        var s scanner.Scanner
        s.Init(strings.NewReader(file_scanner.Text()))
        var newBST bool = true
        var tree *Node;
        var converted int;
        for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
            converted, _ = strconv.Atoi(s.TokenText())
            if newBST{
                tree = &Node{Value: converted}
                newBST = false
            } else {
                tree.Insert(converted)
            }
        }
        //bst_chan <- tree
        var test[] int
        var count int = 0
        test = tree.in_order_traversal(test, &count)
        var hash int = tree.computeHash()
        fmt.Printf("hash of %d: %d\n", bst_id, hash)
        bst_hashmap[hash] = append(bst_hashmap[hash], bst_id)
        fmt.Println("end of tree parsing")
        bst_id++
    }
    fmt.Println(bst_hashmap)
    
    
    //code diagram
    /*
    1) read file--> if not end of line, insert into binary tree, compute hash, store hashMapBST: list.append(bst id)
    2) for each key in hashMapBST: compare equality, separate out into linkage somehow
    
    
    */
    
    
}