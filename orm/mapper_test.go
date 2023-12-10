package orm_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	helperModel "git.innovasive.co.th/backend/models"
	"github.com/Pheethy/psql"
	"github.com/Pheethy/psql/orm"
	"github.com/Pheethy/sqlx"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v2"
)

type Order struct {
	TableName struct{}               `json:"-" db:"orders" pk:"ID"`
	ID        *uuid.UUID             `json:"id" db:"id" type:"uuid"`
	Type      string                 `json:"type" db:"type" type:"string"`
	Name      string                 `json:"name" db:"name" type:"string"`
	Ppu       float64                `json:"ppu" db:"ppu" type:"float64"`
	Status    int                    `json:"status" db:"status" type:"int32"`
	Enable    bool                   `json:"enable" db:"enable" type:"bool"`
	OrderDate *helperModel.Date      `json:"order_date" db:"order_date" type:"date"`
	CreatedAt *helperModel.Timestamp `json:"created_at" db:"created_at" type:"timestamp"`
	ChefID    helperModel.ZeroUUID   `json:"-" db:"chef_id" type:"zerouuid"`

	Chef     *Chef      `json:"chef" db:"-" fk:"fk_field1:ChefID,fk_field2:ID"`
	Toppings []*Topping `json:"toppings" db:"-" fk:"fk_field1:ID,fk_field2:OrderId"`
	Batters  []*Batter  `json:"batters" db:"-" fk:"fk_field1:ID,fk_field2:OrderId"`
}

type Batter struct {
	TableName struct{}   `json:"-" db:"batters" pk:"ID"`
	ID        string     `json:"id" db:"id" type:"string"`
	Type      string     `json:"type" db:"type" type:"string"`
	OrderId   *uuid.UUID `json:"-" db:"order_id" type:"uuid"`
}

type Topping struct {
	TableName struct{}   `json:"-" db:"toppings" pk:"ID"`
	ID        int        `json:"id" db:"id" type:"int32"`
	Type      string     `json:"type" db:"type" type:"string"`
	OrderId   *uuid.UUID `json:"order_id" db:"order_id" type:"uuid"`
}

type Chef struct {
	TableName struct{}   `json:"-" db:"chefs" pk:"ID"`
	ID        *uuid.UUID `json:"id" db:"id" type:"uuid"`
	Name      string     `json:"name" db:"name" type:"string"`
}

func (t Order) String() string {
	bu, _ := json.Marshal(t)
	return string(bu)
}
func (t Batter) String() string {
	bu, _ := json.Marshal(t)
	return string(bu)
}
func (t Topping) String() string {
	bu, _ := json.Marshal(t)
	return string(bu)
}
func (t Chef) String() string {
	bu, _ := json.Marshal(t)
	return string(bu)
}

func TestMapper(t *testing.T) {
	var getData = func(t *testing.T, path string) []*Order {
		bu, err := ioutil.ReadFile(path)
		if err != nil {
			t.Error(err)
		}
		var orders = []*Order{}
		if err := json.Unmarshal(bu, &orders); err != nil {
			t.Error(err)
		}
		return orders
	}

	var runQuery = func(t *testing.T, client *psql.Client, sql string, options ...orm.MapperOption) []*Order {
		rows, err := client.GetClient().Queryx(sql)
		if err != nil {
			t.Error(err)
		}
		defer rows.Close()
		if options == nil {
			options = append(options, orm.NewMapperOption())
		}
		mapper, err := orm.Orm(new(Order), rows, options[0])
		if err != nil {
			t.Error(err)
		}
		return mapper.GetData().([]*Order)
	}

	t.Run("success_with_nodata", func(t *testing.T) {
		db, dbmock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}

		sqlxDB := sqlx.NewDb(db, "sqlmock")
		defer db.Close()

		client := new(psql.Client)
		client.SetDB(sqlxDB)

		rows := sqlmock.NewRows([]string{
			"orders.id", "orders.type", "orders.name", "orders.ppu", "orders.status", "orders.enable", "orders.order_date", "orders.created_at",
		})

		sql := `SELECT (.+) orders`
		dbmock.ExpectQuery(sql).WillReturnRows(rows)

		epOrders := runQuery(t, client, sql)
		assert.Equal(t, len(epOrders), 0)
	})

	t.Run("success_with_order_struct_only", func(t *testing.T) {
		orders := getData(t, "../testdata/orders.json")
		db, dbmock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}

		sqlxDB := sqlx.NewDb(db, "sqlmock")
		defer db.Close()

		client := new(psql.Client)
		client.SetDB(sqlxDB)

		rows := sqlmock.NewRows([]string{
			"orders.id", "orders.type", "orders.name", "orders.ppu", "orders.status", "orders.enable", "orders.order_date", "orders.created_at",
		})

		for _, order := range orders {
			rows.AddRow(
				order.ID, order.Type, order.Name, order.Ppu, order.Status, order.Enable, time.Time(*order.OrderDate), order.CreatedAt.ToTime(),
			)
		}

		sql := `SELECT (.+) orders`
		dbmock.ExpectQuery(sql).WillReturnRows(rows)

		epOrders := runQuery(t, client, sql)
		assert.Equal(t, len(epOrders), len(orders))

		for index := range epOrders {
			assert.Equal(t, epOrders[index].ID, orders[index].ID)
			assert.Equal(t, epOrders[index].Name, orders[index].Name)
			assert.Equal(t, epOrders[index].Ppu, orders[index].Ppu)
			assert.Equal(t, epOrders[index].Status, orders[index].Status)
			assert.Equal(t, epOrders[index].Type, orders[index].Type)
			assert.Equal(t, epOrders[index].Enable, orders[index].Enable)
			assert.Equal(t, epOrders[index].CreatedAt.String(), orders[index].CreatedAt.String())
			assert.Equal(t, epOrders[index].OrderDate.String(), orders[index].OrderDate.String())

			assert.NotNil(t, epOrders[index].ID)
			assert.NotZero(t, epOrders[index].Name)
			assert.NotZero(t, epOrders[index].Ppu)
			assert.NotZero(t, epOrders[index].Status)
			assert.NotZero(t, epOrders[index].Type)
			assert.NotZero(t, epOrders[index].Enable)
			assert.NotNil(t, epOrders[index].CreatedAt.String())
			assert.NotNil(t, epOrders[index].OrderDate.String())
		}
	})

	t.Run("success_with_order_left_join_topping", func(t *testing.T) {
		orders := getData(t, "../testdata/orders.json")
		db, dbmock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}

		sqlxDB := sqlx.NewDb(db, "sqlmock")
		defer db.Close()

		client := new(psql.Client)
		client.SetDB(sqlxDB)

		rows := sqlmock.NewRows([]string{
			"orders.id", "orders.type", "orders.name", "orders.ppu", "orders.status", "orders.enable", "orders.order_date", "orders.created_at",
			"toppings.id", "toppings.type", "toppings.order_id",
		})

		for _, order := range orders {
			rows.AddRow(
				order.ID, order.Type, order.Name, order.Ppu, order.Status, order.Enable, order.OrderDate.String(), order.CreatedAt.String(),
				nil, nil, nil,
			)
			if len(order.Toppings) > 0 {
				for _, topping := range order.Toppings {
					rows.AddRow(
						order.ID, order.Type, order.Name, order.Ppu, order.Status, order.Enable, order.OrderDate.String(), order.CreatedAt.String(),
						topping.ID, topping.Type, order.ID,
					)
				}
			}
		}

		sql := `SELECT (.+) orders`
		dbmock.ExpectQuery(sql).WillReturnRows(rows)

		epOrders := runQuery(t, client, sql)
		assert.Equal(t, len(epOrders), len(orders))

		for index := range epOrders {
			assert.Equal(t, epOrders[index].ID.String(), orders[index].ID.String())
			assert.Equal(t, epOrders[index].Name, orders[index].Name)
			assert.Equal(t, epOrders[index].Ppu, orders[index].Ppu)
			assert.Equal(t, epOrders[index].Status, orders[index].Status)
			assert.Equal(t, epOrders[index].Type, orders[index].Type)
			assert.Equal(t, epOrders[index].Enable, orders[index].Enable)
			assert.Equal(t, epOrders[index].CreatedAt.String(), orders[index].CreatedAt.String())
			assert.Equal(t, epOrders[index].OrderDate.String(), orders[index].OrderDate.String())

			assert.NotNil(t, epOrders[index].ID)
			assert.NotZero(t, epOrders[index].Name)
			assert.NotZero(t, epOrders[index].Ppu)
			assert.NotZero(t, epOrders[index].Status)
			assert.NotZero(t, epOrders[index].Type)
			assert.NotZero(t, epOrders[index].Enable)
			assert.NotNil(t, epOrders[index].CreatedAt.String())
			assert.NotNil(t, epOrders[index].OrderDate.String())
			switch index {
			case 2:
				assert.Len(t, epOrders[index].Toppings, 0)
			default:
				assert.Greater(t, len(epOrders[index].Toppings), 0)
				assert.Equal(t, len(epOrders[index].Toppings), len(orders[index].Toppings))
				for toppingIndex := range orders[index].Toppings {
					assert.Equal(t, orders[index].Toppings[toppingIndex].ID, epOrders[index].Toppings[toppingIndex].ID)
					assert.Equal(t, orders[index].Toppings[toppingIndex].Type, epOrders[index].Toppings[toppingIndex].Type)
				}
			}
		}
	})

	t.Run("success_with_order_left_join_both_topping_batters", func(t *testing.T) {
		orders := getData(t, "../testdata/orders.json")
		db, dbmock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}

		sqlxDB := sqlx.NewDb(db, "sqlmock")
		defer db.Close()

		client := new(psql.Client)
		client.SetDB(sqlxDB)

		rows := sqlmock.NewRows([]string{
			"orders.id", "orders.type", "orders.name", "orders.ppu", "orders.status", "orders.enable", "orders.order_date", "orders.created_at",
			"toppings.id", "toppings.type", "toppings.order_id",
			"batters.id", "batters.type", "batters.order_id",
		})

		for _, order := range orders {
			rows.AddRow(
				order.ID, order.Type, order.Name, order.Ppu, order.Status, order.Enable, order.OrderDate.String(), order.CreatedAt.String(),
				nil, nil, nil,
				nil, nil, nil,
			)
			if len(order.Toppings) > 0 {
				for _, topping := range order.Toppings {
					rows.AddRow(
						order.ID, order.Type, order.Name, order.Ppu, order.Status, order.Enable, order.OrderDate.String(), order.CreatedAt.String(),
						topping.ID, topping.Type, order.ID,
						nil, nil, nil,
					)
				}
			}
			if len(order.Batters) > 0 {
				for _, item := range order.Batters {
					rows.AddRow(
						order.ID, order.Type, order.Name, order.Ppu, order.Status, order.Enable, order.OrderDate.String(), order.CreatedAt.String(),
						nil, nil, nil,
						item.ID, item.Type, order.ID,
					)
				}
			}
		}

		sql := `SELECT (.+) orders`
		dbmock.ExpectQuery(sql).WillReturnRows(rows)

		epOrders := runQuery(t, client, sql)
		assert.Equal(t, len(epOrders), len(orders))

		for index := range epOrders {
			assert.Equal(t, epOrders[index].ID.String(), orders[index].ID.String())
			assert.Equal(t, epOrders[index].Name, orders[index].Name)
			assert.Equal(t, epOrders[index].Ppu, orders[index].Ppu)
			assert.Equal(t, epOrders[index].Status, orders[index].Status)
			assert.Equal(t, epOrders[index].Type, orders[index].Type)
			assert.Equal(t, epOrders[index].Enable, orders[index].Enable)
			assert.Equal(t, epOrders[index].CreatedAt.String(), orders[index].CreatedAt.String())
			assert.Equal(t, epOrders[index].OrderDate.String(), orders[index].OrderDate.String())

			assert.NotNil(t, epOrders[index].ID)
			assert.NotZero(t, epOrders[index].Name)
			assert.NotZero(t, epOrders[index].Ppu)
			assert.NotZero(t, epOrders[index].Status)
			assert.NotZero(t, epOrders[index].Type)
			assert.NotZero(t, epOrders[index].Enable)
			assert.NotNil(t, epOrders[index].CreatedAt.String())
			assert.NotNil(t, epOrders[index].OrderDate.String())
			switch index {
			case 2:
				assert.Len(t, epOrders[index].Toppings, 0)
			default:
				assert.Greater(t, len(epOrders[index].Toppings), 0)
				assert.Equal(t, len(epOrders[index].Toppings), len(orders[index].Toppings))
				for toppingIndex := range orders[index].Toppings {
					assert.Equal(t, orders[index].Toppings[toppingIndex].ID, epOrders[index].Toppings[toppingIndex].ID)
					assert.Equal(t, orders[index].Toppings[toppingIndex].Type, epOrders[index].Toppings[toppingIndex].Type)
				}
				assert.Greater(t, len(epOrders[index].Batters), 0)
				assert.Equal(t, len(epOrders[index].Batters), len(orders[index].Batters))
				for batterIndex := range orders[index].Batters {
					assert.Equal(t, orders[index].Batters[batterIndex].ID, epOrders[index].Batters[batterIndex].ID)
					assert.Equal(t, orders[index].Batters[batterIndex].Type, epOrders[index].Batters[batterIndex].Type)
				}
			}
		}
	})
	t.Run("success_with_order_left_join_both_topping_batters_and_join_chef", func(t *testing.T) {
		orders := getData(t, "../testdata/orders.json")
		db, dbmock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}

		sqlxDB := sqlx.NewDb(db, "sqlmock")
		defer db.Close()

		client := new(psql.Client)
		client.SetDB(sqlxDB)

		rows := sqlmock.NewRows([]string{
			"orders.id", "orders.type", "orders.name", "orders.ppu", "orders.status", "orders.enable", "orders.order_date", "orders.created_at", "orders.chef_id",
			"toppings.id", "toppings.type", "toppings.order_id",
			"batters.id", "batters.type", "batters.order_id",
			"chefs.id", "chefs.name", "chefs.order_id",
		})

		for _, order := range orders {
			var chefId *uuid.UUID
			var chefName *string
			if order.Chef != nil {
				chefId = order.Chef.ID
				chefName = &order.Chef.Name
			}
			rows.AddRow(
				order.ID, order.Type, order.Name, order.Ppu, order.Status, order.Enable, order.OrderDate.String(), order.CreatedAt.String(), chefId,
				nil, nil, nil,
				nil, nil, nil,
				chefId, chefName, order.ID,
			)
			if len(order.Toppings) > 0 {
				for _, topping := range order.Toppings {
					rows.AddRow(
						order.ID, order.Type, order.Name, order.Ppu, order.Status, order.Enable, order.OrderDate.String(), order.CreatedAt.String(), chefId,
						topping.ID, topping.Type, order.ID,
						nil, nil, nil,
						chefId, chefName, order.ID,
					)
				}
			}
			if len(order.Batters) > 0 {
				for _, item := range order.Batters {
					rows.AddRow(
						order.ID, order.Type, order.Name, order.Ppu, order.Status, order.Enable, order.OrderDate.String(), order.CreatedAt.String(), chefId,
						nil, nil, nil,
						item.ID, item.Type, order.ID,
						chefId, chefName, order.ID,
					)
				}
			}
		}

		sql := `SELECT (.+) orders`
		dbmock.ExpectQuery(sql).WillReturnRows(rows)

		epOrders := runQuery(t, client, sql)
		assert.Equal(t, len(epOrders), len(orders))

		for index := range epOrders {
			assert.Equal(t, epOrders[index].ID.String(), orders[index].ID.String())
			assert.Equal(t, epOrders[index].Name, orders[index].Name)
			assert.Equal(t, epOrders[index].Ppu, orders[index].Ppu)
			assert.Equal(t, epOrders[index].Status, orders[index].Status)
			assert.Equal(t, epOrders[index].Type, orders[index].Type)
			assert.Equal(t, epOrders[index].Enable, orders[index].Enable)
			assert.Equal(t, epOrders[index].CreatedAt.String(), orders[index].CreatedAt.String())
			assert.Equal(t, epOrders[index].OrderDate.String(), orders[index].OrderDate.String())

			assert.NotNil(t, epOrders[index].ID)
			assert.NotZero(t, epOrders[index].Name)
			assert.NotZero(t, epOrders[index].Ppu)
			assert.NotZero(t, epOrders[index].Status)
			assert.NotZero(t, epOrders[index].Type)
			assert.NotZero(t, epOrders[index].Enable)
			assert.NotNil(t, epOrders[index].CreatedAt.String())
			assert.NotNil(t, epOrders[index].OrderDate.String())
			switch index {
			case 2:
				assert.Len(t, epOrders[index].Toppings, 0)
				assert.Nil(t, epOrders[index].Chef)
			default:
				assert.Greater(t, len(epOrders[index].Toppings), 0)
				assert.Equal(t, len(epOrders[index].Toppings), len(orders[index].Toppings))
				for toppingIndex := range orders[index].Toppings {
					assert.Equal(t, orders[index].Toppings[toppingIndex].ID, epOrders[index].Toppings[toppingIndex].ID)
					assert.Equal(t, orders[index].Toppings[toppingIndex].Type, epOrders[index].Toppings[toppingIndex].Type)
				}
				assert.Greater(t, len(epOrders[index].Batters), 0)
				assert.Equal(t, len(epOrders[index].Batters), len(orders[index].Batters))
				for batterIndex := range orders[index].Batters {
					assert.Equal(t, orders[index].Batters[batterIndex].ID, epOrders[index].Batters[batterIndex].ID)
					assert.Equal(t, orders[index].Batters[batterIndex].Type, epOrders[index].Batters[batterIndex].Type)
				}

				assert.Equal(t, epOrders[index].Chef.ID.String(), orders[index].Chef.ID.String())
				assert.Equal(t, epOrders[index].Chef.Name, orders[index].Chef.Name)
			}
		}
	})
	t.Run("success_with_join_chef", func(t *testing.T) {
		orders := getData(t, "../testdata/orders.json")
		db, dbmock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}

		sqlxDB := sqlx.NewDb(db, "sqlmock")
		defer db.Close()

		client := new(psql.Client)
		client.SetDB(sqlxDB)

		rows := sqlmock.NewRows([]string{
			"orders.id", "orders.type", "orders.name", "orders.ppu", "orders.status", "orders.enable", "orders.order_date", "orders.created_at", "orders.chef_id",
			"chefs.id", "chefs.name", "chefs.order_id",
		})

		for _, order := range orders {
			var chefId *uuid.UUID
			var chefName *string
			if order.Chef != nil {
				chefId = order.Chef.ID
				chefName = &order.Chef.Name
			}
			rows.AddRow(
				order.ID, order.Type, order.Name, order.Ppu, order.Status, order.Enable, order.OrderDate.String(), order.CreatedAt.String(), chefId,
				chefId, chefName, order.ID,
			)
		}

		sql := `SELECT (.+) orders`
		dbmock.ExpectQuery(sql).WillReturnRows(rows)

		epOrders := runQuery(t, client, sql)
		assert.Equal(t, len(epOrders), len(orders))

		for index := range epOrders {
			assert.Equal(t, epOrders[index].ID.String(), orders[index].ID.String())
			assert.Equal(t, epOrders[index].Name, orders[index].Name)
			assert.Equal(t, epOrders[index].Ppu, orders[index].Ppu)
			assert.Equal(t, epOrders[index].Status, orders[index].Status)
			assert.Equal(t, epOrders[index].Type, orders[index].Type)
			assert.Equal(t, epOrders[index].Enable, orders[index].Enable)
			assert.Equal(t, epOrders[index].CreatedAt.String(), orders[index].CreatedAt.String())
			assert.Equal(t, epOrders[index].OrderDate.String(), orders[index].OrderDate.String())

			assert.NotNil(t, epOrders[index].ID)
			assert.NotZero(t, epOrders[index].Name)
			assert.NotZero(t, epOrders[index].Ppu)
			assert.NotZero(t, epOrders[index].Status)
			assert.NotZero(t, epOrders[index].Type)
			assert.NotZero(t, epOrders[index].Enable)
			assert.NotNil(t, epOrders[index].CreatedAt.String())
			assert.NotNil(t, epOrders[index].OrderDate.String())
			switch index {
			case 2:
				assert.Nil(t, epOrders[index].Chef)
			default:

				assert.Equal(t, epOrders[index].Chef.ID.String(), orders[index].Chef.ID.String())
				assert.Equal(t, epOrders[index].Chef.Name, orders[index].Chef.Name)
			}
		}
	})
	t.Run("success_with_option_disabled_binding", func(t *testing.T) {
		orders := getData(t, "../testdata/orders.json")
		db, dbmock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}

		sqlxDB := sqlx.NewDb(db, "sqlmock")
		defer db.Close()

		client := new(psql.Client)
		client.SetDB(sqlxDB)

		rows := sqlmock.NewRows([]string{
			"orders.id", "orders.type", "orders.name", "orders.ppu", "orders.status", "orders.enable", "orders.order_date", "orders.created_at", "orders.chef_id",
			"toppings.id", "toppings.type", "toppings.order_id",
			"batters.id", "batters.type", "batters.order_id",
			"chefs.id", "chefs.name", "chefs.order_id",
		})

		for _, order := range orders {
			var chefId *uuid.UUID
			var chefName *string
			if order.Chef != nil {
				chefId = order.Chef.ID
				chefName = &order.Chef.Name
			}
			rows.AddRow(
				order.ID, order.Type, order.Name, order.Ppu, order.Status, order.Enable, order.OrderDate.String(), order.CreatedAt.String(), chefId,
				nil, nil, nil,
				nil, nil, nil,
				chefId, chefName, order.ID,
			)
			if len(order.Toppings) > 0 {
				for _, topping := range order.Toppings {
					rows.AddRow(
						order.ID, order.Type, order.Name, order.Ppu, order.Status, order.Enable, order.OrderDate.String(), order.CreatedAt.String(), chefId,
						topping.ID, topping.Type, order.ID,
						nil, nil, nil,
						chefId, chefName, order.ID,
					)
				}
			}
			if len(order.Batters) > 0 {
				for _, item := range order.Batters {
					rows.AddRow(
						order.ID, order.Type, order.Name, order.Ppu, order.Status, order.Enable, order.OrderDate.String(), order.CreatedAt.String(), chefId,
						nil, nil, nil,
						item.ID, item.Type, order.ID,
						chefId, chefName, order.ID,
					)
				}
			}
		}

		sql := `SELECT (.+) orders`
		dbmock.ExpectQuery(sql).WillReturnRows(rows)

		epOrders := runQuery(t, client, sql, orm.NewMapperOption().SetDisableBinding())
		assert.Equal(t, len(epOrders), len(orders))

		for index := range epOrders {
			assert.Equal(t, epOrders[index].ID.String(), orders[index].ID.String())
			assert.Equal(t, epOrders[index].Name, orders[index].Name)
			assert.Equal(t, epOrders[index].Ppu, orders[index].Ppu)
			assert.Equal(t, epOrders[index].Status, orders[index].Status)
			assert.Equal(t, epOrders[index].Type, orders[index].Type)
			assert.Equal(t, epOrders[index].Enable, orders[index].Enable)
			assert.Equal(t, epOrders[index].CreatedAt.String(), orders[index].CreatedAt.String())
			assert.Equal(t, epOrders[index].OrderDate.String(), orders[index].OrderDate.String())

			assert.NotNil(t, epOrders[index].ID)
			assert.NotZero(t, epOrders[index].Name)
			assert.NotZero(t, epOrders[index].Ppu)
			assert.NotZero(t, epOrders[index].Status)
			assert.NotZero(t, epOrders[index].Type)
			assert.NotZero(t, epOrders[index].Enable)
			assert.NotNil(t, epOrders[index].CreatedAt.String())
			assert.NotNil(t, epOrders[index].OrderDate.String())
			assert.Equal(t, len(epOrders[index].Toppings), 0)
		}
	})
}
