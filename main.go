package main

func main() {
	parser := NewParser()
	err := parser.ParseFile("script")
	if err != nil {
		panic(err)
	}
}
