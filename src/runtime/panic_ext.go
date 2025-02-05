/**
 * Copyright 2025 ByteDance Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package runtime

import "unsafe"

var maxFatalReoverTracebackDepth int = 128

//go:nosplit
func canGoExit(gp *g, pc, sp uintptr) (ok bool) {
	if fatalHandler == nil {
		return false
	}
	// var ss *stkframe
	systemstack(func() {
		gentraceback(pc, sp, 0, gp, 0, nil, maxFatalReoverTracebackDepth, func(s *stkframe, p unsafe.Pointer) bool {
			if fatalHandler.CheckRecovery(gp.goid, s.fn._Func(), s.pc) {
				ok = true
				return false
			}
			return true
		}, nil, 0)
	})
	// if ok {
	// 	runDeferOnce(gp, ss)
	// }
	return
}

// FatalHandler is a hook used to check if a goroutine with specific stack-frame can be exit alone,
// instead of fatal the whole process
//
// WARNING: its object must be concurrent-safe, since it may be execute on multiple threads.
type FatalHandler interface {
	// CheckRecovery checks each stack-frame of a goroutine,
	// tells if the goroutine can be exit alone.
	//
	// WARNING: this func will be called on systems-stack instead of the panic goroutine.
	// and it must be mark as go:nosplit
	CheckRecovery(gid uint64, fn *Func, pc uintptr) (recover bool)
}

var fatalHandler FatalHandler

// RegisterFatalHandler registers a FatalHandler for the whole process
func RegisterFatalHandler(h FatalHandler) {
	fatalHandler = h
}
