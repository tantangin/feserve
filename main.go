package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/cors"
	mlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/ryanbekhen/feserve/internal/config"
	"github.com/ryanbekhen/feserve/internal/handler"
	"github.com/ryanbekhen/feserve/internal/logger"
	mw "github.com/ryanbekhen/feserve/internal/middleware"
	"github.com/ryanbekhen/feserve/internal/router"
)

var conf *config.Config

func init() {
	conf = config.Load()
}

func main() {
	logger := logger.New(logger.Config{
		Timezone: conf.TimeZone,
	})

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ProxyHeader:           conf.ProxyHeader,
		CompressedFileSuffix:  ".feserve.gz",
		ErrorHandler:          handler.ErrorHandler,
	})

	app.Use(mw.CustomHeaderMiddleware)

	if conf.AllowOrigins != "" {
		app.Use(cors.New(cors.Config{
			AllowOrigins: conf.AllowOrigins,
		}))
	}

	app.Use(cache.New(cache.Config{
		ExpirationGenerator: func(c *fiber.Ctx, cfg *cache.Config) time.Duration {
			cachetime, _ := strconv.Atoi(c.GetRespHeader("Cache-Time", "0"))
			return time.Second * time.Duration(cachetime)
		},
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.Path()
		},
	}))

	app.Use(mlogger.New(mlogger.Config{
		TimeZone:   conf.TimeZone,
		TimeFormat: time.RFC3339,
		Format:     "[${time}] - ${ip} - ${status} ${method} ${url} ${ua}\n",
	}))

	app.Static("/", conf.PublicDir, fiber.Static{
		Compress: true,
		ModifyResponse: func(c *fiber.Ctx) error {
			c.Response().Header.Add("Cache-Time", "86400")
			return nil
		},
	})

	router.Builder(app)

	addr := fmt.Sprintf("%s:%s", conf.Host, conf.Port)

	logger.Info("app listen on ", addr)
	if err := app.Listen(addr); err != nil {
		logger.Info(err.Error())
		os.Exit(1)
	}
}
