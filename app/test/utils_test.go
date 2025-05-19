package test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
)

func TestUUID(t *testing.T) {
	uuid := uuid.New().String()
	fmt.Println(len(uuid))
}
