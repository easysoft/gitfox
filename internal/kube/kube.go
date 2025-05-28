// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package kube

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type ClientConfig struct {
	QPS   float32
	Burst int
}

// New returns a kubernetes client.
// It tries first with in-cluster config, if it fails it will try with out-of-cluster config.
func New(cc *ClientConfig) (client kubernetes.Interface, err error) {
	client, err = NewInCluster(cc)
	if err == nil {
		return
	}

	dir, err := os.UserHomeDir()
	if err != nil {
		return
	}
	dir = filepath.Join(dir, ".kube", "config")

	client, err = NewFromConfig(cc, dir)
	if err != nil {
		return
	}

	return
}

// NewFromConfig returns a new out-of-cluster kubernetes client.
func NewFromConfig(cc *ClientConfig, path string) (client kubernetes.Interface, err error) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return
	}

	cc.apply(config)

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return
	}

	client = clientset

	return
}

// NewInCluster returns a new in-cluster kubernetes client.
func NewInCluster(cc *ClientConfig) (client kubernetes.Interface, err error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return
	}

	cc.apply(config)

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return
	}

	client = clientset

	return
}

func (cc *ClientConfig) apply(config *rest.Config) {
	if cc.QPS > 0.0 {
		config.QPS = cc.QPS // the default is rest.DefaultQPS which is 5.0
	}

	if cc.Burst > 0 {
		config.Burst = cc.Burst // the default is rest.DefaultBurst which is 10
	}
}
