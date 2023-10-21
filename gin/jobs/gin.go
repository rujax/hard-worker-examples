package jobs

import "fmt"

type GinJob struct {
	Message string
}

func (gj *GinJob) Work() {
	fmt.Printf("Message: %s\n", gj.Message)
}
