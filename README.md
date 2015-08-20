# Juno Server

## Directory tree
	bin - keeps compiled binary 
	src - keeps source code of the project
	vendor - keeps dependencies
	
## Install from source

GO version 1.4 or higher is required
project uses gb (http://getgb.io/) to handle dependencies,
to install it:

	go get github.com/constabulary/gb/...

now to build project type:

	gb build

## How to Run
server require two environment variable JUNO_PORT and JUNO_MONGO_URL,
I have created test mongo database on mongolab.com, so to start server type:

	JUNO_PORT=8888 JUNO_MONGO_URL=juser:jpass@ds031613.mongolab.com:31613/junodb bin/juno
	
there is acceptance test in file src/juno/juno_test.go
It dumps request/response and can provide an idea what API looks like:

	JUNO_PORT=8888 gb test juno
	