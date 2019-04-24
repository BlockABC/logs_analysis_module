package logs_analysis_module

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ApiResp struct {
	Errno  int64       `json:"errno"`
	Errmsg string      `json:"errmsg"`
	Data   interface{} `json:"data"`
}

type tApiLog struct {
	ApiPtah  string        `json:"path"`
	Ip       string        `json:"ip"`
	Latency  time.Duration `json:"latency"`
	Code     int           `json:"errno"`
	LctTime  time.Time     `json:"lctTime"`
	CallTime int64         `json:"call_time"`
	Result   string        `json:"errmsg"`
}

type AnalysisLogger interface {
	// Infof formats message according to format specifier
	// and writes to log with level = Info.
	Infof(format string, params ...interface{})

	// Errorf formats message according to format specifier
	// and writes to log with level = Error.
	Errorf(format string, params ...interface{})
}

type Logger struct{}

func (Logger) Infof(format string, params ...interface{}) {
	log.Printf(format, params...)
}

func (Logger) Errorf(format string, params ...interface{}) {
	log.Printf(format, params...)
}

type RecordRequest struct {
	//es
	es        *elasticClient
	//log
	logger    AnalysisLogger
	//log prefix
	logPrefix string
	// Do I need to print?
	logMode   bool
}

func NewRecordRequest(es *elasticClient, logger AnalysisLogger, logPrefix string, logMode bool) *RecordRequest {
	if logger == nil {
		logger = Logger{}
	}
	if logPrefix == "" {
		logPrefix = "[EOSPARK API]"
	}
	return &RecordRequest{
		es:        es,
		logger:    logger,
		logPrefix: logPrefix,
		logMode:   logMode,
	}
}

func (rq *RecordRequest) SetLogger(logger AnalysisLogger) {
	if logger == nil {
		logger = Logger{}
	}
	rq.logger = logger
}

func (rq *RecordRequest) LogMode(enable bool) *RecordRequest {
	rq.logMode = enable
	return rq
}

func (rq *RecordRequest) SetLogPrefix(logPrefix string) *RecordRequest {
	if logPrefix == "" {
		logPrefix = "[EOSPARK API]"
	}
	rq.logPrefix = logPrefix
	return rq
}

func (rq *RecordRequest) RecordRequestMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		blw := &bufferedWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request

		// Log only when path is not being skipped
		// Stop timer
		c.Next()
		end := time.Now()
		latency := end.Sub(start)

		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		comment := c.Errors.ByType(gin.ErrorTypePrivate).String()
		if raw != "" {
			path = path + "?" + raw
		}
		defer func() {
			if blw.body.String() == "" {
				return
			}
			var api ApiResp
			if err := json.Unmarshal(blw.body.Bytes(), &api); err != nil {
				api.Errmsg = "Result parsing failed. " + blw.body.String()
				api.Errno = -1
			}


			var apiLog tApiLog
			apiLog.ApiPtah = path
			apiLog.Latency = latency / 1000
			apiLog.Code = int(api.Errno)
			apiLog.Ip = clientIP
			apiLog.Result = api.Errmsg
			apiLog.LctTime = time.Now()
			apiLog.CallTime = time.Now().UnixNano()
			esClient := rq.es.elasticClient
			if esClient != nil && esClient.Index() != nil {
				response, err := esClient.
					Index().
					Id(fmt.Sprintf("%d", time.Now().UnixNano())).
					Index(rq.es.esIndex + time.Now().Format("2006-01-02")).
					Type(rq.es.esType).
					BodyJson(apiLog).
					Do(context.Background())
				log.Print(response,err)
			}
		}()


		if statusCode != http.StatusOK {
			if rq.logMode {
				rq.logger.Errorf(rq.logPrefix+" %v|%3d|%s|%4s|%s|%s  %s",
					end.Format("2006/01/02-15:04:05"),
					statusCode,
					latency,
					clientIP,
					method,
					path,
					comment,
				)
			}
		} else {
			if rq.logMode {
				rq.logger.Infof(rq.logPrefix+" %v|%3d|%s|%4s|%s|%s  %s",
					end.Format("2006/01/02-15:04:05"),
					statusCode,
					latency,
					clientIP,
					method,
					path,
					comment,
				)
			}
		}
	}
}
