package tlc

import (
	"fmt"
	"os"
	"sync"
)

const distributedSelectorFactoryProperty = "tlc2.tool.distributed.selector.factory"

// BlockSelection is the coordinator's selection boundary. The built-in
// BlockSelector remains concrete; linked Go extensions may supply another policy.
type BlockSelection interface {
	GetBlocks(StateQueue, *DistributedWorkerSmartProxy) []*TLCStateMut
	SetMaxTXSize(int)
	GetAverageBlockCnt() int64
}

// BlockSelectorFactory selects a policy for the supplied coordinator.
type BlockSelectorFactory func(*TLCServer) BlockSelection

// BlockSelectorFactoryConstructor creates a fresh factory for each selector
// creation. Returned construction errors are reported before built-in fallback;
// panics and failures in the returned factory propagate to the caller.
type BlockSelectorFactoryConstructor func() (BlockSelectorFactory, error)

var blockSelectorFactories = struct {
	sync.RWMutex
	constructors map[string]BlockSelectorFactoryConstructor
}{constructors: make(map[string]BlockSelectorFactoryConstructor)}

// RegisterBlockSelectorFactory binds the original factory property to linked Go
// code. Names are exact, without trimming. A nil constructor removes a binding.
// Register before creating coordinators; callbacks run outside the registry lock.
func RegisterBlockSelectorFactory(name string, constructor BlockSelectorFactoryConstructor) {
	blockSelectorFactories.Lock()
	defer blockSelectorFactories.Unlock()
	if constructor == nil {
		delete(blockSelectorFactories.constructors, name)
	} else {
		blockSelectorFactories.constructors[name] = constructor
	}
}

func loadBlockSelectorFactory(name string) BlockSelectorFactory {
	blockSelectorFactories.RLock()
	constructor := blockSelectorFactories.constructors[name]
	blockSelectorFactories.RUnlock()
	if constructor == nil {
		// Selecting the upstream base factory explicitly means built-in policy.
		if name != "tlc2.tool.distributed.selector.BlockSelectorFactory" {
			fmt.Fprintf(os.Stderr, "Block selector factory %q is not registered\n", name)
		}
		return nil
	}
	factory, err := constructor()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot construct block selector factory %q: %v\n", name, err)
		return nil
	}
	return factory
}
