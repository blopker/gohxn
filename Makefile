run:
	@go run gohxn.go

deploy:
	@git push dokku master

git:
	@git remote add dokku dokku@ssh.kbl.io:gohxn
