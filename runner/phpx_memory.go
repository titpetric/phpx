package runner

import (
	goruntime "runtime"
	gometrics "runtime/metrics"
)

// This file is the fork's memory accounting: the metric samples and the
// working set behind the real-usage read and the walk's reuse. The methods
// memory_get_usage() answers sit on Runtime in runtime.go.

// memVisitKeep bounds the visited set a walk keeps between walks.
const memVisitKeep = 1 << 12

// realUsageMetricNames are the heap classes that add up to HeapSys.
var realUsageMetricNames = [...]string{
	"/memory/classes/heap/objects:bytes",
	"/memory/classes/heap/unused:bytes",
	"/memory/classes/heap/free:bytes",
	"/memory/classes/heap/released:bytes",
}

// realUsageSamples holds the samples one real-usage read fills.
type realUsageSamples [len(realUsageMetricNames)]gometrics.Sample

// read fills the samples and answers the heap's system bytes.
func (s *realUsageSamples) read() uint64 {
	for i, name := range realUsageMetricNames {
		s[i].Name = name
	}
	gometrics.Read(s[:])
	var total uint64
	for i := range s {
		if s[i].Value.Kind() != gometrics.KindUint64 {
			return realUsageFallback()
		}
		total += s[i].Value.Uint64()
	}
	return total
}

// realUsageFallback answers HeapSys with ReadMemStats.
func realUsageFallback() uint64 {
	var stats goruntime.MemStats
	goruntime.ReadMemStats(&stats)
	return stats.HeapSys
}

// walkVisited answers the visited set a walk reuses between walks.
func (rt *Runtime) walkVisited() visitedSet {
	if rt.memVisit.arrays == nil {
		rt.memVisit = newVisitedSet()
	}
	rt.clearWalkVisited()
	return rt.memVisit
}

// clearWalkVisited drops a walk's keys, and the maps past memVisitKeep.
func (rt *Runtime) clearWalkVisited() {
	if len(rt.memVisit.arrays)+len(rt.memVisit.objects) > memVisitKeep {
		rt.memVisit = visitedSet{}
		return
	}
	clear(rt.memVisit.arrays)
	clear(rt.memVisit.objects)
}

// walkVMValue counts one live value a running flat VM reported.
func (rt *Runtime) walkVMValue(v any) {
	rt.memVMTotal += DeepSize(v, rt.memVisit)
}
