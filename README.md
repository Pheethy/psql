> [!IMPORTANT]
> - สามารถ Join ได้เฉพาะตัว Root เท่านั้น
> - จำนวน Pagination ใช้ key `total_row`
> - การตั้งค่าใน struct เหมือนเดิมทุกอย่าง

```golang
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	helperModel "git.innovasive.co.th/backend/models"
	"git.innovasive.co.th/backend/psql"
	"git.innovasive.co.th/backend/psql/orm"
	"github.com/Pheethy/sqlx"
	"github.com/gofrs/uuid"
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

func main() {
	mockOrders := getMockData()
	/* เป็นการจำลอง สถานการณืว่าเชื่อมต่อ database แล้วเท่านั้น */
	client, dbmock := getPsqlClient()
	setMockingRows(mockOrders, dbmock)

	/* query ตาม ปกติ */
	/*
		SELECT
			"orders.id", "orders.type", "orders.name", "orders.ppu", "orders.status", "orders.enable", "orders.order_date", "orders.created_at", "orders.chef_id",
			"toppings.id", "toppings.type", "toppings.order_id",
			"batters.id", "batters.type", "batters.order_id",
			"chefs.id", "chefs.name", "chefs.order_id",
	*/
	sql := fmt.Sprintf(
		`
			SELECT
				%s,
				%s,
				%s,
				%s
			FROM
				orders
			LEFT JOIN
				chefs
			ON
				orders.chef_id = chefs.id
			LEFT JOIN
				batters
			ON
				orders.id = batters.order_id
			LEFT JOIN
				toppings
			ON
				orders.id = toppings.order_id
		`,
		orm.GetSelector(Order{}),
		orm.GetSelector(Topping{}),
		orm.GetSelector(Batter{}),
		orm.GetSelector(Chef{}),
	)
	rows, err := client.GetClient().Queryx(sql)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	/* เรียก ฟังก์ชันเดียว ข้างใน left join ให้แล้ว */
    mainModel := new(Order)
	mapper, err := orm.Orm(mainModel, rows, orm.NewMapperOption())
	if err != nil {
		panic(err)
	}

	orders := mapper.GetData().([]*Order)
	fmt.Println(orders)
}

```