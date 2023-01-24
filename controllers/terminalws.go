package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/astaxie/beego"
	"gopkg.in/igm/sockjs-go.v2/sockjs"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

func (self TerminalSockjs) Read(p []byte) (int, error) {
	var reply string
	var msg map[string]uint16
	reply, err := self.conn.Recv()
	if err != nil {
		return 0, err
	}
	if err := json.Unmarshal([]byte(reply), &msg); err != nil {
		return copy(p, reply), nil
	} else {
		self.sizeChan <- &remotecommand.TerminalSize{
			msg["cols"],
			msg["rows"],
		}
		return 0, nil
	}
}

func (self TerminalSockjs) Write(p []byte) (int, error) {
	err := self.conn.Send(string(p))
	return len(p), err
}

type TerminalSockjs struct {
	conn      sockjs.Session
	sizeChan  chan *remotecommand.TerminalSize
	path      string
	context   string
	namespace string
	pod       string
	container string
}

func (self *TerminalSockjs) Next() *remotecommand.TerminalSize {
	size := <-self.sizeChan
	beego.Debug(fmt.Sprintf("terminal size to width: %d height: %d", size.Width, size.Height))
	return size
}

func buildConfigFromContextFlags(context, kubeconfigPath string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		}).ClientConfig()
}

func Handler(t *TerminalSockjs, cmd string) error {
	config, err := buildConfigFromContextFlags(t.context, t.path)
	if err != nil {
		return err
	}
	groupversion := schema.GroupVersion{
		Group:   "",
		Version: "v1",
	}
	config.GroupVersion = &groupversion
	config.APIPath = "/api"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = scheme.Codecs
	restclient, err := rest.RESTClientFor(config)
	if err != nil {
		return err
	}
	req := restclient.Post().
		Resource("pods").
		Name(t.pod).
		Namespace(t.namespace).
		SubResource("exec").
		Param("container", t.container).
		Param("stdin", "true").
		Param("stdout", "true").
		Param("stderr", "true").
		Param("command", cmd).Param("tty", "true")
	req.VersionedParams(
		&v1.PodExecOptions{
			Container: t.container,
			Command:   []string{},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		},
		scheme.ParameterCodec,
	)
	executor, err := remotecommand.NewSPDYExecutor(
		config, http.MethodPost, req.URL(),
	)
	if err != nil {
		return err
	}
	return executor.Stream(remotecommand.StreamOptions{
		Stdin:             t,
		Stdout:            t,
		Stderr:            t,
		Tty:               true,
		TerminalSizeQueue: t,
	})
}

func (self TerminalSockjs) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.FormValue("path")
	context := r.FormValue("context")
	namespace := r.FormValue("namespace")
	pod := r.FormValue("pod")
	container := r.FormValue("container")
	Sockjshandler := func(session sockjs.Session) {
		t := &TerminalSockjs{session, make(chan *remotecommand.TerminalSize),
			path, context, namespace, pod, container}
		if err := Handler(t, "/bin/sh"); err != nil {
			fmt.Println("got error", err)
			beego.Error(err)
			beego.Error(Handler(t, "/bin/bash"))
		}
	}
	sockjs.NewHandler("/terminal/ws", sockjs.DefaultOptions, Sockjshandler).ServeHTTP(w, r)
}

func LogsHandler(t *TerminalSockjs, cmd string) error {
	config, err := buildConfigFromContextFlags(t.context, t.path)
	if err != nil {
		return err
	}
	groupversion := schema.GroupVersion{
		Group:   "",
		Version: "v1",
	}
	config.GroupVersion = &groupversion
	config.APIPath = "/api"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = scheme.Codecs
	restclient, err := rest.RESTClientFor(config)
	if err != nil {
		return err
	}
	req := restclient.Get().
		Resource("pods").
		Name(t.pod).
		Namespace(t.namespace).
		SubResource("logs").
		Param("container", t.container).
		Param("stdin", "true").
		Param("stdout", "true").
		Param("stderr", "true").
		Param("command", cmd).Param("tty", "true")
	req.VersionedParams(
		&v1.PodLogOptions{
			Container: t.container,
			Follow:    true,
		},
		scheme.ParameterCodec,
	)
	executor, err := remotecommand.NewSPDYExecutor(
		config, http.MethodPost, req.URL(),
	)
	if err != nil {
		return err
	}
	return executor.Stream(remotecommand.StreamOptions{
		Stdin:             t,
		Stdout:            t,
		Stderr:            t,
		Tty:               true,
		TerminalSizeQueue: t,
	})
}
