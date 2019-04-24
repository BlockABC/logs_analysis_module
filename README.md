# logs_analysis_module

## 日志分析模块 使用Elasticsearch 实现，使用Grafana 图表分析


## 使用

```ssh

go get github.com/BlockABC/logs_analysis_module

```

## Demo main.go

```go

    type MyLogger struct{}
    
    func (MyLogger) Infof(format string, params ...interface{}) {
    	//TODO 写自己的log 主件
    }
    
    func (MyLogger) Errorf(format string, params ...interface{}) {
    	//TODO 写自己的log 主件
    }
    
    func main() {
    	//初始化es
    	//esIndex 相当于数据库名称 必须小写
    	//esType  相当于数据库表名称
    	es, err := logs_analysis_module.New("http://localhost:9200/","eospark_test_","api")
    	if err != nil {
    		panic(err)
    	}
    
    	//gin 相关
    	router := gin.New()
    	//logMiddleware := logs_analysis_module.NewRecordRequest(es, MyLogger{}, "[TEST]", true)
    	logMiddleware := logs_analysis_module.NewRecordRequest(es, nil, "[TEST]", true)
    
    	// 使用日志分析插件
    	router.Use(logMiddleware.RecordRequestMiddleware())
    
    	router.GET("/test", func(c *gin.Context) {
    		c.JSON(http.StatusOK, gin.H{"errno": 0, "errmsg": "Success", "data": gin.H{"symbol_list": gin.H{"symbol": "EOS", "code": "eosio.token", "balance": "2.7937"}}})
    	})
    
    	router.Run(":8181")
    }


```

![](./cache.png)