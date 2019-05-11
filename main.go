package main

import (
	"net/http"
	"flag"
	"github.com/gorilla/mux"
	"log"
	"time"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"sync"
)

var (
	//任务队列
	q = &JobQueue{
		lock:  new(sync.Mutex),
		Queue: make(map[string]*Job),
	}
	QUEUECOUNTS = 0 //队列变化指示器
	//东八区
	loc, _ = time.LoadLocation("Asia/Chongqing")
)

//启动队列
func InitJobQueue() {
	q.Initialize()

	//endless loop
	for {
		for ft, j := range q.Queue {
			//超过最大失败次数时,从队列中移除
			if j.fails > j.MaxTries {
				//将键删除
				q.Del(j.Id)
			}

			//取出所有超时任务
			if time.Now().In(loc).After(j.Time) {
				go func() {
					//发送get请求，触发业务接口
					response, err := http.Get(j.URL)
					if err != nil {
						j.fails += 1

						//约定为10秒后重试,最多重试MaxTries次
						j.Time = j.Time.Add(time.Second * 10)
						return
					}

					//对方返回成功时,从队列中移除
					if response.StatusCode == 200 {
						//将键删除
						delete(q.Queue, ft)
					}
				}()
			}
		}
		if len(q.Queue) > 0 {
			fmt.Println(fmt.Sprintf("%v\t%d", time.Now(), len(q.Queue)))
		}
		//每10稍保持一次
		if QUEUECOUNTS != len(q.Queue) && time.Now().In(loc).Second()%10 == 0 {
			fmt.Println(fmt.Sprintf("队列已自动保存:%d", QUEUECOUNTS))
			if ok, _ := q.Save(); ok {
				QUEUECOUNTS = len(q.Queue)
			}
		}

		time.Sleep(time.Second * 1)
	}
	fmt.Println("end server...")
}

func main() {
	go InitJobQueue()

	var addr = flag.String("addr", ":8080", "http service address")
	flag.Parse()

	// 构造路由表
	r := mux.NewRouter()

	//查看列表
	r.HandleFunc("/list.html", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		data, err := json.Marshal(q.Queue)
		if err != nil {
			fmt.Println(err)
		}

		// 将结果透传给前端
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(data); err != nil {
			fmt.Println("RegisterHandler:", err.Error())
		}
	})

	//添加任务
	r.HandleFunc("/add.html", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		// 跨域测试的应答
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Content-Length, Authorization, Accept,X-Requested-With")
		w.Header().Set("Access-Control-Allow-Methods", "PUT,POST,GET,DELETE,OPTIONS")
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusOK)
			return
		}

		variables, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err)
		}

		var f map[string]string
		if err := json.Unmarshal(variables, &f); err != nil {
			fmt.Println("add.html:", err.Error())
		}
		if len(f["time"]) <= 0 || len(f["url"]) <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		jobInstance, err := NewJob(f["time"], f["url"], f["maxtries"])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		q.Add(jobInstance)

		// 将结果透传给前端
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(variables)))
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(variables); err != nil {
			fmt.Println("RegisterHandler:", err.Error())
		}
	})

	//删除任务
	r.HandleFunc("/del.html", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		// 跨域测试的应答
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Content-Length, Authorization, Accept,X-Requested-With")
		w.Header().Set("Access-Control-Allow-Methods", "PUT,POST,GET,DELETE,OPTIONS")
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusOK)
			return
		}

		variables, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err)
		}

		var f map[string]string
		if err := json.Unmarshal(variables, &f); err != nil {
			fmt.Println("RegisterHandlerParams:", err.Error())
		}

		if len(f["id"]) <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		q.Del(f["id"])

		// 将结果透传给前端
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(variables)))
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(variables); err != nil {
			fmt.Println("RegisterHandler:", err.Error())
		}
	})
	http.Handle("/", r)
	fmt.Println("server starts")
	if err := http.ListenAndServe(*addr, r); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
