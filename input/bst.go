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

func identicalTrees(a *Node, b *Node) bool {
    //TODO: test for empty trees!
    /*1. both empty */
    if (a == nil && b == nil) {
        return true
    }
    
    /* 2. both non-empty -> compare them */
    if (a.Value==b.Value) {
        if (identicalTrees(a.Left, b.Left) && identicalTrees(a.Right, b.Right)){
            return true
        }
    }
    /* 3. one empty, one not -> false */
    return false
}

func equalTrees(a *Node, b *Node) bool {
    count, count2 := 0,0
    var first_result []int
    var second_result []int
    first_result = a.in_order_traversal(first_result, &count)
    second_result = b.in_order_traversal(second_result, &count2)
    
    for i:=0; i < len(first_result); i++ {
        if (first_result[i] != second_result[i]){
            return false
        }
    }
    return true
}


func parse_args() InputArgs {
    hash_workers := flag.Int("hash-workers", 1, "an int")
    data_workers := flag.Int("data-workers", 1, "an int")
    comp_workers := flag.Int("comp-workers", 1, "an int")
    input_file   := flag.String("input", "", "a string")
    
    flag.Parse()
    
    args := InputArgs{hash_workers: hash_workers,
                      data_workers: data_workers,
                      comp_workers: comp_workers,
                      input_file: input_file}
    return args
}


func run_sequential(input_file *string, bst_list *[]*Node, bst_hashmap *map[int][]int,
                   tree_equal *map[int][]int){
    
    file, err := os.Open(*input_file)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    
    file_scanner := bufio.NewScanner(file)
    // optionally, resize scanner's capacity for lines over 64K ... needed?
    var bst_id int = 0
    for file_scanner.Scan() {
        var tree *Node;
        var s scanner.Scanner
        s.Init(strings.NewReader(file_scanner.Text()))
        var newBST bool = true
        for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
            var converted int;
            converted, _ = strconv.Atoi(s.TokenText())
            if newBST{
                tree = &Node{Value: converted}
                newBST = false
            } else {
                tree.Insert(converted)
            }
        }
        *bst_list = append(*bst_list, tree)
        var hash int = tree.computeHash()
        (*bst_hashmap)[hash] = append((*bst_hashmap)[hash], bst_id)
        bst_id++
    }
    //iterate over hashmaps, if key > 1, compare trees inside
    
    var tree_group int = -1
    
    for hash, bstids := range *bst_hashmap {
        this_group_visited := make(map[int]bool)
        if (len(bstids) > 1) {
            fmt.Printf("Compare values in key: %d\n", hash)
            for i:=0; i < len(bstids); i++ {
                if (!this_group_visited[i]){
                    
                    //node hasn't been visited yet, create new group in tree
                    tree_group++
                    (*tree_equal)[tree_group] = append((*tree_equal)[tree_group], bstids[i])
                    
                    var node *Node = (*bst_list)[ bstids[i] ]
                    
                    this_group_visited[i] = true
                    for j:=i+1; j < len(bstids); j++ {
                        if (!this_group_visited[j]){
                            //next node hasn't been visited, compare with node
                            var next_node *Node = (*bst_list)[ bstids[j] ]
                            var equal bool = equalTrees(node, next_node)
                            if equal{
                                (*tree_equal)[tree_group] = append((*tree_equal)[tree_group], bstids[j])
                                this_group_visited[j] = true //grouped nextnode, remove it from iterations
                            }
                        }
                    }
                }
            }
        }// else no need to print groups with only one tree
    }
}

func main() {
    
    bst_hashmap := make(map[int][]int)
    var bst_list []*Node;
    tree_equal := make(map[int][]int)
    
    
    fmt.Println("Running go BST sequential")
    
    /*******   PARSING ARGUMENTS   *******/ 
    var args InputArgs = parse_args()
    
    fmt.Println("hash workers:", args.hash_workers)
    fmt.Println("data workers:", args.data_workers)
    fmt.Println("comp workers:", args.comp_workers)
    fmt.Println("input file:", args.input_file)
    
    run_sequential(args.input_file, &bst_list, &bst_hashmap, &tree_equal)
    fmt.Println(bst_hashmap)
    fmt.Println(bst_list)
    fmt.Println(tree_equal)
    
    //code diagram
    /*
    1) read file--> if not end of line, insert into binary tree, compute hash, store hashMapBST: list.append(bst id)
    2) for each key in hashMapBST: compare equality, separate out into linkage somehow
    
    
    */
    
    
}