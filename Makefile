now 		  := $(shell date "+%Y-%m-%dT%H:%M:%S.%3N%z")


auto_commit:
	git add .
	git commit -am "$(now)"
	git pull
	git push