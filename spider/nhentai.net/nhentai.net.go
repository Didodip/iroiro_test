package nhentai_net

import (
	m "awesomeProject/spider/mail"
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type book struct {
	num int
	title string
	url string
}

var (
	disk = "G://shunxu-/"
	bannedCharacter = [...]string{"?", "*", ":", "\"", "<", ">", "\\", "/", "|"}
	imgType = [...]string{".jpg", ".png", ".gif"}
	wg sync.WaitGroup
)

const (
	userAgent  ="Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.88 Safari/537.36"
	cookies = "__cfduid=df811481cff8f3879712ea5a02b18feb01576840808; _ga=GA1.2.1931960219.1576840808; _gid=GA1.2.1273994293.1577848755"
)
//排错
func handleError (err error, reason string)  {
	if err != nil {
		fmt.Println(reason,err)
	}
}
//返回title和url列表
func getGalleries(page int)(imgList[]book)  {

	req,err:=http.NewRequest("GET","https://nhentai.net/tag/full-color/?page="+strconv.Itoa(page), nil)
	handleError(err,"Http.get.page.req:{GetGalleries}")
	req.Header.Set("user-agent",userAgent)
	req.Header.Set("cookie",cookies)
	resp, err := (&http.Client{}).Do(req)
	if err!=nil {
		fmt.Println("ConnectPageErr.")
		m.Send("HTTP CONNECTION ERROR.PAGE:"+strconv.Itoa(page),"")
		os.Exit(2)
	}
	handleError(err,"Http.get.page.resp:{GetGalleries}")
	defer resp.Body.Close()
	pageBytes,err:=ioutil.ReadAll(resp.Body)

	handleError(err,"ioutil.ReadAll.pageBytes:{GetGalleries}")
	docs, err := goquery.NewDocumentFromReader(bytes.NewReader(pageBytes))
	handleError(err,"goquery.NewDocument:{GetGalleries}")
	var innerImg []book

	docs.Find("div.index-container.container div.gallery a").Each(func(i int, s *goquery.Selection) {
		imgPath, exists := s.Find("img").Attr("data-src")
		//print(s.Text())
		title := s.Find("div.caption").Text()
		//print(imgPath)
		if !exists {
			imgPath, exists := s.Find("img").Attr("src")

			innerImg = append(innerImg, book{
				num:   i,
				title: title,
				url:   imgPath,
			})
			if !exists{
				//print("noo")
				os.Exit(66)
			}
		}else {
			innerImg = append(innerImg, book{
				num:   i,
				title: title,
				url:   imgPath,
			})
		}

	})
	//fmt.Println(innerImg[0])
	return innerImg
}
//图片url
func getImgUrl(gallery int, page int,imgType string)(url string){
	return "https://i.nhentai.net/galleries/"+strconv.Itoa(gallery)+"/"+strconv.Itoa(page)+imgType
}
//下载单张图片
func downloadOneImg(imgUrl, path string)bool{
	resp,err:=http.Get(imgUrl)
	if err!=nil {
		return false
	}
	code:=resp.StatusCode
	if code!=200 {
		return false
	}
	body,err:=ioutil.ReadAll(resp.Body)
	if err!=nil {
		return false
	}
	file, _ := os.Create(path)
	_, err = io.Copy(file, bytes.NewReader(body))
	if err!=nil {
		fmt.Println(err)
		return false
	}
	return true
}
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
func getGallery(url string, num int) int {
	re, _ := regexp.Compile(`\d+`)
	all := re.FindAll([]byte(url),-1)
	if len(all)<num+1 {
		fmt.Println("can't get gallery id(List Out Of Range):{GetGallery}")
		os.Exit(0)
		return 0
	}else {
		a,_:= strconv.Atoi(string(all[num]))
		return a
	}
}


func StartSpider(start,end int)  {
	var h =start
	for {
		if h==end {
			break
		}
		SpiderOnePage(h)
		h++
	}
}
//爬一页面
func SpiderOnePage(h int)  {
	start := time.Now()
	gs:=getGalleries(h)
	//SpiderOnePage
	i :=  0
	////fmt.Println("page",h)
	fmt.Println("now page",h)
	for {
		if i==len(gs) {
			break
		}
		title := reflectReadBook(gs[i],"title")
		url := reflectReadBook(gs[i],"url")
		gallery := getGallery(url,0)
		path := disk+checkSpecial(title)+"/"
		//chooseType := 0
		if pathExists(path) {
			////fmt.Println(title,"---exists")
			i++
		}else {
			fmt.Println(title,"-")
			err:=os.MkdirAll(path,os.ModePerm)
			if err!=nil {
				fmt.Println(err)
				os.Exit(1)
			}
			//协程：一页上的本子25个协程爬取。
			//spiderOneBook

			wg.Add(1)
			go func() {
				defer wg.Done()
				SpiderOneBook(gallery,path)
				//parseUrls("https://movie.douban.com/top250?start="+strconv.Itoa(25*i))
			}()
			////fmt.Print("--completed.")
			////fmt.Println()
			i++
		}
	}
	wg.Wait()
	elapsed := time.Since(start)
	fmt.Println("Took", elapsed)
}
//爬一本
func SpiderOneBook(gallery int,path string)  {
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
		if !status {
			et+=1
			if chooseType<len(imgType)-1 {
				chooseType+=1
			}else {
				chooseType-=len(imgType)-1
			}
			//fmt.Println(status,imgUrl,title)
		}else {
			////fmt.Print(p," ")
			p++
			et=0
		}
		//fmt.Println(status,imgUrl)
	}
}
//v-2020.1.5(以后基本不改了)
