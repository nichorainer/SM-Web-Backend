package utils

import "fmt"

func NewProductID(seq int) string {
  return fmt.Sprintf("P-%04d", seq)
}