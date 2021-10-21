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
    "sync"
)

type Job struct {
    tree        *Node
    bst_id      int
    Name      string
}


type JobChannel chan Job

type JobQueue chan chan Job

type Worker struct {
    ID      int
    JobChan JobChannel
    Queue   JobQueue   // shared between all workers and dispatchers.
    Quit    chan struct{}
    ReturnQueue chan HashBstID
    SyncWaitGroup *sync.WaitGroup
}

func (wr *Worker) Start() {
    go func() {
        for {
            wr.Queue <- wr.JobChan //worker puts its jobChan on queue (channel of channels). wr.Queue populated by dispatcher.process()
            select {
            case job := <-wr.JobChan:
                //fmt.Println("worker id ", wr.ID, job.bst_id)
                //fmt.Println("updated return queue ", wr.ID, job.bst_id)
                //fmt.Println("computed hash ", wr.ID, job.bst_id)
                
                var hash int = job.tree.computeHash() //todo: extract below to outside, iterate over list of trees with hash_workers
                hash_bst_pair := HashBstID{hash: hash, bst_id: job.bst_id}
                wr.ReturnQueue <- hash_bst_pair
                
                wr.SyncWaitGroup.Done()
            case <-wr.Quit:
                fmt.Println("worker id calling quit", wr.ID)
                close(wr.JobChan)
                return
            }
        }
    }() 
}

func (wr *Worker) StartAndUpdateMap() {
    go func() {
        for {
            wr.Queue <- wr.JobChan //worker puts its jobChan on queue (channel of channels). wr.Queue populated by dispatcher.process()
            select {
            case job := <-wr.JobChan:
                var hash int = job.tree.computeHash() //todo: extract below to outside, iterate over list of trees with hash_workers
                hash_bst_pair := HashBstID{hash: hash, bst_id: job.bst_id}
                wr.ReturnQueue <- hash_bst_pair
                wr.SyncWaitGroup.Done()
            case <-wr.Quit:
                fmt.Println("worker id calling quit", wr.ID)
                close(wr.JobChan)
                return
            }
        }
    }() 
}


type disp struct {
    Workers            []*Worker  // this is the list of workers that dispatcher tracks
    WorkChan           JobChannel // client submits a job to this channel
    Queue              JobQueue   // this is the shared JobPool between the workers
    ReturnQueue        chan HashBstID
    SyncWaitGroup      *sync.WaitGroup
    UseWorkersForData  bool
}



func (d *disp) Start() *disp {
    //defer d.SyncWaitGroup.Done()
    l := len(d.Workers)
    for i := 1; i <= l; i++ {
        wrk := &Worker{ID: i, JobChan: make(JobChannel), Queue: d.Queue, Quit: make(chan struct{}), 
                       ReturnQueue: d.ReturnQueue, SyncWaitGroup: d.SyncWaitGroup}
        wrk.Start()
        d.Workers = append(d.Workers, wrk)
    }
    go d.process()
    return d
}

func (d *disp) process() {
    for {
        select {
        case job := <-d.WorkChan: // listen to a submitted job on WorkChannel
            jobChan := <-d.Queue  // pull out an available jobchannel from queue
            jobChan <- job        // submit the job on the available jobchannel
        }
    }
}

func (d *disp) Submit(job Job) {
    d.WorkChan <- job
}


/*****************************************************************************************/

type InputArgs struct {
    hash_workers *int
    data_workers *int
    comp_workers *int
    input_file *string
    run_mode int
}

type HashBstID struct {
    hash int
    bst_id int
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
        if n.Right==nil {
            n.Right = &Node{Value : value}
        } else {
            n.Right.Insert(value)
        }
    } else if (value < n.Value) {
        if n.Left==nil {
            n.Left = &Node{Value : value}
        } else {
            n.Left.Insert(value)
        }
    } else {
        fmt.Println("already exists")
    }
}
    
func build_trees(input_file *string, bst_list *[]*Node){
    file, err := os.Open(*input_file)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    
    file_scanner := bufio.NewScanner(file)
    // optionally, resize scanner's capacity for lines over 64K ... needed?
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
    }
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





func compare_trees(bst_list *[]*Node, bst_hashmap *map[int][]int,
                   tree_equal *map[int][]int){
    
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

func compute_hash(tree *Node, bst_hashmap *map[int][]int, bst_id int, 
                 queue chan<- HashBstID, wg *sync.WaitGroup){
    
    defer wg.Done()
    
    var hash int = tree.computeHash() //todo: extract below to outside, iterate over list of trees with hash_workers
    hash_bst_pair := HashBstID{hash: hash, bst_id: bst_id}
    queue <- hash_bst_pair
}

func run_sequential(bst_list *[]*Node, bst_hashmap *map[int][]int,
                   tree_equal *map[int][]int, args InputArgs){
    
    build_trees(args.input_file, bst_list)
    
    queue := make(chan HashBstID, len(*bst_list))
    
    /*
    var wg sync.WaitGroup
    var bst_id int = 0
    for _, tree := range *bst_list{
        wg.Add(1)
        go compute_hash(tree, bst_hashmap, bst_id, queue, &wg)
        bst_id++
    }
    wg.Wait()
    close(queue)
    for pair := range queue{
        (*bst_hashmap)[pair.hash] = append((*bst_hashmap)[pair.hash], pair.bst_id)
    }*/
    
    var wg sync.WaitGroup
    wg.Add(len(*bst_list))
    dd := &disp{ Workers:  make([]*Worker, *args.hash_workers),
                 WorkChan: make(JobChannel),
                 Queue:    make(JobQueue),
                 ReturnQueue: queue,
                 SyncWaitGroup: &wg,
               }
    dd.Start()
    for i, tree := range *bst_list {
        dd.Submit(Job{ tree:tree, Name:fmt.Sprintf("JobID::%d", i), bst_id: i} )
    }
    //close(dd.WorkChan)
    wg.Wait()
    close(queue)
    for pair := range queue{
        (*bst_hashmap)[pair.hash] = append((*bst_hashmap)[pair.hash], pair.bst_id)
    }
    compare_trees(bst_list, bst_hashmap, tree_equal)
    //iterate over hashmaps, if key > 1, compare trees inside
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
    
    run_sequential(&bst_list, &bst_hashmap, &tree_equal, args)
    fmt.Println(bst_hashmap)
    fmt.Println(bst_list)
    fmt.Println(tree_equal)
    
    //code diagram
    /*
    1) read file--> if not end of line, insert into binary tree, compute hash, store hashMapBST: list.append(bst id)
    2) for each key in hashMapBST: compare equality, separate out into linkage somehow
    
    
    */
}