package common

import (
	"errors"
	"fmt"
	client "github.com/influxdata/influxdb1-client/v2"
	"inner/conf/platform_conf"
	"time"
)

type InfluxDb struct {
	Cli      client.Client
	Database string
}

func ConnInflux() client.Client {
	defer func() {
		if r := recover(); r != nil {
			Log.Error(errors.New(fmt.Sprint(r)))
		}
	}()
	cli, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     cf.InfluxdbConfig["addr"].(string),
		Username: cf.InfluxdbConfig["username"].(string),
		Password: cf.InfluxdbConfig["password"].(string),
	})
	if err != nil {
		fmt.Println(err)
	}
	return cli
}

func (Influx *InfluxDb) Query(cmd string, tz bool) (res []client.Result, err error) {
	defer func() {
		if r := recover(); r != nil {
			Log.Error(errors.New(fmt.Sprint(r)))
			platform_conf.Qch <- 1
		}
	}()
	if tz {
		cmd = cmd + (" tz('Asia/Shanghai')")
	}
	q := client.Query{
		Command:  cmd,
		Database: Influx.Database,
	}
	if response, err := Influx.Cli.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}
	return res, nil
}

func (Influx *InfluxDb) WritesPoints(measurement string, tags map[string]string, fields map[string]interface{}) error {
	defer func() {
		if r := recover(); r != nil {
			Log.Error(errors.New(fmt.Sprint(r)))
			platform_conf.Qch <- 1
		}
	}()
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  Influx.Database,
		Precision: "s", //精度，默认ns
	})
	pt, err := client.NewPoint(measurement, tags, fields, time.Now())
	bp.AddPoint(pt)
	err = Influx.Cli.Write(bp)
	return err
}

func (Influx *InfluxDb) CreateDatabase(database string) error {
	defer func() {
		if r := recover(); r != nil {
			Log.Error(errors.New(fmt.Sprint(r)))
			platform_conf.Qch <- 1
		}
	}()
	q := client.NewQuery("CREATE DATABASE "+database, "", "")
	res, _ := Influx.Cli.Query(q)
	if res.Error() != nil {
		Log.Error(database + ": Create database failed")
		return res.Error()
	}
	return nil
}
func (Influx *InfluxDb) CreatePolicy(database, duration string) error {
	defer func() {
		if r := recover(); r != nil {
			Log.Error(errors.New(fmt.Sprint(r)))
			platform_conf.Qch <- 1
		}
	}()
	q := client.NewQuery("CREATE RETENTION POLICY "+database+" ON "+database+" DURATION "+duration+" REPLICATION 1 SHARD DURATION "+duration+" DEFAULT", "", "")
	res, _ := Influx.Cli.Query(q)
	if res.Error() != nil {
		Log.Error(database + ": Create policy failed")
		return res.Error()
	}
	return nil
}

func (Influx *InfluxDb) CreateContinuousQuery(database, FromMeasurement, ToMeasurement, Duration string) error {
	defer func() {
		if r := recover(); r != nil {
			Log.Error(errors.New(fmt.Sprint(r)))
			platform_conf.Qch <- 1
		}
	}()
	q := client.NewQuery("CREATE CONTINUOUS QUERY \""+ToMeasurement+"\" on \""+
		database+"\" BEGIN SELECT max(*) INTO \""+
		ToMeasurement+"\" FROM \""+FromMeasurement+"\" GROUP BY time("+Duration+"),* END", "", "")
	_, err = Influx.Cli.Query(q)
	return err
}
