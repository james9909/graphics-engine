build:
	go build -o main

run:
	./main script

clean:
	rm -f *.png
	rm -f *.ppm
	rm -f main
