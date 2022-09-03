/*
Example: POST http://localhost:8111/api/fct1/sf11  BODY:{"p2":255, "p3":"Hi there!", "p1":true}
*/

package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"shiLink/shilink"
)

func main() {
	var (
		p1 bool
		p2 int
		p3 string

		root = &cobra.Command{
			Use:   "shiLink",
			Short: "ShiLink Server Testing",
		}

		fct1 = &cobra.Command{
			Use:   "fct1",
			Short: "execute function 1",
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println("Executing function 1")
			},
		}

		fct2 = &cobra.Command{
			Use:   "fct2",
			Short: "execute function 2",
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println("Executing function 2")
			},
		}

		sfct11 = &cobra.Command{
			Use:   "sf11",
			Short: "subfunction 1.1",
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println("Executing sub function 2")
				fmt.Printf("P1: %v\nP2: %v\nP3: %v\n", p1, p2, p3)
			},
		}
	)

	sfct11.Flags().BoolVar(&p1, "p1", false, "test bool parameter")
	sfct11.Flags().IntVar(&p2, "p2", 0, "test int parameter")
	sfct11.Flags().StringVar(&p3, "p3", "", "test string parameter")

	fct1.AddCommand(sfct11)
	root.AddCommand(fct1, fct2)

	shilink.UseShiLink(root)

	root.Execute()
}
