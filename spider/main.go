package main

import (
	"awesomeProject/spider/nh"
	"fmt"
)
var a=[...]int{1,2,2,2,2,2}
func main()  {
	fmt.Println("哈哈---")
	//fmt.Println(nh.GetGalley("/g/315621/",0))

	nh.StartSpider("/search/?q=kantai+chinese+full-color&page=",16)//tag:url上要加&page=或者？page=;stop:page数+1
	//nh.DownloadOneDoujin(316913,"H:/test/monaka udon/")
	//nh.SpiderOneBook(304296,"H:/test/monaka udon/")
	//req,err:=http.NewRequest("GET","https://nhentai.net/tag/full-color/?page=1", nil)
	//fmt.Println("https://nhentai.net/tag/full-color/?page="+strconv.Itoa(1))
	//req,err:=http.NewRequest("GET","https://www.baidu.com", nil)
	//handleError(err,"Http.get.page.req:{GetGalleries}")
	//req.Header.Set("user-agent",userAgent)
	//req.Header.Set("cookie",cookies)

}
