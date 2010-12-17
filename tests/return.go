package main

func sayhi() string {
	return "Hello world!"
}

func echo(x string) string {
	return x
}

func main() {
	println(echo(sayhi()))
}
