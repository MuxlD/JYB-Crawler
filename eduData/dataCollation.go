package eduData

import (
	"JYB_Crawler.Vn/Basics"
	"JYB_Crawler.Vn/elasticsearch"
	"context"
	"log"
	"strconv"
	"time"
)

var (
	chromedpTimeout int //爬取网络超时时间
	done            chan struct{}
	tsCh            chan Basics.TsUrl
	tsCapacity      = 1000            //仓库容量
	pendCh          chan Basics.TsUrl //正常爬取出错的待处理链接
	pendCapacity    = 100
)

type TsCrawler struct {
	ctx context.Context
}

func NewTsCrawler(ctx context.Context) *TsCrawler {
	return &TsCrawler{
		ctx,
	}
}

func StartContext(ctx context.Context, goroutineNum, cralerTimeout int) {
	//构建类型切片，设置通用超时时间
	chromedpTimeout = cralerTimeout
	ts := NewTsCrawler(ctx)
	//获取所有类型链接,为变量everyType赋值
	err := ts.FindAllType(Basics.JYBFL)
	if err != nil {
		log.Println("FindAllType error, info：", err)
		return
	}

	ts.Do(goroutineNum)
}

func (ts *TsCrawler) Do(goroutineNum int) {
	//记录初始ts
	//设置仓库容量
	tsCh = make(chan Basics.TsUrl, tsCapacity)
	pendCh = make(chan Basics.TsUrl, pendCapacity)
	//生产者信号通道
	done = make(chan struct{})

	//生产者函数，对变量everyType的值进行处理,完善type对象空白字段
	go ts.AllLink()
	//建立批量插入任务
	indexCtx, indexCancel := context.WithCancel(context.Background())
	//多消费者，es生产者
	for i := 1; i <= goroutineNum; i++ {
		go ts.Crawler(strconv.Itoa(i), indexCtx, ts.ctx)
		time.Sleep(time.Second * 3)
	}
	err := elasticsearch.BulkInsert(indexCtx)
	if err != nil {
		log.Println("BulkInsert error, info:", err)
		//出错时，关闭批量插入es的任务,主goroutine结束
		indexCancel()
	}
}

//主要爬虫程序
func (ts *TsCrawler) Crawler(chromeId string, indexCtx, ctx context.Context) {
	log.Println("chrome ID", chromeId, ":Successfully entered goroutine...")
	//新的浏览器
	chrome := NewChromedp(ctx)
	//分配浏览器
	contx, cancel := chrome.AssignBrowser(Basics.Allo)

	var stop bool
	var ok, pend bool
	var tsCraw Basics.TsUrl
	for {
		select {
		//indexCancel(),当es插入异常时关闭,批量插入失败
		case <-indexCtx.Done():
			log.Printf("bulk insert or CrawlerByUrl error...\n")
			//关闭浏览器
			cancel()
			chrome.Close()
			return
		case <-done:
			//生产者结束信号
			stop = true
		case tsCraw, ok = <-tsCh:
			//通道执行close()内容数据被全部取出后 tsCraw = nil, ok = false
			//如果通道没有执行close()，当通道内容数据被全部取出后，程序不会再进入该case
			if !ok && stop && !pend {
				log.Printf("所有通道内容已消费完," + chromeId + "号协程退出...\n")
				//所有内容写入完成关闭es写入通道
				close(elasticsearch.Docsc)
				//关闭浏览器
				chrome.Close()
				return
			}
			//消费函数
			err := ts.CrawlerByUrl(tsCraw, contx)
			if err != nil {
				log.Println("tsCh:", err)
			}
		case tsCraw, pend = <-pendCh:
			if !pend {
				log.Printf("pendCh通道内容已消费完")
				continue
			}
			err := crawlerByPendUrl(tsCraw, contx)
			if err != nil {
				log.Println("pendCh:", err)
			}
		}
	}
}
