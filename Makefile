build:
	go build -o main

run:
	./main script.mdl

clean:
	rm -rf frames/
	rm -f *.gif
	rm -f *.png
	rm -f *.ppm
	rm -f main
