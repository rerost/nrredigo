package nrredigo

import (
	"strings"

	"github.com/gomodule/redigo/redis"
	newrelic "github.com/newrelic/go-agent"
)

func wrapConn(c redis.Conn, txn newrelic.Transaction, cfg *Config) redis.Conn {
	return &wrappedConn{
		Conn: c,
		txn:  txn,
		cfg:  cfg,
	}
}

type wrappedConn struct {
	redis.Conn
	txn newrelic.Transaction
	cfg *Config
}

func (c *wrappedConn) Do(commandName string, args ...interface{}) (interface{}, error) {
	if c.txn != nil {
		seg := c.createSegment(commandName)
		seg.ParameterizedQuery = formatCommand(commandName, args)
		defer seg.End()
	}
	return c.Conn.Do(commandName, args...)
}

func (c *wrappedConn) Send(commandName string, args ...interface{}) error {
	if c.txn != nil {
		seg := c.createSegment(commandName)
		seg.ParameterizedQuery = formatCommand(commandName, args)
		defer seg.End()
	}
	return c.Conn.Send(commandName, args...)
}

func (c *wrappedConn) Flush() error {
	if c.txn != nil {
		seg := c.createSegment("flush")
		defer seg.End()
	}
	return c.Conn.Flush()
}

func (c *wrappedConn) Receive() (interface{}, error) {
	if c.txn != nil {
		seg := c.createSegment("receive")
		defer seg.End()
	}
	return c.Conn.Receive()
}

func (c *wrappedConn) createSegment(cmdName string) newrelic.DatastoreSegment {
	return newrelic.DatastoreSegment{
		StartTime:    newrelic.StartSegmentNow(c.txn),
		Product:      newrelic.DatastoreRedis,
		Operaiton:    strings.ToLower(cmdName),
		Host:         c.cfg.Host,
		PortPathOrID: c.cfg.PortPathOrID,
		DatabaseName: c.cfg.DBName,
	}
}
