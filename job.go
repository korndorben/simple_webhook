package main

import (
	"sync"
	"time"
	"strconv"
	"fmt"
	"os"
	"io/ioutil"
	"encoding/json"
)

var configFile = "./jobs.queue"

//任务队列结构
type JobQueue struct {
	lock  *sync.Mutex
	Queue map[string]*Job
}

//持久化到磁盘
func (q *JobQueue) Save() (bool, error) {
	data, err := json.Marshal(q.Queue)
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	q.lock.Lock()
	if err = ioutil.WriteFile(configFile, data, 777); err != nil {
		fmt.Println(err)
		return false, err
	}
	q.lock.Unlock()
	return true, nil
}

//从磁盘加载数据
func (q *JobQueue) Initialize() {
	if _, err := os.Stat(configFile); err != nil {
		fmt.Println(err)
		return
	}

	variables, err := ioutil.ReadFile(configFile)
	if err != nil || len(variables) <= 0 {
		fmt.Println(err)
		return
	}
	var queue map[string]*Job
	if err = json.Unmarshal(variables, &queue); err != nil {
		fmt.Println(err)
		return
	}
	if nil != queue {
		q.Queue = queue
	}
}

func (q JobQueue) Del(id string) {
	delete(q.Queue, id)
}
func (q JobQueue) Add(j *Job) {
	q.Queue[j.Id] = j
}

//任务结构
type Job struct {
	lock     sync.Mutex
	Id       string    `json:"id,omitempty"`
	Time     time.Time `json:"time,omitempty"`
	URL      string    `json:"url,omitempty"`
	MaxTries int       `json:"max_tries,omitempty"`
	fails    int       `json:"fails,omitempty"`
}

//添加任务
func NewJob(timestring, url, maxtries string) (*Job, error) {
	j := &Job{}
	j.URL = url
	j.Id = Md5(fmt.Sprintf("%s%s%s", timestring, url, maxtries))
	j.MaxTries = 0 //默认不重试
	if max, err := strconv.Atoi(maxtries); err != nil {
		//使用有效的最大重试次数覆盖默认值
		j.MaxTries = max
	}

	//转换时间
	t, err := time.ParseInLocation("2006-01-02 15:04:05", timestring, loc)
	if err != nil {
		fmt.Print(err)
		return nil, err
	}
	j.Time = t
	return j, nil
}
