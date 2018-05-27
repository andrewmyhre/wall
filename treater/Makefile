default:
	docker build -t andrewmyhre/treater-compiler -f compiler.Dockerfile .
	docker run -v ${PWD}:/go/src/treater andrewmyhre/treater-compiler go build
	docker build -t andrewmyhre/treater .
