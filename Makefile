build:
	go -C src vet
	go -C src build

clean:
	rm src/nest
