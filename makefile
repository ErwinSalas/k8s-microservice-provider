deploy-zip:
	go build -o terraform-provider-k8s-microservice && zip terraform-provider-k8s-microservice.zip terraform-provider-k8s-microservice && mv terraform-provider-k8s-microservice.zip ~/providers

deploy-bin:
	go build -o terraform-provider-k8s-microservice && mv terraform-provider-k8s-microservice ~/providers