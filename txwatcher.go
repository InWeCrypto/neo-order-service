package orderservice

import (
	"fmt"

	"github.com/denverdino/aliyungo/push"
	"github.com/dynamicgo/config"
	"github.com/dynamicgo/slf4go"
	"github.com/inwecrypto/gomq"
	kafka "github.com/inwecrypto/gomq-kafka"
	"github.com/inwecrypto/neo-order-service/model"
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

// TxWatcher tx event watcher
type TxWatcher struct {
	mq      gomq.Consumer
	dbmodel *model.DBModel
	slf4go.Logger
	pushClient *push.Client
	appkey     int64
}

// NewTxWatcher .
func NewTxWatcher(conf *config.Config) (*TxWatcher, error) {

	mq, err := kafka.NewAliyunConsumer(conf)

	if err != nil {
		return nil, err
	}

	db, err := openDB(conf)

	if err != nil {
		return nil, err
	}

	client := push.NewClient(
		conf.GetString("nos.push.user", "xxxx"),
		conf.GetString("nos.push.password", "xxxxx"),
	)

	return &TxWatcher{
		mq:         mq,
		dbmodel:    model.NewDBModel(conf, db),
		Logger:     slf4go.Get("txwatcher"),
		pushClient: client,
		appkey:     conf.GetInt64("nos.push.appkey", 0),
	}, nil
}

// Run run watcher
func (watcher *TxWatcher) Run() {
	for {
		select {
		case message, ok := <-watcher.mq.Messages():
			if ok {
				if err := watcher.confirm(string(message.Key())); err != nil {
					watcher.ErrorF("process tx confirm error")
					continue
				}

				go watcher.notify(string(message.Key()))

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
	orderModel := model.OrderModel{DBModel: watcher.dbmodel}

	return orderModel.Confirm(txid)
}

func (watcher *TxWatcher) notify(txid string) {
	orderModel := model.OrderModel{DBModel: watcher.dbmodel}
	walletModel := model.WalletModel{DBModel: watcher.dbmodel}

	order, err := orderModel.Order(txid)

	if err != nil {
		watcher.ErrorF("[message push] get order err :%s", err)
		return
	}

	if order == nil {
		watcher.DebugF("[message push] unknown order :%s", txid)
		return
	}

	from, err := walletModel.GetByAddress(order.From)

	if err != nil {
		watcher.ErrorF("[message push] get from wallet %s err :%s", order.From, err)
		return
	}

	to, err := walletModel.GetByAddress(order.To)

	if err != nil {
		watcher.ErrorF("[message push] get to wallet %s err :%s", order.To, err)
		return
	}

	if from != nil {
		message := fmt.Sprintf("%s转出成功：%s", assetName(order.Asset), order.Value)
		watcher.pushMessage(message, from.UserID)
	}

	if to != nil {
		message := fmt.Sprintf("%s转入成功：%s", assetName(order.Asset), order.Value)
		watcher.pushMessage(message, from.UserID)
	}

	if err != nil {
		watcher.ErrorF("message push error :%s", err)
	}
}

func (watcher *TxWatcher) pushMessage(message string, target string) {
	pushArgs := &push.PushArgs{
		AppKey:      watcher.appkey,
		Target:      push.PushTargetAll,
		TargetValue: target,
		PushType:    push.PushTypeMessage,
		DeviceType:  push.PushDeviceTypeAll,
		Title:       "转账通知",
		Body:        message,
		Summary:     message,
	}

	_, err := watcher.pushClient.Push(pushArgs)

	if err != nil {
		watcher.ErrorF("push message to %s err, %s", target, err)
	} else {
		watcher.DebugF("push message to %s : %s", target, message)
	}
}
