package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

// WorkFuncWithContext 执行用户自定义的业务逻辑
type WorkFuncWithContext func(context.Context, Node) error

// Execute 执行
func Execute(ctx context.Context, g *Graph, fn WorkFuncWithContext, taskConcurrency int) error {
	s := &scheduler{
		ch: make(chan struct{}, taskConcurrency),
		wg: make(map[Node]schMeta),
	}

	for nodeName := range g.Nodes {
		s.init(nodeName)
	}

	var eg errgroup.Group
	for n := range g.Nodes {
		n := n

		eg.Go(func() error {
			c, cancel := context.WithTimeout(ctx, time.Minute)
			defer cancel()

			err := s.execute(c, n, g.GetPrevious(n), fn)

			return err
		})
	}

	return eg.Wait()
}

type schMeta struct {
	err error
	wg  *sync.WaitGroup
}

// scheduler 并发调度执行
type scheduler struct {
	ch chan struct{}
	wg map[Node]schMeta
}

func (s *scheduler) init(n Node) {
	s.wg[n] = schMeta{wg: &sync.WaitGroup{}, err: nil}
	s.wg[n].wg.Add(1)
}

func (s *scheduler) execute(ctx context.Context, n Node, dependency []Node, fn WorkFuncWithContext) (err error) {
	log.Infof("start handling node: %v, dep: %v", n, dependency)
	defer func() {
		log.Infof("finish handling node: %+v, dep: %+v, msg: %+v, time: %+v", n, dependency, err, time.Now().String())

		// 任务执行完毕后释放
		<-s.ch
		s.wg[n].wg.Done()
	}()

	failures := make([]string, 0)
	for _, dep := range dependency {
		s.wg[dep].wg.Wait()

		// 被依赖节点有失败时，所有依赖于此节点的后续节点的操作不再执行
		if s.wg[dep].err != nil {
			failures = append(failures, fmt.Sprintf("node: %v, err: %s", dep, s.wg[dep].err.Error()))
		}
	}

	// 统计是否有失败的情况，若存在则直接退出依赖(直接或间接，即以此节点为根的所有节点)此节点的其它任务。注意：不依赖此节点的任务不受影响
	if len(failures) > 0 {
		return fmt.Errorf("skipped node: %v. this is triggered by failure depencies, err: %+v", n, failures)
	}

	// 获取任务执行权限
	s.ch <- struct{}{}

	if err = fn(ctx, n); err != nil {
		if schMetadata, ok := s.wg[n]; ok {
			schMetadata.err = fmt.Errorf("work func failed, err: %s", err.Error())
		}
	}

	return err
}
