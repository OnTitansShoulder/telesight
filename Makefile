default: install

install:
	bash ./install.sh

clean:
	rm -rf telesight

start:
	bash ./start.sh

stop:
	bash ./stop.sh
