package orm

import (
	"fmt"
	"github.com/misakacoder/kagome/cond"
	"github.com/misakacoder/kagome/str"
	"github.com/misakacoder/kagome/times"
	"time"
)

type Partition interface {
	Type() string
	Strategy() string
	Parts() []Partition
}

func BuildRangePartitionType(column string) string {
	return fmt.Sprintf("range columns(`%s`)", column)
}

func BuildListPartitionType(column string) string {
	return fmt.Sprintf("list columns(`%s`)", column)
}

func BuildMonthPartitionStrategy(tm time.Time) string {
	parts := 14
	joiner := str.NewJoiner(", ", "( ", " )")
	tm = times.YearBeginning(tm)
	for i := 0; i < parts; i++ {
		partName := cond.Ternary(i == 0, "before", tm.AddDate(0, -1, 0).Format("2006-01"))
		if i == parts-1 {
			joiner.Append("partition `after` values less than maxvalue")
		} else {
			joiner.Append(fmt.Sprintf("partition `%s` values less than ('%s')", partName, tm.Format(time.DateTime)))
		}
		tm = tm.AddDate(0, 1, 0)
	}
	return joiner.String()
}

func BuildListPartitionStrategy(parts [][]any) string {
	joiner := str.NewJoiner(", ", "( ", " )")
	for i, part := range parts {
		values := str.NewJoiner(", ", "(", ")")
		for _, value := range part {
			if v, ok := value.(string); ok {
				values.Append(fmt.Sprintf("'%s'", v))
			} else {
				values.Append(fmt.Sprintf("%s", v))
			}
		}
		joiner.Append(fmt.Sprintf("PARTITION `p%d` VALUES IN %s", i, values.String()))
	}
	return joiner.String()
}
