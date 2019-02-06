package controller

import (
	"github.com/vernemq/vmq-operator/pkg/controller/vernemq"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, vernemq.Add)
}
