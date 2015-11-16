build: ubuntu_test.docker busybox_test.docker
ubuntu_test.docker: ubuntu_dockerfile
	docker build -t ubuntu_test -f ubuntu_dockerfile .
	docker save -o ubuntu_test.docker ubuntu_test
	docker rmi ubuntu_test
busybox_test.docker: busybox_dockerfile
	docker build -t busybox_test -f busybox_dockerfile .
	docker save -o busybox_test.docker busybox_test
	docker rmi busybox_test

