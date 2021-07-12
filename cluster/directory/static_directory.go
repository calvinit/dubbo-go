/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package directory

import (
	perrors "github.com/pkg/errors"
)

import (
	"github.com/apache/dubbo-go/cluster/router/chain"
	"github.com/apache/dubbo-go/common"
	"github.com/apache/dubbo-go/protocol"
)

type staticDirectory struct {
	BaseDirectory
	invokers []protocol.Invoker
	// healthState
	serviceHealthState *protocol.ServiceHealthState
}

// NewStaticDirectory Create a new staticDirectory with invokers
func NewStaticDirectory(invokers []protocol.Invoker) *staticDirectory {
	var url *common.URL

	if len(invokers) > 0 {
		url = invokers[0].GetURL()
	}
	dir := &staticDirectory{
		BaseDirectory: NewBaseDirectory(url),
		invokers:      invokers,
		// init serviceHealthState
		serviceHealthState: protocol.NewServiceState(url.ServiceKey()),
	}

	return dir
}

//for-loop invokers ,if all invokers is available ,then it means directory is available
func (dir *staticDirectory) IsAvailable() bool {
	if len(dir.invokers) == 0 {
		return false
	}
	for _, invoker := range dir.invokers {
		if !invoker.IsAvailable() {
			return false
		}
	}
	return true
}

// List List invokers
func (dir *staticDirectory) List(invocation protocol.Invocation) []protocol.Invoker {
	l := len(dir.invokers)
	invokers := make([]protocol.Invoker, l)
	copy(invokers, dir.invokers)
	routerChain := dir.RouterChain()

	if routerChain == nil {
		return invokers
	}
	dirUrl := dir.GetURL()
	return routerChain.Route(dirUrl, invocation)
}

// Fetch ServiceHealthState
func (dir *staticDirectory) ServiceHealthState() *protocol.ServiceHealthState {
	return dir.serviceHealthState
}

// Destroy Destroy
func (dir *staticDirectory) Destroy() {
	dir.BaseDirectory.Destroy(func() {
		for _, ivk := range dir.invokers {
			ivk.Destroy()
		}
		dir.invokers = []protocol.Invoker{}
	})
}

// BuildRouterChain build router chain by invokers
func (dir *staticDirectory) BuildRouterChain(invokers []protocol.Invoker) error {
	if len(invokers) == 0 {
		return perrors.Errorf("invokers == null")
	}
	url := invokers[0].GetURL()
	routerChain, e := chain.NewRouterChain(url)
	if e != nil {
		return e
	}
	routerChain.SetInvokers(dir.invokers)
	dir.SetRouterChain(routerChain)
	return nil
}
