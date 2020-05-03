package eduData

import (
	"JYB_Crawler.Vn/Basics"
	"JYB_Crawler.Vn/elasticsearch"
	"errors"
	"github.com/jinzhu/gorm"
	"log"
)

func (ts *TsCrawler) AllLink() {
	db := Basics.GetDB()
	db.Model(Basics.Type{}).Find(&Basics.EveryType)
	if Basics.EveryType[19].Count != 0 {
		//如果最后一种类型的机构数量不为0；
		//说明所有的链接已全部存入数据库；
		//可直接从数据库中取出放入通道tsCh。
		err := getAllEduInDB(db)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		ts.getAllEdu(db)
	}
	//关闭通道，通知所有类目下的商品获取完成
	close(done)
	log.Println("发送第一生产者结束信号,url获取结束...")
	//将完善过的type对象批量插入es
	err := elasticsearch.TpBulkInsert()
	if err != nil {
		log.Println("TpBulkInsert error,info:", err)
		return
	}
	log.Println("将完善过的type对象批量插入es")
}

func getAllEduInDB(db *gorm.DB) error {
	var dbAllTs []Basics.TsUrl
	var total int
	db = db.Model(Basics.TsUrl{}).Count(&total)
	//验证数据库中是否有数据
	if total <= 0 {
		return errors.New("table ts_urls is empty")
	}
	db.Find(&dbAllTs)
	for _, tst := range dbAllTs {
		tsCh <- tst
	}
	close(tsCh)
	return nil
}
