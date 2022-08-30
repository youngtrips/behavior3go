/*
从原生工程文件加载
*/
package main

import (
	"fmt"
	b3 "github.com/youngtrips/behavior3go"
	. "github.com/youngtrips/behavior3go/config"
	. "github.com/youngtrips/behavior3go/core"
	. "github.com/youngtrips/behavior3go/examples/share"
	. "github.com/youngtrips/behavior3go/loader"
)

func main() {
	projectConfig, ok := LoadRawProjectCfg("example.b3")
	if !ok {
		fmt.Println("LoadRawProjectCfg err")
		return
	}

	//自定义节点注册
	maps := b3.NewRegisterStructMaps()
	maps.Register("Log", new(LogTest))

	var firstTree *BehaviorTree
	//载入
	for _, v := range projectConfig.Data.Trees {
		tree := CreateBevTreeFromConfig(&v, maps)
		tree.Print()
		if firstTree == nil {
			firstTree = tree
		}
	}

	//输入板
	board := NewBlackboard()
	//循环每一帧
	for i := 0; i < 5; i++ {
		firstTree.Tick(i, board)
	}
}
