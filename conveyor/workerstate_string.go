// Code generated by "stringer -type=WorkerState"; DO NOT EDIT.

package conveyor

import "strconv"

const _WorkerState_name = "UnknownReadInputQueueReadResponseQueueProcessElements"

var _WorkerState_index = [...]uint8{0, 7, 21, 38, 53}

func (i WorkerState) String() string {
	if i < 0 || i >= WorkerState(len(_WorkerState_index)-1) {
		return "WorkerState(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _WorkerState_name[_WorkerState_index[i]:_WorkerState_index[i+1]]
}
