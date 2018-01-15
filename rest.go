package orderservice

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dynamicgo/config"
	"github.com/dynamicgo/slf4go"
	"github.com/gin-gonic/gin"
	"github.com/go-xorm/xorm"
	"github.com/inwecrypto/neo-order-service/model"
	"github.com/inwecrypto/neodb"
)

// HTTPServer .
type HTTPServer struct {
	engine *gin.Engine
	slf4go.Logger
	laddr string
	db    *xorm.Engine
}

// NewHTTPServer .
func NewHTTPServer(cnf *config.Config) (*HTTPServer, error) {

	if !cnf.GetBool("order.debug", true) {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())

	username := cnf.GetString("order.neodb.username", "xxx")
	password := cnf.GetString("order.neodb.password", "xxx")
	port := cnf.GetString("order.neodb.port", "6543")
	host := cnf.GetString("order.neodb.host", "localhost")
	scheme := cnf.GetString("order.neodb.schema", "postgres")

	db, err := xorm.NewEngine(
		"postgres",
		fmt.Sprintf(
			"user=%v password=%v host=%v dbname=%v port=%v sslmode=disable",
			username, password, host, scheme, port,
		),
	)

	if err != nil {
		return nil, err
	}

	service := &HTTPServer{
		engine: engine,
		Logger: slf4go.Get("neo-order-service"),
		laddr:  cnf.GetString("order.laddr", ":8000"),
		db:     db,
	}

	service.makeRouters()

	return service, nil
}

// Run run http service
func (service *HTTPServer) Run() error {
	return service.engine.Run(service.laddr)
}

func (service *HTTPServer) makeRouters() {
	service.engine.POST("/wallet/:userid/:address", func(ctx *gin.Context) {

		if err := service.createWallet(ctx.Param("userid"), ctx.Param("address")); err != nil {
			service.ErrorF("create wallet error :%s", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	})

	service.engine.DELETE("/wallet/:userid/:address", func(ctx *gin.Context) {
		if err := service.deleteWallet(ctx.Param("address"), ctx.Param("userid")); err != nil {
			service.ErrorF("create wallet error :%s", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	})

	service.engine.POST("/order", func(ctx *gin.Context) {
		var order *Order

		if err := ctx.ShouldBindJSON(&order); err != nil {
			service.ErrorF("parse order error :%s", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := service.createOrder(order); err != nil {
			service.ErrorF("create order error :%s", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	})

	// service.engine.POST("/order/:tx", func(ctx *gin.Context) {
	// 	orderModel := model.OrderModel{DBModel: service.dbmodel}

	// 	if err := orderModel.Confirm(ctx.Param("tx")); err != nil {
	// 		service.ErrorF("confirm order error :%s", err)
	// 		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 		return
	// 	}
	// })

	service.engine.GET("/order/:tx", func(ctx *gin.Context) {
		if orders, err := service.getOrder(ctx.Param("tx")); err != nil {
			service.ErrorF("get orders error :%s", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusOK, orders)
		}
	})

	service.engine.GET("/orders/:address/:asset/:offset/:size", func(ctx *gin.Context) {
		offset, err := parseInt(ctx, "offset")

		if err != nil {
			service.ErrorF("parse page parameter error :%s", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		size, err := parseInt(ctx, "size")

		if err != nil {
			service.ErrorF("parse page parameter error :%s", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		orders, err := service.getPagedOrders(ctx.Param("address"), ctx.Param("asset"), offset, size)

		if err != nil {
			service.ErrorF("get paged orders error :%s", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, orders)
	})
}

func parseInt(ctx *gin.Context, name string) (int, error) {
	result, err := strconv.ParseInt(ctx.Param(name), 10, 32)

	return int(result), err
}

func (service *HTTPServer) createWallet(userid string, address string) error {

	wallet := &neodb.Wallet{
		Address: address,
		UserID:  userid,
	}

	_, err := service.db.Insert(wallet)
	return err
}

func (service *HTTPServer) deleteWallet(userid string, address string) error {

	wallet := &neodb.Wallet{
		Address: address,
		UserID:  userid,
	}

	_, err := service.db.Delete(wallet)
	return err
}

// Order neo order object
type Order struct {
	Tx          string  `json:"tx" form:"tx" binding:"required"`
	From        string  `json:"from" form:"from" binding:"required"`
	To          string  `json:"to" form:"to" binding:"required"`
	Asset       string  `json:"asset" form:"asset" binding:"required"`
	Value       string  `json:"value" form:"value" binding:"required"`
	CreateTime  string  `json:"createTime" form:"createTime"`
	ConfirmTime string  `json:"confirmTime" form:"confirmTime"`
	Context     *string `json:"context,omitempty" xorm:"json"`
}

func (service *HTTPServer) getPagedOrders(address, asset string, offset, size int) ([]*Order, error) {

	service.DebugF("get address(%s) orders(%s) (%d,%d)", address, asset, offset, size)

	torders := make([]*neodb.Order, 0)

	err := service.db.
		Where(`("from" = ? or "to" = ?) and asset = ?`, address, address, asset).
		Desc("create_time").
		Limit(size, offset).
		Find(&torders)

	if err != nil {
		return make([]*Order, 0), err
	}

	orders := make([]*Order, 0)

	for _, torder := range torders {
		createTime := torder.CreateTime.Format(time.RFC3339Nano)

		var confirmTime string

		if torder.ConfirmTime != nil {
			confirmTime = torder.ConfirmTime.Format(time.RFC3339Nano)
		}

		orders = append(orders, &Order{
			Tx:          torder.TX,
			From:        torder.From,
			To:          torder.To,
			Asset:       torder.Asset,
			Value:       torder.Value,
			Context:     torder.Context,
			CreateTime:  createTime,
			ConfirmTime: confirmTime,
		})
	}

	return orders, nil
}

func (service *HTTPServer) createOrder(order *Order) error {

	tOrder := &neodb.Order{
		TX:      order.Tx,
		From:    order.From,
		To:      order.To,
		Asset:   order.Asset,
		Value:   order.Value,
		Context: order.Context,
		Block:   -1,
	}

	_, err := service.db.Insert(tOrder)
	return err
}

func (service *HTTPServer) getOrder(tx string) ([]*Order, error) {

	service.DebugF("get order by tx %s", tx)

	torders := make([]*neodb.Order, 0)

	err := service.db.Where("t_x = ?", tx).Find(&torders)

	if err != nil {
		return make([]*Order, 0), err
	}

	orders := make([]*Order, 0)

	for _, torder := range torders {
		createTime := torder.CreateTime.Format(time.RFC3339Nano)

		var confirmTime string

		if torder.ConfirmTime != nil {
			confirmTime = torder.ConfirmTime.Format(time.RFC3339Nano)
		}

		orders = append(orders, &Order{
			Tx:          torder.TX,
			From:        torder.From,
			To:          torder.To,
			Asset:       torder.Asset,
			Value:       torder.Value,
			Context:     torder.Context,
			CreateTime:  createTime,
			ConfirmTime: confirmTime,
		})
	}

	return orders, nil
}

func parsePage(offset string, size string) (*model.Page, error) {
	if offset == "" {
		return nil, fmt.Errorf("offset parameter required")
	}

	if size == "" {
		return nil, fmt.Errorf("size parameter required")
	}

	offseti, err := strconv.ParseUint(offset, 10, 32)

	if err != nil {
		return nil, fmt.Errorf("offset parameter parse failed,%s", err)
	}

	sizei, err := strconv.ParseUint(size, 10, 32)

	if err != nil {
		return nil, fmt.Errorf("size parameter parse failed,%s", err)
	}

	return &model.Page{
		Offset: uint(offseti),
		Size:   uint(sizei),
	}, nil
}
