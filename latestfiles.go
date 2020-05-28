package main

import (
    "sync"
    "time"
    "html/template"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/dynamodb"
    "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var dbSess = session.Must(session.NewSessionWithOptions(session.Options{
    SharedConfigState: session.SharedConfigEnable,
}))
var db = dynamodb.New(dbSess)

type LatestFileName struct {
    Bucket  string
    Channel string
}

type LatestFileInfo struct {
    ObjectKey   string
    RecheckTime time.Time
}

var latestFilesMutex sync.Mutex

var latestFilesCache = make(map[LatestFileName]LatestFileInfo)

func latestFileURL(bucket string, channel string) (template.URL, error) {
    latestFilesMutex.Lock()
    defer latestFilesMutex.Unlock()

    info, ok := latestFilesCache[LatestFileName{bucket, channel}]
    if !ok || time.Now().After(info.RecheckTime) {
        res, err := db.GetItem(&dynamodb.GetItemInput{
            TableName: aws.String("LatestFiles"),
            Key: map[string]*dynamodb.AttributeValue{
                "Bucket": { S: aws.String(bucket) },
                "Channel": { S: aws.String(channel) },
            },
        })
        if err != nil {
            return "", err
        }

        var item struct {
            ObjectKey string
        }

        err = dynamodbattribute.UnmarshalMap(res.Item, &item)
        if err != nil {
            return "", err
        }

        info.ObjectKey = item.ObjectKey
        info.RecheckTime = time.Now().Add(time.Second * 20)
        latestFilesCache[LatestFileName{bucket, channel}] = info
    }

    return template.URL("https://" + bucket + "/" + info.ObjectKey), nil
}

func init() {
    tpl.Funcs(map[string]interface{}{
        "LatestFileURL": latestFileURL,
    })
}
