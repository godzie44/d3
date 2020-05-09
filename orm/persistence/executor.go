package persistence

type executor struct {
	storage     Pusher
	afterExecFn func(act CompositeAction)
}

func NewExecutor(storage Pusher, afterExec func(act CompositeAction)) *executor {
	return &executor{storage, afterExec}
}

func (e *executor) Exec(graph *PersistGraph) error {
	nodes := graph.filterRoots()

	var allNodesExecuted bool
	for !allNodesExecuted {
		allNodesExecuted = true
		for _, startNode := range nodes {
			if execRes, err := e.execNode(startNode); err != nil {
				return err
			} else {
				allNodesExecuted = execRes && allNodesExecuted
			}
		}
	}

	return nil
}

func (e *executor) execNode(node CompositeAction) (bool, error) {
	var result bool
	if result = node.subscriptionsResolved(); result && !node.wasExecuted() {
		if err := node.exec(e.storage); err != nil {
			return false, err
		}
		e.afterExecFn(node)
	}

	for _, childNode := range node.children() {
		if execRes, err := e.execNode(childNode); err != nil {
			return false, err
		} else {
			result = result && execRes
		}
	}

	return result, nil
}
