build:
	go build -o main

run: build
	./main

clean:
	rm *.png
	rm *.ppm
	rm main
