
build: runner backup

docker: clean build
	docker build -t flynn-backup .

runner: src/runner.go
	go build -o dist/runner src/runner.go

backup: src/backup/backup_cluster.go
	go build -o dist/backup_cluster src/backup/backup_cluster.go

clean:
	rm -rf dist
