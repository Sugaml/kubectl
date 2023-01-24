package controllers

import (
	"github.com/astaxie/beego"
)

type TerminalController struct {
	beego.Controller
}

func (ctl *TerminalController) Get() {
	ctl.Data["path"] = ctl.GetString("path")
	ctl.Data["context"] = ctl.GetString("context")
	ctl.Data["namespace"] = ctl.GetString("namespace")
	ctl.Data["pod"] = ctl.GetString("pod")
	ctl.Data["container"] = ctl.GetString("container")
	ctl.TplName = "terminal.html"
}
