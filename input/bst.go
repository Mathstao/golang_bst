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
    "time"
    "errors"
)



/*****************************************************************************************/

type Job struct {
    tree        *Node
    bst_id      int
    Name        string
}


type JobChannel chan Job

type JobQueue chan chan Job

type Worker struct {
    ID             int
    JobChan        JobChannel
    Queue          JobQueue   // shared between all workers and dispatchers.
    Quit           chan struct{}
    ReturnQueue    chan HashBstID
    SyncWaitGroup  *sync.WaitGroup
    WorkersUpdate  bool
    Mutex          *sync.Mutex
    Bst_hashmap    *map[int][]int
}

type disp struct {
    Workers            []*Worker  // this is the list of workers that dispatcher tracks
    WorkChan           JobChannel // client submits a job to this channel
    Queue              JobQueue   // this is the shared JobPool between the workers
    ReturnQueue        chan HashBstID
    SyncWaitGroup      *sync.WaitGroup
    WorkersUpdate      bool
    Mutex              *sync.Mutex
    Bst_hashmap        *map[int][]int
}



func (wr *Worker) Start() {
    go func() {
        for {
            wr.Queue <- wr.JobChan //worker puts its jobChan on queue (channel of channels). wr.Queue populated by dispatcher.process()
            select {
            case job := <-wr.JobChan:
                var hash int = job.tree.computeHash() //todo: extract below to outside, iterate over list of trees with hash_workers
                if wr.WorkersUpdate {
                    wr.Mutex.Lock()
                    (*wr.Bst_hashmap)[hash] = append((*wr.Bst_hashmap)[hash], job.bst_id)
                    wr.Mutex.Unlock()
                } else {
                    pair := HashBstID{hash: hash, bst_id: job.bst_id}
                    wr.ReturnQueue <- pair
                }
                wr.SyncWaitGroup.Done()
            case <-wr.Quit:
                fmt.Println("worker id calling quit", wr.ID)
                close(wr.JobChan)
                return
            }
        }
    }() 
}




func (d *disp) Start() *disp {
    l := len(d.Workers)
    for i := 0; i < l; i++ {
        wrk := &Worker{ID: i, JobChan: make(JobChannel), Queue: d.Queue, Quit: make(chan struct{}), 
                       ReturnQueue: d.ReturnQueue, SyncWaitGroup: d.SyncWaitGroup,
                       WorkersUpdate: d.WorkersUpdate, Mutex: d.Mutex, Bst_hashmap: d.Bst_hashmap,}
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
/*****************************************************************************************/

type InputArgs struct {
    hash_workers *int
    data_workers *int
    comp_workers *int
    input_file *string
    run_mode int
}

type WorkersArgs struct {
    sequential              bool
    hashwrkrs_update_map    bool
    exit_hash               bool
    exit_map                bool
    extra_credit            bool
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

func compute_hash(tree *Node, bst_hashmap *map[int][]int, bst_id int, 
                 queue chan<- HashBstID, wg *sync.WaitGroup){
    
    defer wg.Done()
    
    var hash int = tree.computeHash() //todo: extract below to outside, iterate over list of trees with hash_workers
    hash_bst_pair := HashBstID{hash: hash, bst_id: bst_id}
    queue <- hash_bst_pair
}

func hash_worker(id int, job_channel <-chan Job, bst_hashmap *map[int][]int,
                 wg *sync.WaitGroup, mutex *sync.Mutex, queue chan<- HashBstID, workers_update bool){
    for job := range job_channel {
        var hash int = job.tree.computeHash() //todo: extract below to outside, iterate over list of trees with hash_workers
        if workers_update {
            mutex.Lock()
            (*bst_hashmap)[hash] = append((*bst_hashmap)[hash], job.bst_id)
            mutex.Unlock()
        } else {
            pair := HashBstID{hash: hash, bst_id: job.bst_id}
            queue <- pair
        }
        wg.Done()
    }   
}

func data_worker(id int, queue <-chan HashBstID, bst_hashmap *map[int][]int,
                 wg_data *sync.WaitGroup, mutex_data *sync.Mutex){
    for job := range queue {
		mutex_data.Lock()
        (*bst_hashmap)[job.hash] = append((*bst_hashmap)[job.hash], job.bst_id)
        mutex_data.Unlock()
		wg_data.Done()
	}
}

func run(bst_list *[]*Node, bst_hashmap *map[int][]int,
         tree_equal *map[int][]int, args InputArgs, 
         worker_args WorkersArgs, test_hash_time bool) {
    
    var wg sync.WaitGroup
    var mutex sync.Mutex
    var wg_data sync.WaitGroup
    var mutex_data sync.Mutex
    
    tree_visited:= make(map[int]Visited)
    
    /***STEP 1***/
    build_trees(args.input_file, bst_list) //always sequential to preserve tree ordering
    fmt.Println("number of trees: ", len(*bst_list))
    job_channel := make(chan Job, len(*bst_list)) //HAS TO GO AFTER BUILD TREES!!!!
    queue := make(chan HashBstID, len(*bst_list)) //HAS TO GO AFTER BUILD TREES!!!!
    
    /***STEP 2--hash trees***/
    if (worker_args.sequential){
        if (worker_args.exit_map && worker_args.exit_hash){
            for _, tree := range(*bst_list){
                tree.computeHash()
            }
            return
        } else {
            for bst_id, tree := range(*bst_list){
                var hash int = tree.computeHash()
                (*bst_hashmap)[hash] = append((*bst_hashmap)[hash], bst_id)
            }
            compare_trees(bst_list, bst_hashmap, tree_equal)
            return
        }
    }
    if test_hash_time {
        start := time.Now()
        for bst_id, tree := range *bst_list{
            wg.Add(1)
            go compute_hash(tree, bst_hashmap, bst_id, queue, &wg)
        }
        wg.Wait()
        elapsed := time.Since(start)
        fmt.Printf("hashing took %s\n", elapsed)
    } else {
        wg.Add(len(*bst_list))
        fmt.Println("hash workers update map ", worker_args.hashwrkrs_update_map)
       
        /*
        dd := &disp{Workers:  make([]*Worker, *args.hash_workers),
                    WorkChan: make(JobChannel, len(*bst_list)),
                    Queue:    make(JobQueue),
                    ReturnQueue: queue,
                    SyncWaitGroup: &wg,
                    WorkersUpdate: worker_args.hashwrkrs_update_map,
                    Mutex: &mutex,
                    Bst_hashmap: bst_hashmap}       
        dd.Start()
        start := time.Now()
        for i, tree := range *bst_list {
            dd.Submit(Job{ tree:tree, Name:fmt.Sprintf("JobID::%d", i), bst_id: i} )
        }
        */
        
        for id:=0; id < *args.hash_workers; id++ {
            go hash_worker(id, job_channel, bst_hashmap,
                           &wg, &mutex,  queue,worker_args.hashwrkrs_update_map)
        }
        start := time.Now()
        for i, tree := range *bst_list{
            job_channel <- Job{ tree:tree, Name:fmt.Sprintf("JobID::%d", i), bst_id: i}
        }
        close(job_channel)
        
        wg.Wait()
        
        elapsed_hash := time.Since(start)
        fmt.Printf("hashing took %s\n", elapsed_hash)
        close(queue)
        if (!worker_args.hashwrkrs_update_map && *args.data_workers!=1){
            wg_data.Add(len(*bst_list))
            start_group := time.Now()
            for id:=0; id < *args.data_workers; id++ {
                go data_worker(id, queue, bst_hashmap,
                               &wg_data, &mutex_data)
            }
            wg_data.Wait()
            elapsed_group := time.Since(start_group)
            fmt.Printf("grouping took took %s\n", elapsed_group)
        }
    }
    
    //close(queue)
    
    if worker_args.exit_hash{
        return
    }
    
    /***STEP 3--find duplicates(data_workers)***/
    if *args.data_workers==1 {
        for pair := range queue {
            (*bst_hashmap)[pair.hash] = append((*bst_hashmap)[pair.hash], pair.bst_id)
        }
    }
    
    if worker_args.exit_map {
        return
    }
    /***STEP 4--compare equal trees***/
    if (*args.comp_workers==1){
        compare_trees(bst_list, bst_hashmap, tree_equal) 
    } else {
        compare_trees_parallel(bst_list, bst_hashmap, tree_equal, *args.comp_workers, &tree_visited)
    }
    
}


func build_worker_args(args InputArgs) WorkersArgs {
    exit_hash, exit_map, hashwrkrs_update_map, sequential, extra_credit := false, false, false, false, false

    if *args.data_workers==0{ exit_hash = true }
    if *args.comp_workers==0{ exit_map = true }
    if *args.hash_workers==1{ sequential = true }
    if !sequential && (*args.hash_workers==*args.data_workers){ hashwrkrs_update_map = true }
    if !sequential && *args.data_workers > 1 && !hashwrkrs_update_map{ extra_credit = true }
    
    worker_args := WorkersArgs{exit_map: exit_map, exit_hash: exit_hash,
                               hashwrkrs_update_map: hashwrkrs_update_map,
                               sequential: sequential, extra_credit: extra_credit}
    return worker_args
}

func parse_args() InputArgs {
    hash_workers := flag.Int("hash-workers", 1, "an int")
    data_workers := flag.Int("data-workers", 0, "an int")
    comp_workers := flag.Int("comp-workers", 0, "an int")
    input_file   := flag.String("input", "", "a string")
    
    flag.Parse()
    
    args := InputArgs{hash_workers: hash_workers,
                      data_workers: data_workers,
                      comp_workers: comp_workers,
                      input_file: input_file}
    return args
}

func main() {
    
    var test_hash_time bool = false //TODO: for testing purposes: test spawning one go routine per tree vs worker pool 
    
    var bst_list []*Node; //list of trees
    bst_hashmap := make(map[int][]int) //hash id: list of bst_ids
    tree_equal := make(map[int][]int) //groups of equal trees
    
    var args InputArgs = parse_args()
    /*******   PRINTING INPUTS ******/
    fmt.Println("hash workers:", *args.hash_workers, "data workers:", *args.data_workers, "comp workers:", 
                *args.comp_workers, "input file:", args.input_file)

    worker_args := build_worker_args(args)
    run(&bst_list, &bst_hashmap, &tree_equal, args, worker_args, test_hash_time)
    
    //fmt.Println(bst_hashmap)
    //fmt.Println(bst_list)
    //fmt.Println(tree_equal)
    for key,value := range tree_equal {
        if len(value) <= 1 {
            delete(tree_equal, key)
        }
    }
    fmt.Println(tree_equal)
    
}

type Queue struct {
    mu *sync.Mutex
    capacity int
    q        []CompareTreesPair
}

// FifoQueue 
type FifoQueue interface {
    Insert()
    Remove()
}

// Insert inserts the item into the queue
func (q *Queue) Insert(item CompareTreesPair) bool {
    q.mu.Lock()
    defer q.mu.Unlock()
    if len(q.q) < int(q.capacity) {
        q.q = append(q.q, item)
        return true
    }
    return false
}

// Remove removes the oldest element from the queue
func (q *Queue) Remove() (CompareTreesPair, error) {
    var item CompareTreesPair
    q.mu.Lock()
    defer q.mu.Unlock()
    if len(q.q) > 0 {
        item := q.q[0]
        q.q = q.q[1:]
        return item, nil
    }
    return item, errors.New("Empty")
}

// CreateQueue creates an empty queue with desired capacity
func CreateQueue(capacity int, mutex *sync.Mutex) *Queue {
    return &Queue{
        capacity: capacity,
        q:        make([]CompareTreesPair, 0, capacity),
        mu:       mutex,
    }
}

type CompareTreesPair struct {
    group_id int
    tree_a_id int
    tree_b_id int
}

type Visited struct {
    isVisited bool
    group_id int
}

func update_global_tree_equal(a_id int, b_id int, 
                              mutex *sync.Mutex, tree_visited *map[int]Visited){
    
    var a_visited Visited;
    var b_visited Visited;
    var a_ok bool;
    var b_ok bool;
    var id int;
    mutex.Lock()
    defer mutex.Unlock()
    
    a_visited, a_ok = (*tree_visited)[a_id]
    b_visited, b_ok = (*tree_visited)[b_id]
    if(!a_ok && !b_ok){
        
        if (a_id < b_id){
            id = a_id
        } else {
            id = b_id
        }
        visited := Visited{isVisited: true, group_id: id}
        (*tree_visited)[a_id] = visited
        (*tree_visited)[b_id] = visited
        return
    }
    if (a_ok && b_ok){
        if (a_id < b_id){
            id = a_visited.group_id
        } else {
            id = b_visited.group_id
        }
        visited := Visited{isVisited: true, group_id: id}
        (*tree_visited)[b_id] = visited
        (*tree_visited)[a_id] = visited
        return
    }
    if (a_ok && !b_ok){
        (*tree_visited)[b_id] = a_visited
        return
    }
    if (b_ok && !a_ok){
        (*tree_visited)[a_id] = b_visited
        return
    }
    return
}

func comp_worker_run(q *Queue, bst_list *[]*Node,
                     mutex_visited *sync.Mutex, tree_visited *map[int]Visited, 
                     continue_working <-chan int, wg *sync.WaitGroup){
    for{
        var my_error error;
        var pair CompareTreesPair;
        for{
            _, ok := <- continue_working;
            if (!ok){
                
                return
            }
            pair, my_error = q.Remove()
            if (my_error==nil){
                break
            }
            time.Sleep(100 * time.Millisecond)
        }
        var node *Node = (*bst_list)[ pair.tree_a_id ]
        var next_node *Node = (*bst_list)[ pair.tree_b_id ]
        var equal bool = equalTrees(node, next_node)
        if equal{
            update_global_tree_equal(pair.tree_a_id, pair.tree_b_id, 
                                     mutex_visited, tree_visited)
        }
        wg.Done()
    }
}


func compare_trees_parallel(bst_list *[]*Node, bst_hashmap *map[int][]int,
                            tree_equal *map[int][]int, comp_workers int, 
                            tree_visited *map[int]Visited){
    var mutex sync.Mutex
    var mutex_visited sync.Mutex
    var wg sync.WaitGroup
    
    continue_working := make(chan int)
    trees_work := CreateQueue(comp_workers, &mutex) //pointer!!
    
    //spin comp_workers up
    for i:=0; i < comp_workers; i++ {
        go comp_worker_run(trees_work, bst_list, &mutex_visited, tree_visited, continue_working, &wg)
    }
    
    for group_id, bstids := range *bst_hashmap { //flatten out elements in hashmap
        if (len(bstids) > 1){
            for i:=0; i < len(bstids); i++ {
                for j:=i+1; j < (len(bstids)); j++ {
                    pair := CompareTreesPair{group_id: group_id, tree_a_id: bstids[i], tree_b_id: bstids[j],}
                    wg.Add(1)
                    for (!trees_work.Insert(pair)){
                        //sleep the main thread until queue is freed
                        time.Sleep(100 * time.Millisecond)
                    }
                    continue_working <- 1
                }
            }
        }
    }
    wg.Wait()
    close(continue_working)
    //fmt.Println(tree_visited)
    final_map := make(map[int]int)
    
    var count int = 0;
    var exists bool;
    var id int
    for key, value := range *tree_visited {
        id, exists = final_map[value.group_id]
        if (!exists){
            final_map[value.group_id] = count
            id = count
            count++
        }
        (*tree_equal)[id] = append((*tree_equal)[id], key)
    }
}

func compare_trees(bst_list *[]*Node, bst_hashmap *map[int][]int,
                   tree_equal *map[int][]int){
    
    var tree_group int = -1
    
    for _, bstids := range *bst_hashmap {
        this_group_visited := make(map[int]bool)
        if (len(bstids) > 1) {
            //fmt.Printf("Compare values in key: %d\n", hash)
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
