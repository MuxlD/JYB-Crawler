package eduData

import (
	"JYB_Crawler.Vn/Basics"
	"JYB_Crawler.Vn/elasticsearch"
	"context"
	"errors"
	"fmt"
	"github.com/chromedp/chromedp"
	"log"
	"time"
)

func crawlerByPendUrl(tsCraw Basics.TsUrl, ctx context.Context) (err error) {
	log.Println("当前学校链接：", tsCraw.Url)

	start := time.Now()

	var bts Basics.TrainingSchool
	var retry bool

	//进入具体的信息爬取
	for i := 0; i < 3; i++ {
		bts, retry = everyPend(ctx, tsCraw.Url)
		if !retry { //retry == false 时跳出循环
			break
		}
	}
	//成功赋值字段name,course,brightSpot,info,campus,phoneNumber
	if bts.PhoneNumber == "" {
		err = errors.New("Discard options:" + tsCraw.Url)
		return
	}

	into(&bts, tsCraw)

	fmt.Println(bts)

	elasticsearch.Docsc <- bts

	log.Printf("抓取成功,爬取耗时：%v\n", time.Since(start))

	return
}

func everyPend(goCtx context.Context, url string) (bts Basics.TrainingSchool, retry bool) {

	ctx, cancel := context.WithTimeout(goCtx, time.Duration(chromedpTimeout)*time.Second) //time.Duration(chromedpTimeout)
	defer cancel()

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
	)
	if err != nil {
		//页面加载不成功
		log.Println("链接跳转失败或页面不存在", url, err)
		return bts, true
	}

	err = chromedp.Run(ctx,
		//等待链接页面
		chromedp.WaitVisible(`.page-container`),
	)
	if err != nil {
		//页面加载不成功
		log.Println("验证页面第一条class失败", url, err)
		return bts, true
	}

	err = chromedp.Run(ctx,
		//验证机构名称
		chromedp.JavascriptAttribute(`.header-agency-outer .header-agency-logo img`, "title", &bts.Name),
		//phoneNumber
		chromedp.Text(".header-agency-outer .header-agency-tel span", &bts.PhoneNumber),
	)
	if err != nil {
		log.Println("机构名称加载不成功", err)
		return
	}
	var courseHtml, campusHtml string
	//key options
	err = chromedp.Run(ctx,
		//course
		chromedp.OuterHTML(".not-certified-class-list", &courseHtml),
		//info
		chromedp.Text(".not-certified-intro p", &bts.Info),
		//campus
		chromedp.InnerHTML(".content-right .school-list-container ", &campusHtml),
	)
	bts.Course = Splice(courseHtml, `</h5>
                        <p>(.*)</p>`)
	bts.Campus = SpliceArr(campusHtml, `<h5>(.*)</h5>
                                        <span>(.*)</span>`)

	return
}

func SpliceArr(bodystr string, rxp string) (c string) {
	result := SelfReg(bodystr, rxp)
	for i := range result {
		c = c + result[i][1] + ":" + result[i][2] + "."
	}
	return
}
