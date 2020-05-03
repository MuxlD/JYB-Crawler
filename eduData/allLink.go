package eduData

import (
	"JYB_Crawler.Vn/Basics"
	"context"
	"errors"
	"fmt"
	"github.com/chromedp/chromedp"
	"github.com/jinzhu/gorm"
	"log"
	"strconv"
	"time"
)

var maxID uint

//生产者函数
func (ts *TsCrawler) getAllEdu(db *gorm.DB) {
	defer close(tsCh)
	//清空数据
	db.Exec("TRUNCATE TABLE ts_urls;")

	chrome := NewChromedp(ts.ctx)
	defer chrome.Close()
	ctx, cancel := chrome.AssignBrowser(Basics.Allo)
	defer cancel()

	//便利出所有的类型
	for id, ets := range Basics.EveryType {
		//方便程序添加主键值
		err := db.Table("ts_urls").Select("max(id)").Row().Scan(&maxID)
		if err != nil {
			log.Println(err)
			//继续下一种类型
			continue
		}
		start := time.Now()

		//获取类型最大页码
		maxPa, err := maxPage(ctx, ets)
		if err != nil {
			log.Println(err)
			//继续下一种类型
			continue
		}
		//update max_page from types where id = id+1
		db.Model(&Basics.Type{}).Where("id = ?", id+1).Update("max_page", maxPa)
		fmt.Println("最大页码为", maxPa)

		var oneUrl Basics.TsUrl
		oneUrl.TypeID = ets.ID

		var count int
		//分页加载
		for n := 1; n <= maxPa; n++ {
			urlHtml, err := pageLink(ctx, ets.TypeUrl, n)
			if err != nil {
				log.Println("getAllEdu html提取失败", err)
				//继续下一页
				continue
			}
			selectUrl(urlHtml,
				`href="(.*)" target="_blank" class="office-rlist-name"`,
				oneUrl, db, &count,
			)
		}
		//update count from types where id = id+1
		db.Model(&Basics.Type{}).Where("id = ?", id+1).Update("count", count)
		log.Printf("抓取成功:%v，爬取耗时：%v\n", ets.TypeUrl, time.Since(start))
	}
}

func pageLink(ctx context.Context, url string, n int) (urlHtml string, err error) {
	ctx0, cancel0 := context.WithTimeout(ctx, time.Duration(chromedpTimeout)*time.Second)
	defer cancel0()

	nowPageUrl := url + "p" + strconv.Itoa(n) + ".html"
	fmt.Println("当前链接：", nowPageUrl)
	err = chromedp.Run(ctx0,
		chromedp.Navigate(nowPageUrl),
		chromedp.WaitVisible(`.mt10`),
		chromedp.OuterHTML(`.mt10 .office-result-list`, &urlHtml),
	)
	return
}

//提取url
func selectUrl(html string, reg string, tst Basics.TsUrl, db *gorm.DB, count *int) {

	result := SelfReg(html, reg)
	for i := range result {
		tst.Url = Basics.JYB + result[i][1]
		*count++
		maxID++
		tst.ID = maxID
		db.Create(&tst)
		tsCh <- tst
	}
}

//最大页码查询
func maxPage(ctx context.Context, ets Basics.Type) (maxPa int, err error) {
	ctx0, cancel := context.WithTimeout(ctx, time.Duration(chromedpTimeout)*time.Second)
	defer cancel()
	err = chromedp.Run(ctx0,
		//页面跳转
		chromedp.Navigate(ets.TypeUrl),
		// 存在类型组，说明成功进入
		chromedp.WaitVisible(`.mt10`),
	)
	if err != nil {
		log.Println("类型首页加载失败...", ets.TypeUrl)
		return
	}
	//页面验证，找出最大页码
	var xPage string //最大页码
	err = chromedp.Run(ctx,
		//获取最大页码
		chromedp.Text(`.mt10 .pagination li:nth-last-child(2) a`, &xPage),
	)
	if err != nil {
		err = errors.New("max_page获取失败")
		return
	}
	maxPa, _ = strconv.Atoi(xPage)
	if maxPa == 0 {
		maxPa = 1
		//页面加载不成功
		log.Println("该类型的最大页码可能为：", maxPa)
	}
	//mul为multipler,用于设置超时倍数
	return
}
