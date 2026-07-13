package init

import (
	"common/config"
	"fmt"

	"github.com/olivere/elastic/v7"
)

func ElasticInit() {
	data := config.DataConfig.Elastic
	var err error
	config.Esc, err = elastic.NewClient(elastic.SetURL(data.Host), elastic.SetSniff(false))
	if err != nil {
		panic(err)
	}
	fmt.Println("elastic链接成功")
}
