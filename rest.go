package orderservice

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dynamicgo/config"
	"github.com/dynamicgo/slf4go"
	"github.com/gin-gonic/gin"
	"github.com/inwecrypto/neo-order-service/model"
)

// HTTPServer .
type HTTPServer struct {
	engine *gin.Engine
	slf4go.Logger
	laddr   string
	dbmodel *model.DBModel
}

// NewHTTPServer .
func NewHTTPServer(cnf *config.Config) (*HTTPServer, error) {

	if !cnf.GetBool("nos.debug", true) {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.Default()

	db, err := openDB(cnf)

	if err != nil {
		return nil, err
	}

	service := &HTTPServer{
		engine:  engine,
		Logger:  slf4go.Get("neo-order-service"),
		laddr:   cnf.GetString("nos.laddr", ":8000"),
		dbmodel: model.NewDBModel(cnf, db),
	}

	service.makeRouters()

	return service, nil
}

func openDB(cnf *config.Config) (*sql.DB, error) {
	driver := cnf.GetString("nos.database.driver", "xxxx")
	username := cnf.GetString("nos.database.username", "xxx")
	password := cnf.GetString("nos.database.password", "xxx")
	port := cnf.GetString("nos.database.port", "6543")
	host := cnf.GetString("nos.database.host", "localhost")
	schema := cnf.GetString("nos.database.schema", "postgres")
	maxconn := cnf.GetInt64("nos.database.maxconn", 10)

	db, err := sql.Open(driver, fmt.Sprintf("user=%v password=%v host=%v dbname=%v port=%v sslmode=disable", username, password, host, schema, port))

	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(int(maxconn))

	return db, nil
}

// Run run http service
func (service *HTTPServer) Run() error {
	return service.engine.Run(service.laddr)
}

func (service *HTTPServer) makeRouters() {
	service.engine.POST("/wallet/:userid/:address", func(ctx *gin.Context) {
		walletModel := model.WalletModel{DBModel: service.dbmodel}

		if err := walletModel.Create(ctx.Param("address"), ctx.Param("userid")); err != nil {
			service.ErrorF("create wallet error :%s", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	})

	service.engine.DELETE("/wallet/:userid/:address", func(ctx *gin.Context) {
		walletModel := model.WalletModel{DBModel: service.dbmodel}

		if err := walletModel.Delete(ctx.Param("address"), ctx.Param("userid")); err != nil {
			service.ErrorF("delete wallet error :%s", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	})

	service.engine.POST("/order", func(ctx *gin.Context) {
		orderModel := model.OrderModel{DBModel: service.dbmodel}

		var order model.Order

		if err := ctx.ShouldBindJSON(&order); err != nil {
			service.ErrorF("create order error :%s", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := orderModel.Create(&order); err != nil {
			service.ErrorF("create order error :%s", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	})

	service.engine.POST("/order/:tx", func(ctx *gin.Context) {
		orderModel := model.OrderModel{DBModel: service.dbmodel}

		if err := orderModel.Confirm(ctx.Param("tx")); err != nil {
			service.ErrorF("confirm order error :%s", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	})

	service.engine.GET("/order/:tx", func(ctx *gin.Context) {
		orderModel := model.OrderModel{DBModel: service.dbmodel}

		ok, err := orderModel.Status(ctx.Param("tx"))

		if err != nil {
			service.ErrorF("get order status error :%s", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, &struct{ Status bool }{
			Status: ok,
		})
	})

	service.engine.GET("/orders/:address/:asset/:offset/:size", func(ctx *gin.Context) {
		orderModel := model.OrderModel{DBModel: service.dbmodel}

		service.DebugF("%v", ctx.Params)

		page, err := parsePage(ctx.Param("offset"), ctx.Param("size"))

		if err != nil {
			service.ErrorF("parse page parameter error :%s", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		orders, err := orderModel.Orders(ctx.Param("address"), ctx.Param("asset"), page)

		if err != nil {
			service.ErrorF("order list error :%s", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, orders)
	})
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
