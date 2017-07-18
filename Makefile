run:
	@go run hxn.go

deploy:
	@git push dokku master

git:
	@git remote add dokku dokku@ssh.kbl.io:gohxn
