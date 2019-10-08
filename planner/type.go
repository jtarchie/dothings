package planner

type planType int

const (
	Parallel planType = iota
	Serial
	Task
	Try
	Success
	Failure
	Finally
)

func (p planType) String() string {
	switch p {
	case Parallel:
		return "parallel"
	case Serial:
		return "serial"
	case Task:
		return "task"
	case Try:
		return "try"
	case Success:
		return "success"
	case Failure:
		return "failure"
	case Finally:
		return "finally"
	}
	return ""
}
