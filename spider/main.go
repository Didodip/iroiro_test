package main

import (
	nh "awesomeProject/spider/nhentai.net"
	"fmt"
)

func main()  {
	//g :=nh.GetGalleries(1250)
	//fmt.Println(nh.ReflectReadBook(g[0],"title"),nh.ReflectReadBook(g[0],"url"))
	fmt.Println("哈哈---")
	//fmt.Println(nh.GetGallery("1234-333",0))
	//fmt.Println(nh.GetImgUrl(564,1,".jpg"))
	//fmt.Println(nh.DownloadOneImg(nh.GetImgUrl(564,5,".jpg"),"D:/1.jpg"))
	//nh.StartSpider(1250,1240)
	nh.StartSpider(1,500)


}



