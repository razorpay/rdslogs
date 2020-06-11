package tracker

import (
	"github.com/gomodule/redigo/redis"
	"github.com/honeycombio/rdslogs/config"

	log "github.com/sirupsen/logrus"
)

// Tracker is an interface to store the marker and other logFile related information
type Tracker interface {
	// Read and Write latest marker
	ReadLatestMarker(dbname string) string
	WriteLatestMarker(dbname string, marker string)
}

// RedisTracker it supports redis backend
type RedisTracker struct {
	Key   string
	Value string
	Pool  *redis.Pool
}

//RedisConn ....
type RedisConn struct {
	c redis.Conn
}

//ReadLatestMarker read the marker
func (r *RedisTracker) ReadLatestMarker(dbname string) string {
	conn := RedisConn{c: r.Pool.Get()}
	data, _ := conn.get(dbname)
	defer conn.c.Close()
	return data
}

//WriteLatestMarker read the marker
func (r *RedisTracker) WriteLatestMarker(dbname string, marker string) {
	conn := RedisConn{c: r.Pool.Get()}
	conn.set(dbname, marker)
	defer conn.c.Close()
}

// NewPool ....
func NewPool() *redis.Pool {
	log.Info("Creating Connection")
	dbconfig := config.RedisDBConfig

	return &redis.Pool{
		// Maximum number of idle connections in the pool.
		MaxIdle: dbconfig.MaxIdle,
		// max number of connections
		MaxActive: dbconfig.MaxActive,
		// Dial is an application supplied function for creating and
		// configuring a connection.
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial(
				"tcp",
				dbconfig.Host+":"+dbconfig.Port,
				redis.DialPassword(dbconfig.Password),
				redis.DialDatabase(dbconfig.Database))
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
}

// set executes the redis SET command
func (rc RedisConn) set(dbname string, marker string) error {
	_, err := rc.c.Do("SET", dbname+".marker", marker)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (rc RedisConn) get(dbname string) (string, error) {
	// Simple GET example with String helper
	key := dbname + ".marker"
	s, err := redis.String(rc.c.Do("GET", key))
	if err != nil {
		log.Error(err)
		return "", err
	}
	return s, nil
}
