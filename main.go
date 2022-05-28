package main

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"main/workflow"
)

/*
	依赖关系：
		- 55 依赖于5; 88 依赖于8；99 依赖于 9
		- 4，5 依赖于 2; 8, 9 依赖于3
		- 2 依赖于 0; 3 依赖于 1
		- 100 依赖于 101； 200 依赖于 201; 300 依赖于 301

					0               1				101				201			301
					|               |				 |				 |			 |
					2               3               100             200         300
				  /   \           /   \
				 4     5         8     9
					   |         |     |
				       55        88    99
	执行关系：
		- 支持设定最大并发运行的任务数;
		- 只有当所有的节点者执行结束后，才任务总体任务结束;
		- 无关([0,1] [4,8] [4,9] [5,8] [5,9])的节点并行执行，相互不影响；
		- 兄弟节点（如[4, 5], [8, 9], 注意：88， 99 并非兄弟关系)并行执行，相互不影响；
		- 若父节点失败，子点将不再执行；若子节点失败，只会停止以此子节点为祖先的其它的节点执行，此节点的兄弟节点不受影响
			例如：
				3 执行失败,那么 8， 9， 88， 99 都不再执行, 因为 3 为他们的公共祖先；
				8 执行失败,那么 88 将不再执行，但是除 88 之外的其它节点均不受影响
*/
func main() {
	// 初始化图关系（仅记录节点间的直接依赖，对于环境的检查待优化，当前的实现有问题
	graph := workflow.NewGraph()
	graph.AddEdge(workflow.Edge{
		Current:  workflow.Node("2"), // 节点 2 依赖于节点 0
		Previous: workflow.Node("0"),
	})

	graph.AddEdge(workflow.Edge{
		Current:  workflow.Node("3"), // 节点 3 依赖于节点 1
		Previous: workflow.Node("1"),
	})
	graph.AddEdge(workflow.Edge{
		Current:  workflow.Node("4"), // 节点 4 依赖于节点 2
		Previous: workflow.Node("2"),
	})

	graph.AddEdge(workflow.Edge{
		Current:  workflow.Node("5"), // 节点 5 依赖于节点 2
		Previous: workflow.Node("2"),
	})
	graph.AddEdge(workflow.Edge{
		Current:  workflow.Node("8"), // 节点 8 依赖于节点 3
		Previous: workflow.Node("3"),
	})

	graph.AddEdge(workflow.Edge{
		Current:  workflow.Node("9"), // 节点 9 依赖于节点 3
		Previous: workflow.Node("3"),
	})
	graph.AddEdge(workflow.Edge{
		Current:  workflow.Node("55"), // 节点 55 依赖于节点 5
		Previous: workflow.Node("5"),
	})

	graph.AddEdge(workflow.Edge{
		Current:  workflow.Node("99"), // 节点 99 依赖于节点 9
		Previous: workflow.Node("9"),
	})
	graph.AddEdge(workflow.Edge{
		Current:  workflow.Node("88"), // 节点 88 依赖于节点 8
		Previous: workflow.Node("8"),
	})

	graph.AddEdge(workflow.Edge{
		Current:  workflow.Node("100"), // 节点 100 依赖于节点 101
		Previous: workflow.Node("101"),
	})

	graph.AddEdge(workflow.Edge{
		Current:  workflow.Node("200"), // 节点 200 依赖于节点 201
		Previous: workflow.Node("201"),
	})

	graph.AddEdge(workflow.Edge{
		Current:  workflow.Node("300"), // 节点 300 依赖于节点 301
		Previous: workflow.Node("301"),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Minute)
	defer cancel()

	// 模拟工作负载的执行
	var sum int32
	fn := func(ctx context.Context, node workflow.Node) error {
		time.Sleep(3 * time.Second)
		atomic.AddInt32(&sum, 1)

		fmt.Printf("sum is added by node: %v\r\n", node)
		return nil
	}

	// 触发并发任务数为 4 的执行
	if err := workflow.Execute(ctx, graph, fn, 4); err != nil {
		fmt.Printf("failed to exec graph, err: %v\r\n", err.Error())
	}

	fmt.Printf("successfull execte with result: sum = %v\r\n", sum)

	<-ctx.Done()
}
