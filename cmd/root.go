/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/hellofreshdevtests/radu2020-recipe-count-test-2020/recipe"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "radu2020-recipe-count-test-2020",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		inputFixtureFileName, _ := cmd.Flags().GetString("fixture")
		inputPostcode, _ := cmd.Flags().GetString("postcode")
		inputFrom, _ := cmd.Flags().GetInt("from")
		inputTo, _ := cmd.Flags().GetInt("to")

		stream := recipe.NewJSONStream()

		recipeHistogram := make(map[string]int)
		postcodeHistogram := make(map[string]int)
		countPerPostcodeAndTime := 0

		fmt.Fprintf(os.Stdin, "Processing entries...\n")

		go func() {
			for data := range stream.Watch() {
				if data.Error != nil {
					fmt.Fprintf(os.Stderr, "err: %v\n", data.Error)
				}
				delivery := data.Recipe.Delivery
				from, to := extractFromTo(delivery)
				if data.Recipe.Postcode == inputPostcode {
					if from >= inputFrom && to <= inputTo {
						countPerPostcodeAndTime++
					}
				}

				v, ok := recipeHistogram[data.Recipe.Recipe]
				if ok {
					recipeHistogram[data.Recipe.Recipe] = v + 1
				} else {
					recipeHistogram[data.Recipe.Recipe] = 1
				}

				v, ok = postcodeHistogram[data.Recipe.Postcode]
				if ok {
					postcodeHistogram[data.Recipe.Postcode] = v + 1
				} else {
					postcodeHistogram[data.Recipe.Postcode] = 1
				}

			}
		}()

		stream.Start(inputFixtureFileName)

		data := Response{
			UniqueRecipeCount: getUniqueRecipeNamesCount(recipeHistogram),
			CountPerRecipe:    getCountPerRecipe(recipeHistogram),
			BusiestPostcode:   getBusiestPostCode(postcodeHistogram),
			CountPerPostcodeAndTime: CountPerPostcodeAndTime{
				Postcode:      inputPostcode,
				From:          strconv.Itoa(inputFrom) + "AM",
				To:            strconv.Itoa(inputTo) + "PM",
				DeliveryCount: countPerPostcodeAndTime,
			},
			MatchByName:             matchByName(recipeHistogram, args),
		}

		b, err := json.MarshalIndent(data, "", "    ")
		if err != nil {
			//
		}
		fmt.Print(string(b))


	},
}

func extractFromTo(s string) (from, to int) {
	delivery := strings.ReplaceAll(s, "AM", "")
	delivery = strings.ReplaceAll(delivery, "PM", "")
	words := strings.Fields(delivery)
	from, _ = strconv.Atoi(words[1])
	to, _ = strconv.Atoi(words[3])
	return
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("fixture", "fixture_short.json", "Use custom fixture file")
	rootCmd.PersistentFlags().String("postcode", "10120", "Search postcode")
	rootCmd.PersistentFlags().Int("from", 5, "Search from")
	rootCmd.PersistentFlags().Int("to", 9, "Search to")

}

type RecipeCount struct {
	Recipe string `json:"recipe"`
	Count  int    `json:"count"`
}

type BusiestPostcode struct {
	Postcode      string `json:"postcode"`
	DeliveryCount int    `json:"delivery_count"`
}

type CountPerPostcodeAndTime struct {
	Postcode      string `json:"postcode"`
	From          string `json:"from"`
	To            string `json:"to"`
	DeliveryCount int    `json:"delivery_count"`
}

type Response struct {
	UniqueRecipeCount       int                     `json:"unique_recipe_count"`
	CountPerRecipe          []RecipeCount           `json:"count_per_recipe"`
	BusiestPostcode         BusiestPostcode         `json:"busiest_postcode"`
	CountPerPostcodeAndTime CountPerPostcodeAndTime `json:"count_per_postcode_and_time"`
	MatchByName             []string                `json:"match_by_name"`
}

// getUniqueRecipeNamesCount takes a histogram as input and returns the number
// of unique recipes (so the ones with occurrence 1)
func getUniqueRecipeNamesCount(m map[string]int) int {
	counter := 0
	for _, v := range m {
		if v == 1 {
			counter++
		}
	}
	return counter
}

type kv struct {
	Key   string
	Value int
}

// sortMapValueDescending
// m - histogram (eg. map[key]occurrence)
// x - first x elements to return, if x is 0 then return all elements.
// Returns a slice of elements sorted by occurrence descending
// eg. If x is 2: [{Spinach Artichoke Pasta Bake 6} {Speedy Steak Fajitas 5}]
func sortMapValueDescending(m map[string]int, x int) []kv {
	var ss []kv
	for k, v := range m {
		ss = append(ss, kv{k, v})
	}

	// Then sorting the slice by value, higher first.
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	// Return the x top values or all values
	if x == 0 {
		return ss
	}
	return ss[:x]
}

// sortMapKeysAscending
// m - histogram (eg. map[key]occurrence)
// x - first x elements to return, if x is 0 then return all elements.
// Returns a slice of elements sorted by keys in ascending order
// eg. If x is 2: [{Recipe:Cajun-Spiced Pulled Pork Count:4} {Recipe:Cheesy Chicken Enchilada Bake Count:1}]
func sortMapKeysAscending(m map[string]int, x int) []kv {
	var ss []kv
	for k, v := range m {
		ss = append(ss, kv{k, v})
	}

	// Then sorting the slice by key, lower first.
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Key < ss[j].Key
	})

	// Return the x top values or all values
	if x == 0 {
		return ss
	}
	return ss[:x]
}

func getBusiestPostCode(m map[string]int) BusiestPostcode {
	// get postcode with the highest occurrence
	kv := sortMapValueDescending(m, 1)[0]
	return BusiestPostcode{
		Postcode:      kv.Key,
		DeliveryCount: kv.Value,
	}
}

// getCountPerRecipe takes input a histogram
// and returns a slice of RecipeCount sorted by Recipe name
func getCountPerRecipe(m map[string]int) []RecipeCount {
	kv := sortMapKeysAscending(m, 0)
	var ss []RecipeCount
	for _, v := range kv {
		ss = append(ss, RecipeCount{v.Key, v.Value})
	}
	return ss
}

func matchByName(m map[string]int, words []string) []string {
	var recipes []string
	if len(words) == 0 {
		words = []string{"Sweet", "Spanish"}
	}
	for k, _ := range m {
		for _, word := range words {
			if strings.Contains(k, word) {
				recipes = append(recipes, k)
			}
		}
	}
	return recipes
}