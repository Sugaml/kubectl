package main

import (
	"os"

	"github.com/astaxie/beego"
	_ "github.com/du2016/web-terminal-in-go/k8s-webshell/routers"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		logrus.Fatalf("error in load env file %v", err)
	} else {
		logrus.Info("Successfully loaded env file.")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}
	beego.Run(port)
}
