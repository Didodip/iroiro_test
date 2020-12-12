package nh

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)
const (
	userAgent  ="Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.88 Safari/537.36"
	cookies = "__cfduid=df811481cff8f3879712ea5a02b18feb01576840808; _ga=GA1.2.1931960219.1576840808; _gid=GA1.2.1273994293.1577848755"
	HttpProxy  = "http://127.0.0.1:1080"
)
type book struct {
	id int
	title string
	imgUrl string//important!我在2020.12修改中误认为这个就是返回gallery id的，实际上这个是返回图片库id的，塔门不一样！
	parodies []string
	characters []string
	artists []string
	languages []string
	page string
	h2title string//原文标题
	h1title string//英文标题
}
var (
	bannedCharacter = [...]string{"?", "*", ":", "\"", "<", ">", "\\", "/", "|", "◯"}
	imgType = [...]string{".jpg", ".png", ".gif"}
	wg sync.WaitGroup
	proxy = func(_ *http.Request) (*url.URL, error) {
		return url.Parse(HttpProxy)
	}

	httpTransport = &http.Transport{
		Proxy: proxy,
	}

	httpClient = &http.Client{
		Transport: httpTransport,
	}
)
var v = atomic.Value{}
var mapBook = make(map[int]book)
var disk = "H://test/kantai/"//结尾一定加"/"

//检查路径是否存在
func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
//获取book值用，返回value
func reflectReadBook(b book, key string) string {
	return reflect.ValueOf(b).FieldByName(key).String()
}
// 检查title中是否含有特殊字符，返回特殊字符修改为 "-" 后的title
func checkSpecial(str string)string  {
	for i :=range bannedCharacter {
		if strings.ContainsAny(str, bannedCharacter[i]) {
			str = strings.Replace(str, bannedCharacter[i], "-", -1 )
			//fmt.Println(bannedCharacter[i])
		}
	}
	return str
}
//返回url中的gallery数字id
//或者返回正则形如../...a.b../..的"a"+"b",ab是数字
func getGallery(url string, num int) int {
	//fmt.Println(url)
	re, _ := regexp.Compile(`\d+`)
	all := re.FindAll([]byte(url),-1)
	if len(all)<num+1 {
		fmt.Println(num+1,all,url)
		fmt.Println("can't get gallery id(List Out Of Range):{GetGallery}")
		os.Exit(0)
		return 0
	}else {
		a,_:= strconv.Atoi(string(all[num]))
		return a
	}
}
//排错
func handleError (err error, reason string)  {
	if err != nil {
		fmt.Println(reason,err)
	}
}

//开始，爬！
//tag:url上要加&或者?
//stop:page数+1
func StartSpider(tag string,stop int)  {
	var now = 1
	for now<stop{
		DownloadOnePage(tag,now)
		now++
	}
}

//多线程爬一页面
func SpiderOnePage(tag string,page int)  {
	start := time.Now()
	//fmt.Println("before")
	gs:=getGalleries("https://nhentai.net"+tag+strconv.Itoa(page))//get的是slice
	//fmt.Println("after")
	//SpiderOnePage
	//i :=  0
	fmt.Println("now page",page,len(gs))
	for i:=range gs{
		if i==len(gs) {
			break
		}
		title := reflectReadBook(gs[i],"title")
		//bookurl := reflectReadBook(gs[i],"url")
		//imgid := getGallery(bookurl,0)
		gallery:=gs[i].id
		fmt.Println(gallery)
		path := disk+checkSpecial(title)+"/"
		if pathExists(path) {
			//i++
		}else {
			err:=os.MkdirAll(path,os.ModePerm)
			if err!=nil {
				fmt.Println("193error,mkdir err")
				//i++
			}else {
				//协程：一页上的本子25个协程爬取。
				//spiderOneBook
				wg.Add(1)
				go func() {
					defer wg.Done()
					//SpiderOneBook(imgid,path)
					DownloadOneDoujin(gallery,path)
				}()
				//i++

			}
		}
	}
	wg.Wait()
	elapsed := time.Since(start)
	fmt.Println("Took", elapsed)
}
//单线程爬一页面
func DownloadOnePage(tag string,page int)  {
	fmt.Println("DownloadOnePage:",page)
	start := time.Now()
	gs:=getGalleries("https://nhentai.net"+tag+strconv.Itoa(page))//get的是slice
	fmt.Println("now page",page,len(gs))
	for i:=range gs{
		gallery:=gs[i].id
		DownloadOneDoujin(gallery,disk)
	}
	elapsed := time.Since(start)
	fmt.Println("Took", elapsed)
}
//返回title和url列表(一个页面的所有) [2020.12 为mapBook和innerImg两个都赋值，]
func getGalleries(tp string)[]book  {
	fmt.Println(tp)
	req,err:=http.NewRequest("GET",tp , nil)
	handleError(err,"getGalleries : req failed.")
	req.Header.Set("user-agent",userAgent)
	req.Header.Set("cookie",cookies)
	resp, _ := httpClient.Do(req)
	if err!=nil {
		handleError(err,"getGalleries : get resp failed(可能是代理设置问题).")
		os.Exit(2)
	}
	defer resp.Body.Close()
	pageBytes,err:=ioutil.ReadAll(resp.Body)
	docs, err := goquery.NewDocumentFromReader(bytes.NewReader(pageBytes))
	var sliceBook []book//这是问题所在！放里面是为了每一个页面都有个新的slice
//我脑抽了吗干什么用slice？？？？（下载单页。）
//那map是用来干什么的？？？？（数据入表。）
	if docs!=nil {
		docs.Find("div.index-container.container div.gallery a").Each(func(i int, s *goquery.Selection) {
			idPath, _ :=s.Attr("href")
			//fmt.Println(idPath)
			id:= getGallery(idPath, 0)
			imgPath, exists := s.Find("img").Attr("data-src")
			//print(s.Text()
			title := s.Find("div.caption").Text()
			//print(imgPath)
			//第i个本子的<img data-src:"...">是否存在？不存在就判断src存在，还不存在就退出！
			if !exists {
				imgPath, exists = s.Find("img").Attr("src")
				if !exists{
					//print("noo")
					os.Exit(66)
				}
			}
			mapBook[id]=book{
				id:         getGallery(imgPath,0),
			}
			sliceBook=append(sliceBook,book{
				id:         id,
				title:      title,
			} )
		})
	}
	//mapBook是累积的，而sliceBook只管当前页面的内容
	fmt.Println(len(mapBook))
	return sliceBook
}
//返回图片url
func getImgUrl(gallery int, page int,imgType string)(url string){
	return "https://i.nhentai.net/galleries/"+strconv.Itoa(gallery)+"/"+strconv.Itoa(page)+imgType
}

//下载单张图片
func downloadOneImg(imgUrl, path string)bool{
	//fmt.Println(imgUrl)
	req,err:=http.NewRequest("GET",imgUrl, nil)
	resp,err:=httpClient.Do(req)
	if err!=nil {
		fmt.Println(err)
		return false
	}
	code:=resp.StatusCode
	if code!=200 {
		fmt.Println("0",imgUrl)
		return false
	}
	body,err:=ioutil.ReadAll(resp.Body)
	if err!=nil {
		fmt.Println("2")
		return false
	}
	file, err := os.Create(path)
	if err!=nil {
		fmt.Println(err)
		return false
	}
	_, err = io.Copy(file, bytes.NewReader(body))
	//fmt.Println(body)
	if err!=nil {
		fmt.Println("3")
		fmt.Println(err)
		return false
	}
	fmt.Println("got ",imgUrl)
	return true
}
//单线程爬一本
func SpiderOneBook(gallery int,path string)  {
	//这个galleries是imgURL的id
	//fmt.Println("start now")
	startb := time.Now()
	chooseType:=0
	p,et:=1,0
	for{
		imgUrl := getImgUrl(gallery,p,imgType[chooseType])
		if et>3  {
			elapsed := time.Since(startb)
			fmt.Println("completed!--",elapsed,"  ",path)
			break
		}
		status := downloadOneImg(imgUrl,path+strconv.Itoa(p)+imgType[chooseType])
		//status := downloadOneImg(imgUrl,path+strconv.Itoa(p)+imgType[chooseType])
		if !status {
			et+=1
			if chooseType<len(imgType)-1 {
				chooseType+=1
			}else {
				chooseType-=len(imgType)-1
			}
		}else {
			p++
			et=0
		}
	}
}
//多线程爬一本（思路：确定page数，开启多线程，爬完停止）
//path是路径
func DownloadOneDoujin(gallery int,path string)  {//这个是gallery
	startingTime:=time.Now()
	req,err:=http.NewRequest("GET","https://nhentai.net/g/"+strconv.Itoa(gallery) , nil)
	handleError(err,"getGalleries : req failed.")
	req.Header.Set("user-agent",userAgent)
	req.Header.Set("cookie",cookies)
	resp, _ := httpClient.Do(req)
	if err!=nil {
		handleError(err,"getGalleries : get resp failed(可能是代理设置问题).")
		os.Exit(2)
	}
	defer resp.Body.Close()

	pageBytes,err:=ioutil.ReadAll(resp.Body)
	docs, err := goquery.NewDocumentFromReader(bytes.NewReader(pageBytes))
	var parodies,characters,tags,artists,languages []string
	var page,h2title,h1title,imgUrl string
	if docs!=nil {
		//Parodies
		docs.Find("div.tag-container.field-name:contains(Parodies) span a span.name").Each(func(i int, s *goquery.Selection) {
			parodies=append(parodies, s.Text())
		})
		//Characters
		docs.Find("div.tag-container.field-name:contains(Characters) span a span.name").Each(func(i int, s *goquery.Selection) {
			characters=append(characters, s.Text())
		})
		//Tags
		docs.Find("div.tag-container.field-name:contains(Tags) span a span.name").Each(func(i int, s *goquery.Selection) {
			tags=append(tags, s.Text())
		})
		//Artists
		docs.Find("div.tag-container.field-name:contains(Artists) span a span.name").Each(func(i int, s *goquery.Selection) {
			artists=append(artists, s.Text())
		})
		//Languages
		docs.Find("div.tag-container.field-name:contains(Languages) span a span.name").Each(func(i int, s *goquery.Selection) {
			languages=append(languages, s.Text())
		})
		//Languages
		imgUrl, _ =docs.Find("div#cover a img.lazyload").Attr("data-src")
		page =docs.Find("div.tag-container.field-name:contains(Pages) span a span.name").Text()
		h2title=docs.Find("h2.title").Text()
		h1title=docs.Find("h1.title").Text()
	}
	mapBook[gallery]=book{
		id: gallery,
		imgUrl: string(getGallery(imgUrl,0)),
		parodies:   parodies,
		characters: characters,
		artists:    artists,
		languages:  languages,
		page:       page,
		h2title:    h2title,
		h1title:    h1title,
	}
	_, err = strconv.Atoi(page)

	path=path+checkSpecial(h2title)+"/"
	//if pathExists(path) {
	//	fmt.Println("exist")
	//
	//}else {
	//	//for{
	//	//	fmt.Println(i)
	//	//	if i==intPage {
	//	//		fmt.Println("breakes")
	//	//		break
	//	//	}
	//	//		err:=os.MkdirAll(path,os.ModePerm)
	//	//		if err!=nil {
	//	//			fmt.Println("317error,mkdir err")
	//	//			i++
	//	//		}else {
	//	//			wg.Add(1)
	//	//			go func() {
	//	//				defer wg.Done()
	//	//				//怎么确定图片类型？
	//					chooseType:=0
	//					for{
	//						if chooseType==len(imgType) {
	//							fmt.Println(len(imgType))
	//							break
	//						}
	//						//fmt.Println(gallery,i+1,imgType[chooseType])
	//						status := downloadOneImg(getImgUrl(gallery,i+1,imgType[chooseType]),path+page+imgType[chooseType])
	//						if status {
	//							break
	//						}else {
	//							//fmt.Println(path+page+imgType[chooseType])
	//							chooseType++
	//						}
	//					}
	//	//			}()
	//	//			i++
	//	//	}
	//	//}
	//}
	fmt.Println(imgUrl)
	fmt.Println(path+page+imgType[0])
	fmt.Println(getGallery(imgUrl,0))
	if !pathExists(path) {
		err:=os.MkdirAll(path,os.ModePerm)
		if err!=nil {
			fmt.Println("mkdir err")
		}
	}
	o :=  1
	intPage,_:=strconv.Atoi(mapBook[gallery].page)
	for o<=intPage {
		o2:=o

		wg.Add(1)

		go func() {
			defer wg.Done()
			ensureImgTypeAndDownload(imgUrl,path,o2,0)
		}()
		o++
	}

	wg.Wait()
	elapsed := time.Since(startingTime)
	fmt.Println("completed!",elapsed,"  ",path)
}

func ensureImgTypeAndDownload(imgUrl,path string,o,i int)bool  {
	if i>=2 {
		return false
	}
	fmt.Println(getImgUrl(getGallery(imgUrl,0),o,imgType[i]))
	status := downloadOneImg(getImgUrl(getGallery(imgUrl,0),o,imgType[i]),path+strconv.Itoa(o)+imgType[i])
	if status {
		//fmt.Println("OK")
		return true
	}else {
		s:=ensureImgTypeAndDownload(imgUrl,path,o,i+1)
		return s
	}
}
//v-2020.1.5(以后基本不改了)
//2020.12.4 改：用代理
