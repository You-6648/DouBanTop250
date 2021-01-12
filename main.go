package main

import (
	"DouBanTop250/model"
	"fmt"
	"github.com/gocolly/colly"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)


func main()  {
	defer model.Db.Close()
	testcollydom()
	//GetTopDetail("")

	//FetchBase()
}

//豆瓣图书标签
//var url = "https://book.douban.com/tag/?view=type"

var url = "https://movie.douban.com/subject/1292052/"

func FetchBase(){
	client := &http.Client{}
	req,err := http.NewRequest("GET",url,nil)

	if err != nil {
		fmt.Printf("")
	}
	req.Header.Set("User-Agent","Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.141 Safari/537.36")


	resp,err := client.Do(req)

	result,err := ioutil.ReadAll(resp.Body)

	fmt.Printf("%s",result)

}


/*
Collector对象接受多种回调方法，有不同的作用，按调用顺序我列出来：
OnRequest。请求前
OnError。请求过程中发生错误
OnResponse。收到响应后
OnHTML。如果收到的响应内容是HTML调用它。
OnXML。如果收到的响应内容是XML 调用它。写爬虫基本用不到，所以上面我没有使用它。
OnScraped。在OnXML/OnHTML回调完成后调用。不过官网写的是Called after OnXML callbacks，实际上对于OnHTML也有效，大家可以注意一下。
*/

func testcollydom(){
	//创建新的采集器
	c := colly.NewCollector(
		//这次在colly.NewCollector里面加了一项colly.Async(true)，表示抓取时异步的
		colly.Async(true),
		//模拟浏览器
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36"),
	)
	//限制采集规则
	//在Colly里面非常方便控制并发度，只抓取符合某个(些)规则的URLS，有一句c.Limit(&colly.LimitRule{DomainGlob: "*.douban.*", Parallelism: 5})，表示限制只抓取域名是douban(域名后缀和二级域名不限制)的地址，当然还支持正则匹配某些符合的 URLS，具体的可以看官方文档。
	c.Limit(&colly.LimitRule{DomainGlob: "*.douban.*",Parallelism:3,Delay:time.Duration(3)*time.Second})
	/*
		另外Limit方法中也限制了并发是5。为什么要控制并发度呢？因为抓取的瓶颈往往来自对方网站的抓取频率的限制，如果在一段时间内达到某个抓取频率很容易被封，所以我们要控制抓取的频率。另外为了不给对方网站带来额外的压力和资源消耗，也应该控制你的抓取机制。
	*/

	c.OnRequest(func(r *colly.Request) {

		//设置cookie
		//r.Headers.Set("Cookie", "")
		fmt.Println("Visiting", r.URL)
	})

	c.OnError(func(_ *colly.Response, err error) {
		fmt.Println("Something went wrong:", err)
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Visited", r.Request.URL)
	})

	c.OnHTML(".hd", func(e *colly.HTMLElement) {
		//log.Println(strings.Split(e.ChildAttr("a", "href"), "/")[4],
		//	strings.TrimSpace(e.DOM.Find("span.title").Eq(0).Text()))
		fmt.Println(strings.Split(e.ChildAttr("a", "href"), "/")[4])
		GetTopDetail(strings.Split(e.ChildAttr("a", "href"), "/")[4])
	})

	c.OnHTML(".paginator a", func(e *colly.HTMLElement) {
		e.Request.Visit(e.Attr("href"))
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished", r.Request.URL)
		time.Sleep(time.Duration(5)*time.Second) //让当前线程睡眠2秒钟
	})

	c.Visit("https://movie.douban.com/top250?start=0&filter=")
	c.Wait()
}

//获取每一个页面的详情
func GetTopDetail(str string) {

	//声明Detail
	var detail model.Top250Detail

	detail.DetailId = str

	//创建新的采集器
	c := colly.NewCollector(
		//这次在colly.NewCollector里面加了一项colly.Async(true)，表示抓取时异步的
		colly.Async(true),
		//模拟浏览器
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36"),
	)


	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.OnError(func(_ *colly.Response, err error) {
		fmt.Println("Something went wrong:", err)
	})

	//获取标题
	c.OnHTML("h1", func(e *colly.HTMLElement) {
		detail.Title = strings.Replace(e.Text,"\n        ","",-1)
	})

	//获取电影详情
	c.OnHTML("#info", func(e *colly.HTMLElement) {
		e.ForEach(".attrs", func(i int, ele *colly.HTMLElement) {
			switch i {
			case 0: detail.Director = ele.Text
			case 1: detail.ScreenWriter = ele.Text
			case 2: detail.MainPlaying = ele.Text
			}
		})
		str,_ := e.DOM.Html()
		detailSplit := strings.Split(str,"<br/>")
		for _, value := range detailSplit {

			if strings.Contains(value,"类型:") {
				typeStr := strings.Split(value,"<span property=\"v:genre\">")

				for key, value := range typeStr {
					if key > 0 {
						value = strings.Replace(value,"</span>","",-1)
						detail.Type +=strings.Trim(value," ")
					}
				}
			}else if strings.Contains(value,"官方网站:") {
				officeweb :=strings.Split(value,"target=\"_blank\">")[1]
				detail.OfficialWebsite = strings.Replace(officeweb,"</a>","",-1)
				fmt.Println(detail.OfficialWebsite)
			}else if strings.Contains(value,"制片国家/地区:") {
				detail.ProductionCountry = strings.Split(value,"</span>")[1]
			}else if strings.Contains(value,"语言:") {
				detail.Language = strings.Split(value,"</span>")[1]
			}else if strings.Contains(value,"上映日期:") {
				dataStr := strings.Split(value,"content=\"")
				for key, value := range dataStr {
					if key > 0 {
						value =  strings.Split(value,"\">")[0]
						detail.ShowTime += value + "/"
					}
				}
				if len(detail.ShowTime) >0 {
					detail.ShowTime = detail.ShowTime[0:len(detail.ShowTime) - 1]
				}
			}else if strings.Contains(value,"片长:") {
				durationSplit := strings.Split(value,"content=\"")[1]
				detail.MovieDuration = strings.Replace(durationSplit,"</span><br/>","",-1)
				detail.MovieDuration = strings.Replace(detail.MovieDuration,"</span>","",-1)
				detail.MovieDuration = strings.Split(detail.MovieDuration,">")[1]
			}else if strings.Contains(value,"又名:") {
				detail.OtherName = strings.Split(value,"</span>")[1]
			}else if strings.Contains(value,"IMDb链接:") {
				link := strings.Split(value," rel=\"nofollow\">")[1]
				detail.IMDB = strings.Replace(link,"</a>","",-1)
			}
		}
	})

	//评分
	c.OnHTML("#interest_sectl", func(element *colly.HTMLElement) {
		element.ForEach(".ratings-on-weight", func(i int, e *colly.HTMLElement) {
			e.ForEach(".item", func(i int, star *colly.HTMLElement) {
				switch i {
				case 0:
					detail.FiveStar =  strings.Replace(star.Text,"\n        ","",-1)
					detail.FiveStar = strings.Replace(detail.FiveStar,"    5星","",-1)
				case 1:
					detail.FourStar = strings.Replace(star.Text,"\n        ","",-1)
					detail.FourStar = strings.Replace(detail.FourStar,"    4星","",-1)
				case 2:
					detail.ThreeStar =strings.Replace(star.Text,"\n        ","",-1)
					detail.ThreeStar = strings.Replace(detail.ThreeStar,"    3星","",-1)
				case 3:
					detail.TwoStar = strings.Replace(star.Text,"\n        ","",-1)
					detail.TwoStar = strings.Replace(detail.TwoStar,"    2星","",-1)
				case 4:
					detail.OneStar = strings.Replace(star.Text,"\n        ","",-1)
					detail.OneStar = strings.Replace(detail.OneStar,"    1星","",-1)
				}
			})
		})
		//评分
		element.ForEach(".rating_num ", func(i int, el *colly.HTMLElement) {
			detail.Score = el.Text
		})

		//多少人评价
		element.ForEach(".rating_sum ", func(i int, el *colly.HTMLElement) {
			detail.NumberEvaluate = strings.Replace(el.Text,"\n                ","",-1)
		})
	})

	//获取简介
	c.OnHTML("#link-report", func(e *colly.HTMLElement) {
		detail.Introduction = strings.Replace(e.Text,"\n                ","",-1)
		detail.Introduction = strings.Trim(detail.Introduction," ")
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished", r.Request.URL)
		time.Sleep(time.Duration(5)*time.Second) //让当前线程睡眠2秒钟
	})
	time.Sleep(time.Duration(5)*time.Second) //让当前线程睡眠5秒钟
	c.Visit("https://movie.douban.com/subject/"+str+"/")
	time.Sleep(time.Duration(5)*time.Second) //让当前线程睡眠5秒钟
	c.Wait()
	AddTop250Detail(&detail)
}

//处理详情
func HandleDetail (html string) string {
	var returnDetail = ""
	if strings.Contains(html,"类型:") {
		typeStr := strings.Split(html,"<span property=\"v:genre\">")

		for key, value := range typeStr {
			if key > 0 {
				value = strings.Replace(value,"</span>","",-1)
				returnDetail +=strings.Trim(value," ")
			}
		}
	}else if strings.Contains(html,"官方网站:") {
		officeweb :=strings.Split(html,"target=\"_blank\">")[1]
		returnDetail = strings.Replace(officeweb,"</a><br/>","",-1)
	}else if strings.Contains(html,"制片国家/地区:") {
		returnDetail = strings.Split(html,"</span>")[1]
	}else if strings.Contains(html,"语言:") {
		returnDetail = strings.Split(html,"</span>")[1]
	}else if strings.Contains(html,"上映日期:") {
		dataStr := strings.Split(html,"content=\"")
		for key, value := range dataStr {
			if key > 0 {
				value =  strings.Split(value,"\">")[0]
				returnDetail += value + "/"
			}
		}
		if len(returnDetail) >0 {
			returnDetail = returnDetail[0:len(returnDetail) - 1]
		}
	}else if strings.Contains(html,"片长:") {
		durationSplit := strings.Split(html,"content=\"")[1]
		returnDetail = strings.Replace(durationSplit,"</span><br/>","",-1)
		returnDetail = strings.Replace(returnDetail,"</span>","",-1)
		returnDetail = strings.Split(returnDetail,">")[1]
	}else if strings.Contains(html,"又名:") {
		returnDetail = strings.Split(html,"</span>")[1]
	}else if strings.Contains(html,"IMDb链接:") {
		link := strings.Split(html," rel=\"nofollow\">")[1]
		returnDetail = strings.Replace(link,"</a>","",-1)
	}
	return returnDetail
}

//将详情添加到数据库中
func AddTop250Detail(m *model.Top250Detail) {
	fmt.Println(m.Introduction)

	r, err := model.Db.Exec("insert into t_douban_top250movie" +
		"(Title,Director, ScreenWriter,MainPlaying,Type,OfficialWebsite,ProductionCountry,Language,ShowTime,MovieDuration" +
		",OtherName,IMDB, Score, NumberEvaluate, FiveStar, FourStar, ThreeStar,TwoStar, OneStar,Introduction,DetailId) " +
		" values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ",
		m.Title, m.Director, m.ScreenWriter, m.MainPlaying, m.Type, m.OfficialWebsite, m.ProductionCountry, m.Language,
		m.ShowTime, m.MovieDuration, m.OtherName, m.IMDB, m.Score, m.NumberEvaluate, m.FiveStar, m.FourStar, m.ThreeStar,m.TwoStar, m.OneStar, m.Introduction,m.DetailId)
	if err != nil {
		fmt.Println("exec failed, ", err)
		return
	}
	id, err := r.LastInsertId()
	if err != nil {
		fmt.Println("exec failed, ", err)
		return
	}
	fmt.Println("insert succ:", id)
}

