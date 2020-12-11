package main

import (
	"context"
	"fmt"
	"github.com/afex/hystrix-go/hystrix"
	httptransport "github.com/go-kit/kit/transport/http"
	"gokit-client/services"
	"gokit-client/utils"
	"golang.org/x/time/rate"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)


//客户端直接调用服务
func main2() {
	tgt,_ := url.Parse("http://localhost:8080")
	client := httptransport.NewClient("GET",tgt,services.GetUserInfo_Request,services.GetUserInfo_Response)
	getUserInfo := client.Endpoint()
	ctx := context.Background()
	res,err := getUserInfo(ctx,services.UserRequest{Uid:102})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	userinfo := res.(services.UserResponse)
	fmt.Println(userinfo.Result)
}

//使用rate包达到api限流
var r = rate.NewLimiter(1,5)
func MyLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter,request *http.Request) {
		if !r.Allow() {
			http.Error(writer,"too many requests",http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(writer,request)
	})
}
func main4() {
	mux := http.NewServeMux()
	mux.HandleFunc("/",func(writer http.ResponseWriter,request *http.Request) {
		writer.Write([]byte("OK!!!"))
	})
	http.ListenAndServe(":8089",MyLimit(mux))
}


func main() {
	configA := hystrix.CommandConfig{
		Timeout:2000,
		MaxConcurrentRequests:5,
		RequestVolumeThreshold:3,
		ErrorPercentThreshold:20,
		SleepWindow:int(time.Second*100),
	}
	hystrix.ConfigureCommand("getuser",configA)
	err := hystrix.Do("getuser",func() error{
		res,err := utils.GetUser()
		fmt.Println(res)
		return err
	},func (e error) error{
		fmt.Println("降级用户")
		return e
	})
	if err != nil {
		log.Fatal(err)

	}

}