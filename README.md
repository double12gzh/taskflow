# taskflow
根据节点间依赖关系实现任务流的控制

## 示例说明

```bash

/*
	依赖关系：
		- 55 依赖于5; 88 依赖于8；99 依赖于 9
		- 4，5 依赖于 2; 8, 9 依赖于3
		- 2 依赖于 0; 3 依赖于 1
		- 100 依赖于 101； 200 依赖于 201; 300 依赖于 301
				    0               1		    101		201   301
				    |               |		     |   	 |	   |
				    2               3       100   200   300
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
```
