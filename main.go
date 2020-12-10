package main

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
	httptransport "github.com/go-kit/kit/transport/http"
	consulapi "github.com/hashicorp/consul/api"
	"gokit-client/services"
	"io"
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

//客户端通过consul调用服务
func main() {
	{
		//第一步，创建client
		config := consulapi.DefaultConfig()
		config.Address = "192.168.1.104:8500" //虚拟机consul服务地址
		api_client, _ := consulapi.NewClient(config)
		client := consul.NewClient(api_client)

		var logger log.Logger
		{
			logger = log.NewLogfmtLogger(os.Stdout)
		}
		{
			tags := []string{"primary"}
			//可实时查询服务实例的状态信息
			instancer := consul.NewInstancer(client,logger,"gokitservice",tags,true)
			{
				factory := func(service_url string) (endpoint.Endpoint,io.Closer,error) {
					tart,_ := url.Parse("http://"+service_url) //192.168.1.103:8080真实服务ip地址
					return httptransport.NewClient("GET",tart,services.GetUserInfo_Request,services.GetUserInfo_Response).Endpoint(),nil,nil
				}
				endpointer := sd.NewEndpointer(instancer,factory,logger)
				endpoints,_ := endpointer.Endpoints()
				fmt.Println("服务有",len(endpoints),"条")

				//go-kit自带负载均衡
				//mylb := lb.NewRoundRobin(endpointer)//轮询
				mylb := lb.NewRandom(endpointer,time.Now().UnixNano())//随机
				for{
					//getUserInfo := endpoints[0]//写死第一条
					getUserInfo, _ := mylb.Endpoint()//轮询客户端获取服务
					ctx := context.Background() //第三步，创建一个context上下文对象
					//第四步，执行
					res,err := getUserInfo(ctx,services.UserRequest{Uid:101})
					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
					//第五步，断言，得到相应值
					userinfo := res.(services.UserResponse)
					fmt.Println(userinfo.Result)
					time.Sleep(time.Second * 3)
				}

			}
		}
	}

}