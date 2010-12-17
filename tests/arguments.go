package main

func sayforward(arg1, arg2 string) {
	print(arg1)
	print(arg2)
}

func saybackwards(arg1 string, arg2 string) {
	print(arg2)
	print(arg1)
}

func main() {
	sayforward("Hel", "lo ")
	saybackwards("!\n", "world")
}
