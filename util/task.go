package util

import (
	"context"
	"sync"
)

type TaskManager struct {
	nodes []TaskNode
}

func NewTaskManager(nodes []TaskNode) *TaskManager {
	return &TaskManager{nodes: nodes}
}

func (t *TaskManager) start(ctx context.Context) (err error) {
	for _, node := range t.nodes {
		err = node.run(ctx)
	}
	return err
}

type TaskNode interface {
	run(ctx context.Context) (err error)
}

type NormalTaskNode struct {
	name string
	fun  func(ctx context.Context) (err error)
}

func NewNormalTaskNode(name string, fun func(ctx context.Context) (err error)) *NormalTaskNode {
	return &NormalTaskNode{name: name, fun: fun}
}

func (n *NormalTaskNode) run(ctx context.Context) (err error) {
	return n.fun(ctx)
}

type VirtualTaskNode struct {
	name  string
	nodes []NormalTaskNode
}

func NewVirtualTaskNode(name string, nodes []NormalTaskNode) *VirtualTaskNode {
	return &VirtualTaskNode{name: name, nodes: nodes}
}

func (v *VirtualTaskNode) run(ctx context.Context) (err error) {
	var wg sync.WaitGroup
	for _, node := range v.nodes {
		wg.Add(1)
		go func(node NormalTaskNode) {
			defer wg.Done()
			err = node.run(ctx)
		}(node)
	}
	wg.Wait()
	return
}
