package cmd

import (
	"BloomFilter/pkg"
	"fmt"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bloom",
	Short: "Bloom фильтр CLI",
	Run: func(cmd *cobra.Command, args []string) {
		filter := pkg.NewBloomFilter(1000)
		filter.Add("mango")
		filter.Add("banana")
		fmt.Println("Exists mango: ", filter.Exists("mango"))
		fmt.Println("Exists apple: ", filter.Exists("apple"))
		fmt.Println("Exists banana: ", filter.Exists("banana"))
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
