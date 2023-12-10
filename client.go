package psql

import (
	"database/sql"

	"github.com/Pheethy/sqlx"
	pg "github.com/lib/pq"
	"github.com/opentracing/opentracing-go"
	"github.com/qustavo/sqlhooks/v2"
)

type Client struct {
	db            *sqlx.DB
	connectionURI string
	driverName    string
	tracer        opentracing.Tracer
}

func NewPsqlConnection(connectionStr string) (*Client, error) {
	addr, err := pg.ParseURL(connectionStr)
	if err != nil {
		return nil, err
	}

	db, err := sqlx.Connect(postgres_driver, addr)
	if err != nil {
		return nil, err
	}

	return &Client{
		db:            db,
		connectionURI: connectionStr,
		driverName:    postgres_driver,
	}, nil
}

func NewPsqlWithTracingConnection(connectionStr string, tracing opentracing.Tracer) (client *Client, err error) {
	addr, err := pg.ParseURL(connectionStr)
	if err != nil {
		return nil, err
	}

	if !isRegisterOTPG {
		sql.Register(opentracing_driver, sqlhooks.Wrap(&pg.Driver{}, NewTracingHook(tracing)))
	}

	db, err := sqlx.Connect(opentracing_driver, addr)
	if err != nil {
		return nil, err
	}

	isRegisterOTPG = true
	return &Client{
		db:            db,
		connectionURI: connectionStr,
		driverName:    opentracing_driver,
		tracer:        tracing,
	}, nil
}

func (c *Client) GetClient() *sqlx.DB {
	return c.db
}

func (c *Client) GetConnectionURI() string {
	return c.connectionURI
}

func (c *Client) SetDB(db *sqlx.DB) {
	c.db = db
}

func (c *Client) IsConnect() bool {
	if err := c.db.Ping(); err == nil {
		return true
	}
	return false
}

func (c *Client) Reconnect() error {
	if c.IsConnect() {
		return nil
	}

	switch c.driverName {
	case postgres_driver:
		client, err := NewPsqlConnection(c.connectionURI)
		if err != nil {
			return err
		}

		c.db = client.GetClient()
	case opentracing_driver:
		addr, err := pg.ParseURL(c.connectionURI)
		if err != nil {
			return err
		}
		db, err := sqlx.Connect(opentracing_driver, addr)
		if err != nil {
			return err
		}

		c.db = db
	}

	return nil
}
