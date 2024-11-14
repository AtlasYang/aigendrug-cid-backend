package app

import "fmt"

func WeightKey(jobID int) string {
	return fmt.Sprintf("weight-%d.pth", jobID)
}
