// Command go-linters runs the centrally maintained Go lint baseline.
package main

import (
 "fmt"
 "os"

 "github.com/rdpetrusek/go-linters/lint"
)

func main() {
 if err := lint.Run(lint.DefaultChecks(), os.Args[1:], os.Stderr); err != nil {
  fmt.Fprintln(os.Stderr, err)
  os.Exit(1)
 }
}
