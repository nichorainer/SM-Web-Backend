package utils

import "fmt"

func NewProductID(seq int) string {
  return fmt.Sprintf("P-%04d", seq) // P-0001, P-0002...
}