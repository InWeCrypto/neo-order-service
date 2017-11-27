package model

import (
	"database/sql"
	"fmt"

	"github.com/dynamicgo/config"
	"github.com/dynamicgo/slf4go"
)

// DBModel .
type DBModel struct {
	db  *sql.DB
	cnf *config.Config
	slf4go.Logger
}

// NewDBModel .
func NewDBModel(conf *config.Config, db *sql.DB) *DBModel {
	return &DBModel{
		cnf:    conf,
		db:     db,
		Logger: slf4go.Get("neo-order-service-model"),
	}
}

// GetSQL .
func (model *DBModel) GetSQL(name string) string {
	if !model.cnf.Has(name) {
		panic(fmt.Sprintf("unknown sql %s", name))
	}

	return model.cnf.GetString(name, "xxx")
}

// Tx execute a tx
func (model *DBModel) Tx(proc func(tx *sql.Tx) error) (reterr error) {
	tx, err := model.db.Begin()

	if err != nil {
		return err
	}

	defer func() {
		if err := recover(); err != nil {
			reterr = err.(error)
		}
	}()

	if err := proc(tx); err != nil {
		if rollbackError := tx.Rollback(); rollbackError != nil {
			model.ErrorF("rollback err, %s", rollbackError)
		}

		return err
	}

	return tx.Commit()
}

// WalletModel model model
type WalletModel struct {
	*DBModel
}

// Wallet .
type Wallet struct {
	Address    string `json:"address"`
	UserID     string `json:"userid"`
	CreateTime string `json:"createTime"`
}

// Create create new model
func (model *WalletModel) Create(address string, userid string) error {

	if address == "" {
		return fmt.Errorf("address param can't be empty string")
	}

	if userid == "" {
		return fmt.Errorf("userid param can't be empty string")
	}

	query := model.GetSQL("nos.orm.wallet.create")

	return model.Tx(func(tx *sql.Tx) error {

		model.DebugF("create wallet sql :%s address:%s userid:%s", query, address, userid)

		_, err := tx.Exec(query, address, userid)

		if err != nil {
			return err
		}

		return nil

	})
}

// Delete delete model indicate by userid and address
func (model *WalletModel) Delete(address string, userid string) error {

	if address == "" {
		return fmt.Errorf("address param can't be empty string")
	}

	if userid == "" {
		return fmt.Errorf("userid param can't be empty string")
	}

	query := model.GetSQL("nos.orm.wallet.delete")

	return model.Tx(func(tx *sql.Tx) error {

		model.DebugF("delete model sql :%s address:%s userid:%s", query, address, userid)

		_, err := tx.Exec(query, address, userid)

		if err != nil {
			return err
		}

		return nil

	})
}

// GetByAddress get wallet by address
func (model *WalletModel) GetByAddress(address string) (find *Wallet, err error) {

	if address == "" {
		return nil, fmt.Errorf("address param can't be empty string")
	}

	query := model.GetSQL("nos.orm.wallet.getbyaddress")

	err = model.Tx(func(tx *sql.Tx) error {

		model.DebugF("get wallet sql :%s address:%s ", query, address)

		rows, err := tx.Query(query, address)

		if err != nil {
			return err
		}

		defer rows.Close()

		if rows.Next() {
			var wallet Wallet

			if err := rows.Scan(&wallet.Address, &wallet.UserID, &wallet.CreateTime); err != nil {
				return err
			}

			find = &wallet
		}

		return nil

	})

	return
}

// OrderModel neo order model
type OrderModel struct {
	*DBModel
}

// Order neo order object
type Order struct {
	Tx          string `json:"tx" form:"tx" binding:"required"`
	From        string `json:"from" form:"from" binding:"required"`
	To          string `json:"to" form:"to" binding:"required"`
	Asset       string `json:"asset" form:"asset" binding:"required"`
	Value       string `json:"value" form:"value" binding:"required"`
	CreateTime  string `json:"createTime" form:"createTime"`
	ConfirmTime string `json:"confirmTime" form:"createTime"`
}

// Create create new ordereapig
func (model *OrderModel) Create(order *Order) error {

	query := model.GetSQL("nos.orm.order.create")

	return model.Tx(func(tx *sql.Tx) error {

		model.DebugF("create order sql :%s order :%s", query, order.Tx)

		_, err := tx.Exec(query, order.Tx, order.From, order.To, order.Asset, order.Value)

		if err != nil {
			return err
		}

		return nil

	})
}

// Confirm confirm order
func (model *OrderModel) Confirm(txid string) error {

	if txid == "" {
		return fmt.Errorf("tx param can't be empty string")
	}

	query := model.GetSQL("nos.orm.order.confirm")

	return model.Tx(func(tx *sql.Tx) error {

		model.DebugF("confirm sql :%s tx :%s", query, txid)

		_, err := tx.Exec(query, txid)

		if err != nil {
			return err
		}

		return nil
	})
}

// Status .
func (model *OrderModel) Status(txid string) (ok bool, err error) {
	if txid == "" {
		return false, fmt.Errorf("txid param can't be empty string")
	}

	query := model.GetSQL("nos.orm.order.status")

	err = model.Tx(func(tx *sql.Tx) error {

		model.DebugF("status sql :%s tx :%s", query, txid)

		rows, err := tx.Query(query, txid)

		if err != nil {
			return err
		}

		defer rows.Close()

		ok = rows.NextResultSet()

		return nil

	})

	return
}

// Orders get orders
func (model *OrderModel) Orders(address string, asset string, page *Page) (orders []*Order, err error) {
	if address == "" {
		return nil, fmt.Errorf("address param can't be empty string")
	}

	query := model.GetSQL("nos.orm.order.list")

	err = model.Tx(func(tx *sql.Tx) error {

		model.DebugF("list sql :%s address:%s page: %d size: %d", query, address, page.Offset, page.Size)

		rows, err := tx.Query(query, address, asset, page.Offset*page.Size, page.Size)

		if err != nil {
			return err
		}

		defer rows.Close()

		for rows.Next() {

			var order Order

			var confirmTime sql.NullString

			err = rows.Scan(
				&order.Tx,
				&order.From,
				&order.To,
				&order.Asset,
				&order.Value,
				&order.CreateTime,
				&confirmTime)

			if err != nil {
				return err
			}

			order.ConfirmTime = confirmTime.String

			orders = append(orders, &order)
		}

		return nil
	})

	return
}

// Order get order by txid
func (model *OrderModel) Order(txid string) (find *Order, err error) {
	if txid == "" {
		return nil, fmt.Errorf("txid param can't be empty string")
	}

	query := model.GetSQL("nos.orm.order.get")

	err = model.Tx(func(tx *sql.Tx) error {

		model.DebugF("get order sql :%s txid:%s", query, txid)

		rows, err := tx.Query(query, txid)

		if err != nil {
			model.ErrorF("query txid %s err: %s", txid, err)
			return err
		}

		defer rows.Close()

		if rows.Next() {

			var order Order

			var confirmTime sql.NullString

			err = rows.Scan(
				&order.Tx,
				&order.From,
				&order.To,
				&order.Asset,
				&order.Value,
				&order.CreateTime,
				&confirmTime)

			if err != nil {
				model.ErrorF("row next txid %s err: %s", txid, err)
				return err
			}

			order.ConfirmTime = confirmTime.String

			find = &order

		}

		return nil
	})

	return
}

// Page .
type Page struct {
	Offset uint `json:"offset" binding:"required"` // page offset number
	Size   uint `json:"size" binding:"required"`   // page size
}
