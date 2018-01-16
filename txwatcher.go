package orderservice

import (
	"fmt"
	"time"

	"github.com/denverdino/aliyungo/push"
	"github.com/dynamicgo/config"
	"github.com/dynamicgo/slf4go"
	"github.com/go-xorm/xorm"
	"github.com/inwecrypto/gomq"
	kafka "github.com/inwecrypto/gomq-kafka"
	"github.com/inwecrypto/neodb"
)

const (
	neoAsset = "0xc56f33fc6ecfcd0c225c4ab356fee59390af8560be0e930faebe74a6daff7c9b"
	gasAsset = "0x602c79718b16e442de58778e148d0b1084e3b2dffd5de6b7b16cee7969282de7"
)

var assetNames = map[string]string{
	neoAsset: "NEO",
	gasAsset: "NEO GAS",
}

func assetName(id string) string {
	name, ok := assetNames[id]

	if !ok {
		name = "unknown asset"
	}

	return name
}

type pushMessage struct {
	message string
	id      string
}

// TxWatcher tx event watcher
type TxWatcher struct {
	mq gomq.Consumer
	db *xorm.Engine
	slf4go.Logger
	pushClient   *push.Client
	appkey       int64
	pushChan     chan *pushMessage
	pushDuration time.Duration
}

// NewTxWatcher .
func NewTxWatcher(conf *config.Config) (*TxWatcher, error) {

	mq, err := kafka.NewAliyunConsumer(conf)

	if err != nil {
		return nil, err
	}

	username := conf.GetString("order.neodb.username", "xxx")
	password := conf.GetString("order.neodb.password", "xxx")
	port := conf.GetString("order.neodb.port", "6543")
	host := conf.GetString("order.neodb.host", "localhost")
	scheme := conf.GetString("order.neodb.schema", "postgres")

	db, err := xorm.NewEngine(
		"postgres",
		fmt.Sprintf(
			"user=%v password=%v host=%v dbname=%v port=%v sslmode=disable",
			username, password, host, scheme, port,
		),
	)
	client := push.NewClient(
		conf.GetString("nos.push.user", "xxxx"),
		conf.GetString("nos.push.password", "xxxxx"),
	)

	return &TxWatcher{
		mq:           mq,
		db:           db,
		Logger:       slf4go.Get("txwatcher"),
		pushClient:   client,
		appkey:       conf.GetInt64("nos.push.appkey", 0),
		pushChan:     make(chan *pushMessage, 100),
		pushDuration: conf.GetDuration("nos.push.duration", time.Second*2),
	}, nil
}

// Run run watcher
func (watcher *TxWatcher) Run() {

	for {
		select {
		case message, ok := <-watcher.mq.Messages():
			if ok {
				if err := watcher.confirm(string(message.Key())); err != nil {
					watcher.ErrorF("process tx confirm error,%s", err)
				}

				watcher.mq.Commit(message)
			}
		case err, ok := <-watcher.mq.Errors():
			if ok {
				watcher.ErrorF("kfka tx event mq err, %s", err)
			}
		}
	}
}

func (watcher *TxWatcher) confirm(txid string) error {
	watcher.DebugF("handle tx %s", txid)

	neoTxs := make([]*neodb.Tx, 0)

	err := watcher.db.Where("t_x = ?", txid).Find(&neoTxs)

	if err != nil {
		return err
	}

	if len(neoTxs) == 0 {
		watcher.WarnF("handle tx %s -- not found", txid)
		return nil
	}

	order := new(neodb.Order)

	order.ConfirmTime = &neoTxs[0].CreateTime
	order.Block = int64(neoTxs[0].Block)

	updated, err := watcher.db.Where("t_x = ?", txid).Cols("confirm_time", "block").Update(order)

	if err != nil {
		return err
	}

	if updated != 0 {
		watcher.DebugF("updated orders(%d) for tx %s", updated, txid)
		return nil
	}

	var orders []*neodb.Order
	wallet := new(neodb.Wallet)

	for _, tx := range neoTxs {

		count, err := watcher.db.Where(`"address" = ? or "address" = ?`, tx.From, tx.To).Count(wallet)

		if err != nil {
			return err
		}

		if count > 0 {

			order := new(neodb.Order)

			order.Asset = tx.Asset
			order.From = tx.From
			order.To = tx.To
			order.TX = tx.TX
			order.Value = tx.Value
			order.CreateTime = tx.CreateTime
			order.ConfirmTime = &tx.CreateTime
			order.Block = int64(tx.Block)
			orders = append(orders, order)
		}
	}

	if len(orders) > 0 {
		_, err = watcher.db.Insert(&orders)

		return err
	}

	return nil
}
