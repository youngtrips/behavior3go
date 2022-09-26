package composites

import (
	_ "fmt"
	"strconv"

	b3 "github.com/youngtrips/behavior3go"
	_ "github.com/youngtrips/behavior3go/config"
	. "github.com/youngtrips/behavior3go/core"
)

type Parallel struct {
	Composite
}

/**
 * Tick method.
 * @method tick
 * @param {b3.Tick} tick A tick instance.
 * @return {Constant} A state constant.
**/
func (this *Parallel) OnTick(tick *Tick) b3.Status {
	//fmt.Println("tick Parallel :", this.GetTitle())
	count := this.GetChildCount()
	maxN := count
	if v, ok := this.GetProperty("MaxSuccessCount"); ok {
		if i, err := strconv.Atoi(v); err == nil {
			maxN = i
		}
	}
	successed := 0
	for i := 0; i < count; i++ {
		var status = this.GetChild(i).Execute(tick)
		if status == b3.SUCCESS {
			successed++
		}
	}
	if successed >= maxN {
		return b3.SUCCESS
	}
	return b3.FAILURE
}
