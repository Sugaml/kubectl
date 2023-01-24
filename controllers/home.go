package controllers

import (
	"context"

	"github.com/astaxie/beego"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type HomeController struct {
	beego.Controller
}

func (c *HomeController) Get() {
	c.TplName = "index.html"
	path := c.GetString("path")
	cont := c.GetString("context")
	c.Data["path"] = path
	c.Data["context"] = cont
	if path == "" {
		c.Data["pods"], c.Data["error"] = "", ""
		return
	}
	c.Data["pods"], c.Data["error"] = getPods(cont, path)
}

func getPods(cont string, path string) (string, error) {
	con, err := buildConfigFromContextFlags(cont, path)
	if err != nil {
		return "", err
	}
	cl, err := kubernetes.NewForConfig(con)
	if err != nil {
		return "", err
	}
	po, err := cl.CoreV1().Pods("").List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return "", err
	}
	html := ""
	for _, p := range po.Items {
		html += "<li><a href='/terminal?path=" + path + "&context=" + cont + "&namespace=" + p.Namespace + "&pod=" +
			p.Name + "&container='>" + p.Namespace + " / " + p.Name + "</a></li>"
	}
	return html, nil
}
